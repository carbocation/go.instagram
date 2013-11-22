package instagram

import (
	"sync"
)

type Cfg struct {
	ClientID, ClientSecret, RedirectURL string
	initialized                         bool
	sync.Mutex
}

//If you haven't called the Initialize function, this will return false
func (c *Cfg) Initialized() bool {
	return c.initialized
}

var Config *Cfg

//This must be called
func Initialize(c *Cfg) {
	c.Lock()
	defer c.Unlock()

	Config = c
	Config.initialized = true
}
