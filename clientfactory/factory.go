package clientfactory

import (
	"context"
	"errors"
	"github.com/ruiboma/warlock/config"
	"google.golang.org/grpc"
	"strings"
	"time"
)

var (
	errorTarget = errors.New("Address is empty or invalid")
)

// PoolFactory object
type PoolFactory struct {
	config *config.Config
}
type condition = int

const (
	// Ready Can be used
	Ready condition = iota
	// Put Not available. Maybe later.
	Put
	// Destroy Failure occurs and cannot be restored
	Destroy
)

// NewPoolFactory get poolFactory
func NewPoolFactory(c *config.Config) *PoolFactory {
	return &PoolFactory{config: c}
}

// Passivate Action before releasing the resource
func (f *PoolFactory) Passivate(conn *grpc.ClientConn) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if conn.WaitForStateChange(ctx, 3) && conn.WaitForStateChange(ctx, 4) && conn.WaitForStateChange(ctx, 0) {
		return true, nil
	}
	return false, f.Destroy(conn)

}

// Activate Action taken after getting the resource
func (f *PoolFactory) Activate(conn *grpc.ClientConn) int {
	stat := conn.GetState()
	switch {
	case stat == 2:
		return Ready
	case stat == 0 || stat == 1 || stat == 3:
		return Put
	default:
		return Destroy
	}

}

// Destroy tears down the ClientConn and all underlying connections.
func (f *PoolFactory) Destroy(conn *grpc.ClientConn) error {
	return conn.Close()
}

// MakeConn Users are not recommended to use this API
func (f *PoolFactory) MakeConn(target string, ops ...grpc.DialOption) (*grpc.ClientConn, error) {
	if target == "" || strings.Index(target, ":") == -1 {
		return nil, errorTarget
	}
	if f.config.DynamicLink == true {
		return grpc.Dial(target, ops...)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return grpc.DialContext(ctx, target, ops...)

}

// InitConn Initialize the create link
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
