// grpc client Connection pool
package warlock

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/ruiboma/warlock/config"

	"github.com/ruiboma/warlock/clientfactory"

	"google.golang.org/grpc"
)

var (
	AcquireError = errors.New("Exceeding the big wait time, you can fix this error by setting the overflow cap or increasing the maximum capacity of the cap.")
)

type closeFunc func()

type Pool struct {
	config      *config.Config
	mlock       sync.Mutex
	conns       chan *grpc.ClientConn
	factory     *clientfactory.PoolFactory
	usageAmount int
	ops         []grpc.DialOption
}

func NewWarlock(c *config.Config, ops ...grpc.DialOption) (*Pool, error) {
	conns := make(chan *grpc.ClientConn, c.MaxCap)
	factory := clientfactory.NewPoolFactory(c)
	pool := &Pool{config: c, conns: conns, factory: factory, ops: ops}
	err := factory.InitConn(conns, ops...)
	if err != nil {
		return nil, err
	} else {
		return pool, nil
	}

}

func (w *Pool) usagelock(add int) {
	w.mlock.Lock()
	defer w.mlock.Unlock()
	w.usageAmount += add
}

//  Fishing a usable link from the pool
func (w *Pool) Acquire() (*grpc.ClientConn, closeFunc, error) {
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

// Recycling available links
func (w *Pool) Close(client *grpc.ClientConn) {
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
func (w *Pool) Getstat() (used, surplus int) {
	return w.usageAmount, len(w.conns)
}

// If you want to end the program, don't forget it
func (w *Pool) ClearPool() {
	w.mlock.Lock()
	defer w.mlock.Unlock()
	for client := range w.conns {
		w.factory.Destroy(client)
	}

}
