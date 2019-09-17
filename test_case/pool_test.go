package test_case

import (
	"reflect"
	"testing"

	"github.com/ruiboma/warlock"
	"github.com/ruiboma/warlock/config"
	"google.golang.org/grpc"
)

func TestNewWarlock(t *testing.T) {
	cases := []struct {
		name      string
		c         *config.Config
		d         []grpc.DialOption
		wantRes   *warlock.Pool
		wantError bool
	}{
		// your test cases
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			res, err := warlock.NewWarlock(test.c, test.d...)
			if (err != nil) != test.wantError {
				t.Errorf("NewWarlock() error = %v, resError %v", err, test.wantError)
			}
			if !reflect.DeepEqual(test.wantRes, res) {
				t.Errorf("NewWarlock() res = %v wantres %v", res, test.wantRes)
			}
		})
	}

}

func TestAcquire(t *testing.T) {
	cases := []struct {
		name      string
		p         *warlock.Pool
		wantconn  *grpc.ClientConn
		wantclose *warlock.CloseFunc
		wanterror bool
	}{
		// your test cases
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			conn, close, err := test.p.Acquire()
			if !reflect.DeepEqual(conn, test.wantconn) {
				t.Errorf("(w *Pool)Acquire() res=%v want %v", conn, test.wantconn)
			}
			if !reflect.DeepEqual(close, test.wantclose) {
				t.Errorf("(w *Pool)Acquire() close=%v want %v", close, test.wantclose)
			}
			if (err != nil) != test.wanterror {
				t.Errorf("(w *Pool)Acquire() errors=%v want %v", (err != nil), test.wanterror)
			}
		})
	}

}
