package instagram

type Cfg struct {
	ClientID, ClientSecret, RedirectURL string
	initialized                         bool
}

//If you haven't called the Initialize function, this will return false
func (c *Cfg) Initialized() bool {
	return c.initialized
}

var Config *Cfg

//This must be called
func Initialize(c *Cfg) {
	Config = c
	Config.initialized = true
	//Config.ClientID, Config.ClientSecret, Config.RedirectURL, Config.initialized = ClientID, ClientSecret, RedirectURL, true
}
