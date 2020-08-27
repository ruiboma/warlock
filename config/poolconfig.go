package config

import (
	"math/rand"
	"sync"
	"time"
)

// Config Configuring the properties of the pool
type Config struct {
	Lock           *sync.RWMutex
	ServerAdds     *[]string
	MaxCap         int64
	DynamicLink    bool
	OverflowCap    bool
	AcquireTimeout time.Duration
	GetTargetFunc  GetTargetFunc
}

type GetTargetFunc func(c *Config) string

func init() {
	rand.Seed(time.Now().Unix())
}

// GetTarget If there are multiple services  will be randomly selected
func (c *Config) GetTarget() string {
	c.Lock.RLock()
	defer c.Lock.RUnlock()
	if c.GetTargetFunc != nil {
		return c.GetTargetFunc(c)
	}
	cLen := len(*c.ServerAdds)
	if cLen <= 0 {
		return ""
	}
	return (*c.ServerAdds)[rand.Int()%cLen]
}
