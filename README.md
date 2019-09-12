# warlock -- grpc client Connection pool
[![Build Status](https://travis-ci.com/ruiboma/warlock.svg?branch=master)](https://travis-ci.com/ruiboma/warlock)
[![GoDoc](https://godoc.org/github.com/ruiboma/warlock?status.svg)](https://godoc.org/github.com/ruiboma/warlock)
[![LICENSE](https://img.shields.io/badge/licence-Apache%202.0-brightgreen.svg?style=flat-square)](https://github.com/ruiboma/warlock/blob/master/LICENSE)
[![GoDoc](https://godoc.org/github.com/ruiboma/warlock/clientfactory?status.svg)](https://godoc.org/github.com/ruiboma/warlock/clientfactory)
[![Open Source Helpers](https://www.codetriage.com/ruiboma/warlock/badges/users.svg)](https://www.codetriage.com/ruiboma/warlock)

**This is the golang grpc client connection pool tool**

# Project Maturity
This project is very young, its function is relatively simple, the basic api will not have more changes.
They will be discussed on Github issues along with any bugs or enhancements

# Goals
Provide a few interfaces, as efficient, stable and flexible as possible. May join the load balancing related module in the future and increase the availability of the connection

# Doc
warlock:    https://godoc.org/github.com/ruiboma/warlock \
clientFactory:    https://godoc.org/github.com/ruiboma/warlock/clientfactory

# HOW TO USE
```go
package main

import (
	"context"
	"fmt"

	"github.com/ruiboma/warlock"
	"github.com/ruiboma/warlock/config"

	"google.golang.org/grpc"
	pb "github.com/ruiboma/examples/helloworld/helloworld"
)

func main() {
	cfg := config.NewConfig()
	cfg.MaxCap = 100
	cfg.ServerAdds = &[]string{"127.0.0.1:50051"}

	pool, err := warlock.NewWarlock(cfg, grpc.WithInsecure())



	conn, close, err := pool.Acquire()
	defer close()                // It is recommended to use this, please do not use conn.Close because this will lead to waste




	c := pb.NewYourClient(conn)
	r, err := c.YourRPCFunc(ctx,balabala..)
    ...


    used, free := pool.Getstat() // Can view usage and free quantities
    /*
    *Maximum number of connections, But this number may be exceeded during the run, use configuration(OverflowCap = false) to avoid overflow,
    *if you need to strictly limit the number of connections
    */
```

# License
*BSD 2 clause - see LICENSE for more details.*