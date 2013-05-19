package instagram

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
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

type InstagramData struct {
	Meta struct {
		Code int
	}
	Data []struct {
		Attribution string
		Tags        []string
		Location    struct {
			Latitude  float64
			Longitude float64
		}
		Comments struct {
			Count int64
			Data  []struct {
				ID          string
				CreatedTime string `json:"created_time"` //Unixtime
				Text        string
				From        struct {
					Username       string
					ProfilePicture string `json:"profile_picture"` //Note: this is really a URL
					ID             string
					FullName       string `json:"full_name"`
				}
			}
		}
		Filter      string
		CreatedTime string `json:"created_time"` //Unixtime
		Link        string //Note: this is really a URL
		Likes       struct {
			Count int64
			Data  []struct {
				Username       string
				ProfilePicture string `json:"profile_picture"` //Note: this is really a URL
				ID             string
				FullName       string `json:"full_name"`
			}
		}
		Images struct {
			LowResolution struct {
				URL    string
				Width  int64
				Height int64
			} `json:"low_resolution"` //Note: really a URL
			//LowResolution string `json:"low_resolution"` //Note: really a URL
			//LowResolution string `json:"low_resolution"` //Note: really a URL
		}
	}
	Pagination struct {
		MaxTagID string `json:"max_tag_id"`
		MinTagID string `json:"min_tag_id"`
		NextURL  string `json:"next_url"`
	}
}

func (ig *Instagram) TagsMediaRecent(tagName string) (string, error) {
	url := "https://api.instagram.com/v1/tags/" + tagName + "/media/recent"

	return ig.getDecode(url)
}

//Make the request with the appropriate authorization and decode the response
// into json
func (ig *Instagram) getDecode(url string) (string, error) {
	resp, err := http.Get(ig.AppendRequestType(url))
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return string(body), err
	}

	var data InstagramData

	if err := json.Unmarshal(body, &data); err != nil {
		return string(body), err
	}

	log.Printf("%+v", data)

	return string(body), nil
}

//This method inspects the instagram instance and the list of the last 5000 requests (vaporware)
//to see if we should be wrapping the request with the app's client_id or with a user's info
//If the 5000th request is more than 1 hour old, we can use our own client_id, otherwise
//we need to show the user an error or tell them to login
func (ig *Instagram) AppendRequestType(url string) string {
	//TODO(james) toggle based on whether or not user has logged in
	return fmt.Sprintf("%s?client_id=%s", url, ig.OauthConfig.ClientId)
}

func check_error(err error) {
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
}

func NewInstagram() (*Instagram, error) {
	Config.Lock()
	defer Config.Unlock()

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

func (ig Instagram) Authenticate(scope string) string {
	//first step: Direct user to Instagram authorization URL
	ig.OauthConfig.AuthURL = AuthorizationURL
	ig.OauthConfig.TokenURL = AccessURL
	ig.OauthConfig.Scope = scope
	return ig.OauthConfig.AuthCodeURL("")
}

func (ig Instagram) GetAccessToken(code string) {
	//Third step of oauth2.0: Do a Post
	t := &oauth.Transport{Config: ig.OauthConfig}
	t.Exchange(code)
	log.Println(t)
	/*check_error(err)
	resp, err := i.client.PostForm(u.String(), url.Values{"client_id": {i.ClientId}, "client_secret": {i.ClientSecret},
		"redirect_uri": {i.RedirectURI}, "code": {code}, "grant_type": {"authorization_code"}})
	check_error(err)
	log.Println("Get Access Token")
	log.Println(resp)*/
}
