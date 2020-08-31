// Package grpcpool  grpc client Connection pool  warlock
package warlock

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ruiboma/warlock/config"

	"github.com/ruiboma/warlock/clientfactory"

	"google.golang.org/grpc"
)

var (
	errAcquire = errors.New("acquire connection timed out, you can fix this error by setting the overflow cap or increasing the maximum capacity of the cap")
	errTimeout = errors.New("warlock: Connection timed out, check the address configuration or network status")
)

//CloseFunc should defer
type CloseFunc func()
type chanStat int

const (
	isClose chanStat = iota
	isOpen
)

type WarOption func(*config.Config)

// Pool connection pool
type Pool struct {
	Config      *config.Config
	mLock       *sync.Mutex
	conns       chan *grpc.ClientConn
	factory     *clientfactory.PoolFactory
	usageAmount int64
	ops         []grpc.DialOption
	ChannelStat chanStat
}

func WithMaxCap(num int64) WarOption {
	return func(i *config.Config) {
		i.MaxCap = num
	}
}

func WithServerAdd(s *[]string) WarOption {
	return func(i *config.Config) {
		i.ServerAdds = s
	}
}

func WithAcquireTimeOut(num time.Duration) WarOption {
	return func(i *config.Config) {
		i.AcquireTimeout = num
	}
}

// Custom get address
func WithGetTargetFunc(g config.GetTargetFunc) WarOption {
	return func(c *config.Config) {
		c.GetTargetFunc = g
	}
}

func OptionNoOverFlow(i *config.Config) {
	i.OverflowCap = false
}

func OptionDynamicLink(i *config.Config) {
	i.DynamicLink = true
}

// NewConfig Get a config object and then customize his properties
func NewConfig(ops ...WarOption) *config.Config {
	c := &config.Config{Lock: &sync.RWMutex{}}
	c.MaxCap = 10
	c.DynamicLink = false
	c.OverflowCap = true
	c.AcquireTimeout = 3 * time.Second
	for _, f := range ops {
		f(c)
	}
	return c
}

// NewWarlock  Get a warlovk (connection pool)
func NewWarlock(c *config.Config, ops ...grpc.DialOption) (*Pool, error) {
	conns := make(chan *grpc.ClientConn, c.MaxCap)
	factory := clientfactory.NewPoolFactory(c)
	pool := &Pool{Config: c, conns: conns, factory: factory, ops: ops, ChannelStat: 1, usageAmount: 0, mLock: &sync.Mutex{}}
	err := factory.InitConn(conns, ops...)
	if err != nil {
		return nil, err
	}
	return pool, nil
}

func (w *Pool) usagelock(add int64) {
	atomic.AddInt64(&w.usageAmount, add)
}

// Acquire  Fishing a usable link from the pool
func (w *Pool) Acquire() (*grpc.ClientConn, CloseFunc, error) {
	ctx, cancel := context.WithTimeout(context.Background(), w.Config.AcquireTimeout*time.Second)
	defer cancel()
	for {
		select {
		case clientConn := <-w.conns:
			con := w.factory.Activate(clientConn)
			switch con {
			case 0:
				w.usagelock(1)
				return clientConn, func() { w.Close(clientConn) }, nil
			case 1:
				w.Close(clientConn)
				continue
			default:
				w.usagelock(-1)
				w.factory.Destroy(clientConn)
				continue
			}
		case <-ctx.Done():
			return nil, nil, errAcquire
		default:
			if w.Config.OverflowCap == false && w.usageAmount >= w.Config.MaxCap {
				continue
			} else {
				Wops := append(w.ops, grpc.WithBlock())
				clientconn, err := w.factory.MakeConn(w.Config.GetTarget(), Wops...)
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

// GetStat Return to the use of resources in the pool
func (w *Pool) GetStat() (used int64, surplus int) {
	return atomic.LoadInt64(&w.usageAmount), len(w.conns)
}

// ClearPool Disconnect the link at the end of the program
func (w *Pool) ClearPool() {
	w.mLock.Lock()
	defer w.mLock.Unlock()
	w.ChannelStat = isClose
	close(w.conns)
	for client := range w.conns {
		w.factory.Destroy(client)
	}

}
