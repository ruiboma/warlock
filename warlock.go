package warlock

import (
	"context"
	"errors"
	"sync"
	"time"
	"warlock/clientfactory"
	"warlock/config"

	"google.golang.org/grpc"
)

var (
	AcquireError = errors.New("Exceeding the big wait time, you can fix this error by setting the overflow cap or increasing the maximum capacity of the cap.")
)

// todo
// When the maximum capacity is reached, it will enter the black hole of frequent establishment and destruction of links,
// which will consume a lot of resources.
type Health struct {
}
type closeFunc func()

type pool struct {
	config      *config.Config
	mlock       sync.Mutex
	conns       chan *grpc.ClientConn
	factory     *clientfactory.PoolFactory
	usageAmount int
	ops         []grpc.DialOption
}

func NewWarlock(c *config.Config, ops ...grpc.DialOption) (*pool, error) {
	conns := make(chan *grpc.ClientConn, c.MaxCap)
	factory := clientfactory.NewPoolFactory(c)
	pool := &pool{config: c, conns: conns, factory: factory, ops: ops}
	err := factory.InitConn(conns, ops...)
	if err != nil {
		return nil, err
	} else {
		return pool, nil
	}

}

func (w *pool) usagelock(add int) {
	w.mlock.Lock()
	defer w.mlock.Unlock()
	w.usageAmount += add
}

func (w *pool) Acquire() (*grpc.ClientConn, closeFunc, error) {
	ctx, cancel := context.WithTimeout(context.Background(), w.config.AcquireTimeout*time.Second)
	defer cancel()
	for {
		select {
		case clientconn := <-w.conns:
			con := w.factory.Activate(clientconn)
			switch con {
			case 0:
				w.usagelock(1)
				return clientconn, func() { w.Close(clientconn) }, nil
			case 1:
				w.Close(clientconn)
				continue
			default:
				w.usagelock(-1)
				w.factory.Destroy(clientconn)
				continue
			}
		case <-ctx.Done():
			return nil, nil, AcquireError
		default:
			if w.config.OverflowCap == false && w.usageAmount >= w.config.MaxCap {
				continue
			} else {
				clientconn, err := w.factory.MakeConn(w.config.GetTarget(), w.ops...)
				if err != nil {
					return nil, nil, err
				}
				w.usagelock(1)
				return clientconn, func() { w.Close(clientconn) }, nil
			}
		}
	}

}

func (w *pool) Close(client *grpc.ClientConn) {
	go func() {
		detect, _ := w.factory.Passivate(client)
		if detect == true {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			select {
			case w.conns <- client:
			case <-ctx.Done():
				w.factory.Destroy(client)
			}
		}
		w.usagelock(-1)

	}()

}

// Return to the use of resources in the pool
func (w *pool) Getstat() (used, surplus int) {
	return w.usageAmount, len(w.conns)
}
