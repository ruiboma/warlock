package testcase

import (
	"reflect"
	"testing"

	"github.com/ruiboma/warlock"
	"github.com/ruiboma/warlock/config"
	"google.golang.org/grpc"
)

var (
	// CFG init config
	CFG *config.Config
	// TPool init Pool
	TPool *warlock.Pool
)

func init() {
	CFG = warlock.NewConfig()
	CFG.ServerAdds = &[]string{"127.0.0.1:50051"}
	CFG.MaxCap = 100
	CFG.OverflowCap = false // true
	tp, err := warlock.NewWarlock(CFG, grpc.WithInsecure())
	TPool = tp
	if err != nil {
		panic(err)
	}
}

// TestNewWarlock test get pool
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

// TestAcquire test acq func
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

// TestClose test recycling
func TestClose(t *testing.T) {
	cfg := warlock.NewConfig()
	cfg.ServerAdds = &[]string{"127.0.0.1:50051"}
	tp, err := warlock.NewWarlock(cfg, grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	tConn, close, err := tp.Acquire()
	defer close()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(tConn.GetState())
	t.Log(tp.Getstat())
	cases := []struct {
		name string
		p    *warlock.Pool
		conn *grpc.ClientConn
	}{
		{name: "t01", p: tp, conn: tConn},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			t.Log(test.conn.GetState())
			if test.conn.GetState() != 2 {
				t.Errorf("(p *Pool)Close() stat %v , wantstat = %s", test.conn.GetState(), "ready")
			}
		})
	}

}

// BenchmarkPara  Benchmark test
func BenchmarkPara(b *testing.B) {
	cases := []struct {
		name string
	}{
		{"01"},
	}
	for _, test := range cases {
		b.Log(test)
		b.RunParallel(func(b *testing.PB) {
			for b.Next() {
				_, close, err := TPool.Acquire()
				if err != nil {
					panic(err)
				}
				close()
			}
		})

	}
	b.Log(TPool.Getstat())
}
