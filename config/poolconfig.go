package config

import (
	"math/rand"
	"sync"
	"time"
)

// Configuring the properties of the pool
type Config struct {
	lock           sync.RWMutex
	ServerAdds     *[]string
	MaxCap         int
	DynamicLink    bool
	OverflowCap    bool
	AcquireTimeout time.Duration
}

func init() {
	rand.Seed(time.Now().Unix())
}

// If there are multiple services  will be randomly selected
func (c *Config) GetTarget() string {
	c.lock.RLock()
	defer c.lock.RUnlock()
	cLen := len(*c.ServerAdds)
	if cLen <= 0 {
		return ""
	}
	return (*c.ServerAdds)[rand.Int()%cLen]

}
