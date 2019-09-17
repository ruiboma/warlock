# warlock -- grpc client Connection pool
[![Build Status](https://travis-ci.com/ruiboma/warlock.svg?branch=master)](https://travis-ci.com/ruiboma/warlock)
[![LICENSE](https://img.shields.io/badge/licence-Apache%202.0-brightgreen.svg?style=flat-square)](https://github.com/ruiboma/warlock/blob/master/LICENSE)
[![GoDoc](https://godoc.org/github.com/ruiboma/warlock?status.svg)](https://godoc.org/github.com/ruiboma/warlock)
<!-- [![GoDoc](https://godoc.org/github.com/ruiboma/warlock/clientfactory?status.svg)](https://godoc.org/github.com/ruiboma/warlock/clientfactory) -->
[![Open Source Helpers](https://www.codetriage.com/ruiboma/warlock/badges/users.svg)](https://www.codetriage.com/ruiboma/warlock)

**This is the golang grpc client connection pool tool**

# Project Maturity
This project function is relatively simple, the basic api will not have more changes.
They will be discussed on [Github issues](https://github.com/ruiboma/warlock/issues) along with any bugs or enhancements

# Goals
Provide a few interfaces, as efficient, stable and flexible as possible. May join the load balancing related module in the future and increase the availability of the connection

# Doc
[warlock](https://godoc.org/github.com/ruiboma/warlock)\
[clientFactory](https://godoc.org/github.com/ruiboma/warlock/clientfactory)

# HOW TO USE
```shell
go get github.com/ruiboma/warlock
```
```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ruiboma/warlock"

	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

const (
	address = "localhost:50051"
)

func main() {

	cfg := warlock.NewConfig()
	cfg.MaxCap = 100
	cfg.OverflowCap = true
	// This configuration may cause the existing link to exceed the total number set.
	// If it overflows for a long time, you need to consider increasing the value of cap.
	cfg.ServerAdds = &[]string{"127.0.0.1:50051"}
	pool, err := warlock.NewWarlock(cfg, grpc.WithInsecure())
	defer pool.ClearPool()  // Close all existing links with the pool before exiting the program

	if err != nil {
		panic(err)
	}
	conn, close, err := pool.Acquire()
	defer close()  // It is recommended to use this, or use  <pool.Close(conn)> func

	if err != nil {
		panic(err)
	}
	



	c := pb.NewYourClient(conn)S
	r, err := c.YourRPCFunc(ctx,balabala..)
    ...



    /*
    *	used, free := pool.Getstat() // Can view usage and free quantities
	*	
    *
    */
```

# License
*BSD 2 clause - see [LICENSE](https://github.com/ruiboma/warlock/blob/master/LICENSE) for more details.*