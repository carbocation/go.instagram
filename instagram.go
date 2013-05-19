package instagram

import (
	"errors"
	"log"
	"net/http"
	"sync"

	"code.google.com/p/goauth2/oauth"
)

//API docs: http://instagram.com/developer

const (
	AuthorizationURL = "https://api.instagram.com/oauth/authorize"
	AccessURL        = "https://api.instagram.com/oauth/access_token"
)

type Jar struct {
	cookies []*http.Cookie
}

type Instagram struct {
	AccessToken string
	client      *http.Client
	jar         *Jar
	Code        string
	OauthConfig *oauth.Config
	mu          *sync.Mutex
}

func check_error(err error) {
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
}

func NewInstagram() (*Instagram, error) {
	if !Config.Initialized() {
		return &Instagram{}, errors.New("Please initialize this library by calling the Initialize(...) function")
	}

	ig := &Instagram{
		client: &http.Client{},
		OauthConfig: &oauth.Config{
			ClientId:     Config.ClientID,
			ClientSecret: Config.ClientSecret,
			RedirectURL:  Config.RedirectURL,
			AuthURL:      AuthorizationURL,
			TokenURL:     AccessURL,
		},
	}

	return ig, nil
}

func (i Instagram) Authenticate(scope string) string {
	//first step: Direct user to Instagram authorization URL
	i.OauthConfig.AuthURL = AuthorizationURL
	i.OauthConfig.TokenURL = AccessURL
	i.OauthConfig.Scope = scope
	return i.OauthConfig.AuthCodeURL("")
}

func (i Instagram) GetAccessToken(code string) {
	//Third step of oauth2.0: Do a Post
	t := &oauth.Transport{Config: i.OauthConfig}
	t.Exchange(code)
	log.Println(t)
	/*check_error(err)
	resp, err := i.client.PostForm(u.String(), url.Values{"client_id": {i.ClientId}, "client_secret": {i.ClientSecret},
		"redirect_uri": {i.RedirectURI}, "code": {code}, "grant_type": {"authorization_code"}})
	check_error(err)
	log.Println("Get Access Token")
	log.Println(resp)*/
}
