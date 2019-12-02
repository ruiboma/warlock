# warlock -- grpc client Connection pool
[![Build Status](https://travis-ci.com/ruiboma/warlock.svg?branch=master)](https://travis-ci.com/ruiboma/warlock)
[![LICENSE](https://img.shields.io/badge/licence-Apache%202.0-brightgreen.svg?style=flat-square)](https://github.com/ruiboma/warlock/blob/master/LICENSE)
[![GoDoc](https://godoc.org/github.com/ruiboma/warlock?status.svg)](https://godoc.org/github.com/ruiboma/warlock)
[![Open Source Helpers](https://www.codetriage.com/ruiboma/warlock/badges/users.svg)](https://www.codetriage.com/ruiboma/warlock)
[![Go Report Card](https://goreportcard.com/badge/github.com/ruiboma/warlock)](https://goreportcard.com/report/github.com/ruiboma/warlock)


**This is the golang grpc client connection pool tool**\
**Complete link state detection mechanism Every link obtained is efficient.**

# Project Maturity
Basic api will not have more changes.
They will be discussed on [Github issues](https://github.com/ruiboma/warlock/issues) along with any bugs or enhancements

# Goals
Provide a few interfaces, as efficient, stable and flexible as possible.

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
	"github.com/ruiboma/warlock"
	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

func main() {

	cfg := warlock.NewConfig()
	cfg.ServerAdds = &[]string{"127.0.0.1:50051"}
    
        //  or
        cfg := warlock.NewConfig(warlock.OptionNoOverFlow, warlock.WithServerAdd(&[]string{"127.0.0.1:50051"}))
	



        pool, err := warlock.NewWarlock(cfg, grpc.WithInsecure())
	conn, close, err := pool.Acquire()
	defer close()

	
    /*
        Connection pool is not necessary for grpc
	used, free := pool.Getstat() // Can view usage and free quantities
	cfg.OverflowCap = true  This configuration may cause the existing link to exceed the total number set.
	If it overflows for a long time you need to consider increasing the value of cap.
	defer pool.ClearPool()  // Close all existing links with the pool before exiting the program
	defer close()  // It is recommended to use this, or <pool.Close(conn)> func 
    */
```

# License
*see [LICENSE](https://github.com/ruiboma/warlock/blob/master/LICENSE) for more details.*
