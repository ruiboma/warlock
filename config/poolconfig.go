package config

import (
	"math/rand"
	"sync"
	"time"
)

type Config struct {
	lock           sync.RWMutex
	ServerAdds     *[]string
	MaxCap         int
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	DynamicLink    bool
	OverflowCap    bool
	AcquireTimeout time.Duration
}

func init() {
	rand.Seed(time.Now().Unix())
}

func NewConfig() *Config {
	c := &Config{}
	c.MaxCap = 10
	c.DynamicLink = true
	c.OverflowCap = true
	c.AcquireTimeout = 3 * time.Second
	return c
}

func (c *Config) GetTarget() string {
	c.lock.RLock()
	defer c.lock.RUnlock()
	cLen := len(*c.ServerAdds)
	if cLen <= 0 {
		return ""
	}
	return (*c.ServerAdds)[rand.Int()%cLen]

}
