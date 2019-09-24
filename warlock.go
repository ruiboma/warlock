// Package warlock grpc client Connection pool
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
	errAcquire = errors.New("Exceeding the big wait time, you can fix this error by setting the overflow cap or increasing the maximum capacity of the cap.")
	errTimeout = errors.New("warlock: Connection timed out, check the address configuration or network status.")
)

//CloseFunc should defer
type CloseFunc func()
type chanStat int

const (
	isClose chanStat = iota
	isOpen
)

// Pool connection pool
type Pool struct {
	config      *config.Config
	mlock       sync.Mutex
	conns       chan *grpc.ClientConn
	factory     *clientfactory.PoolFactory
	usageAmount int
	ops         []grpc.DialOption
	ChannelStat chanStat
}

// NewConfig Get a config object and then customize his properties
func NewConfig() *config.Config {
	c := &config.Config{}
	c.MaxCap = 10
	c.DynamicLink = false
	c.OverflowCap = true
	c.AcquireTimeout = 3 * time.Second
	return c
}

// NewWarlock  Get a warlovk (connection pool)
func NewWarlock(c *config.Config, ops ...grpc.DialOption) (*Pool, error) {
	conns := make(chan *grpc.ClientConn, c.MaxCap)
	factory := clientfactory.NewPoolFactory(c)
	pool := &Pool{config: c, conns: conns, factory: factory, ops: ops, ChannelStat: 1}
	err := factory.InitConn(conns, ops...)
	if err != nil {
		return nil, err
	}
	return pool, nil

}

func (w *Pool) usagelock(add int) {
	w.mlock.Lock()
	defer w.mlock.Unlock()
	w.usageAmount += add
}

// Acquire  Fishing a usable link from the pool
func (w *Pool) Acquire() (*grpc.ClientConn, CloseFunc, error) {
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
			return nil, nil, errAcquire
		default:
			if w.config.OverflowCap == false && w.usageAmount >= w.config.MaxCap {
				continue
			} else {
				Wops := append(w.ops, grpc.WithBlock())
				clientconn, err := w.factory.MakeConn(w.config.GetTarget(), Wops...)
				if err != nil {
					if err == context.DeadlineExceeded {
						return nil, nil, errTimeout
					}
					return nil, nil, err
				}
				w.usagelock(1)
				return clientconn, func() { w.Close(clientconn) }, nil
			}
		}
	}

}

// Close Recycling available links
func (w *Pool) Close(client *grpc.ClientConn) {
	go func() {
		detect, _ := w.factory.Passivate(client)
		if detect == true && w.ChannelStat == isOpen {
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

// Getstat Return to the use of resources in the pool
func (w *Pool) Getstat() (used, surplus int) {
	return w.usageAmount, len(w.conns)
}

// ClearPool Disconnect the link at the end of the program
func (w *Pool) ClearPool() {
	w.mlock.Lock()
	defer w.mlock.Unlock()
	w.ChannelStat = isClose
	close(w.conns)
	for client := range w.conns {
		w.factory.Destroy(client)
	}

}
