package instagram

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
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

type InstagramUser struct {
	Bio            string
	FullName       string `json:"full_name"`
	ID             string
	ProfilePicture string `json:"profile_picture"` //Note: this is really a URL
	Username       string
	Website        string
}

type InstagramComment struct {
	CreatedTime string `json:"created_time"` //Unixtime
	From        InstagramUser
	ID          string
	Text        string
}

type InstagramImage struct {
	URL    string
	Width  int64
	Height int64
}

type InstagramData struct {
	Meta struct {
		Code int
		URL  string
	}
	Data []struct {
		Attribution string
		Caption     InstagramComment
		Comments    struct {
			Count int64
			Data  []InstagramComment
		}
		CreatedTime string `json:"created_time"` //Unixtime
		Filter      string
		ID          string
		Images      struct {
			LowResolution      InstagramImage `json:"low_resolution"`      //Note: really a URL
			StandardResolution InstagramImage `json:"standard_resolution"` //Note: really a URL
			Thumbnail          InstagramImage //Note: really a URL
		}
		Likes struct {
			Count int64
			Data  []InstagramUser
		}
		Link     string //Note: this is really a URL
		Location struct {
			ID        int64
			Latitude  float64
			Longitude float64
			Name      string
		}
		Tags         []string
		Type         string
		User         InstagramUser
		UsersInPhoto []InstagramUser
	}
	Pagination struct {
		MaxTagID string `json:"max_tag_id"`
		MinTagID string `json:"min_tag_id"`
		NextURL  string `json:"next_url"`
	}
}

func (ig *Instagram) TagsMediaRecent(tagName []string) (*InstagramData, error) {
	return ig.tagsMediaRecent(tagName, "")
}

func (ig *Instagram) tagsMediaRecent(tagName []string, maxTagID string) (*InstagramData, error) {
	//TODO(james) parse more than just the first tag
	u, err := url.Parse("https://api.instagram.com/v1/tags/" + tagName[0] + "/media/recent")
	if err != nil {
		return &InstagramData{}, err
	}

	//Construct our query string. If we've been given a maxTagID, add it
	qs := u.Query()
	if maxTagID != "" {
		qs.Set("max_tag_id", maxTagID)
	}

	//return &InstagramData{}, nil
	return ig.getDecode(u, &qs)
}

//Make the request with the appropriate authorization and decode the response
// into json
func (ig *Instagram) getDecode(u *url.URL, qs *url.Values) (*InstagramData, error) {
	var data InstagramData
	
	location := ig.AppendRequestType(u, qs)
	data.Meta.URL = location
	
	resp, err := http.Get(location)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &data, err
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return &data, err
	}

	log.Printf("%+v", data)

	return &data, nil
}

//This method inspects the instagram instance and the list of the last 5000 requests (vaporware)
//to see if we should be wrapping the request with the app's client_id or with a user's info
//If the 5000th request is more than 1 hour old, we can use our own client_id, otherwise
//we need to show the user an error or tell them to login
func (ig *Instagram) AppendRequestType(u *url.URL, qs *url.Values) string {
	//Combine the query strings
	uQueryString := u.Query()
	for k, v := range uQueryString {
		for _, realVal := range v {
			qs.Add(k, realVal)
		}
	}
	//TODO(james) toggle based on whether or not user has logged in
	qs.Set("client_id", ig.OauthConfig.ClientId)

	return u.Scheme + `://` + u.Host + u.Path + `?` + qs.Encode()
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
