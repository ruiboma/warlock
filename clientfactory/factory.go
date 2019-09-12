package clientfactory

import (
	"context"
	"errors"
	"strings"
	"time"
	"warlock/config"

	"google.golang.org/grpc"
)

var (
	targetError = errors.New("Address is empty or invalid")
)

type PoolFactory struct {
	config *config.Config
}
type Condition int

const (
	// Can be used
	Ready Condition = iota
	// Not available. Maybe later.
	Put
	// Failure occurs and cannot be restored
	Destroy
)

func NewPoolFactory(c *config.Config) *PoolFactory {
	return &PoolFactory{config: c}
}

// Action before releasing the resource
func (f *PoolFactory) Passivate(conn *grpc.ClientConn) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if conn.WaitForStateChange(ctx, 3) && conn.WaitForStateChange(ctx, 4) {
		return true, nil
	} else {
		return false, f.Destroy(conn)
	}
}

// Action taken after getting the resource
func (f *PoolFactory) Activate(conn *grpc.ClientConn) Condition {
	stat := conn.GetState()
	switch {
	case stat%2 == 0:
		return Ready
	case stat%2 > 0:
		return Put
	default:
		return Destroy
	}

}

// Destroy tears down the ClientConn and all underlying connections.
func (f *PoolFactory) Destroy(conn *grpc.ClientConn) error {
	return conn.Close()
}

// Users are not recommended to use this API
func (f *PoolFactory) MakeConn(target string, ops ...grpc.DialOption) (*grpc.ClientConn, error) {
	if target == "" || strings.Index(target, ":") == -1 {
		return nil, targetError
	}
	if f.config.DynamicLink == true {
		return grpc.Dial(target, ops...)
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return grpc.DialContext(ctx, target, ops...)
	}

}

func (f *PoolFactory) InitConn(conns chan *grpc.ClientConn, ops ...grpc.DialOption) error {
	l := cap(conns) - len(conns)
	for i := 1; i <= l; i++ {
		cli, err := f.MakeConn(f.config.GetTarget(), ops...)
		if err != nil {
			return err
		}
		conns <- cli
	}
	return nil

}
