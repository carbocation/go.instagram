package instagram

import (
	"encoding/json"
	"errors"
	"fmt"
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

type InstagramResponse struct {
	Meta struct {
		Code int
		URL  string `json:"-"` //This is not part of Instagram's output
	}
	Data       []InstagramData
	Pagination struct {
		MaxTagID string `json:"max_tag_id"`
		MinTagID string `json:"min_tag_id"`
		NextURL  string `json:"next_url"`
	}
}

func (ig *Instagram) TagsMediaRecent(tags []string) (*[]InstagramData, error) {
	/*
		The plan:

		1. Fire off a goroutine for each tag
		2. Collect each response
		3. See if there are any results where all tags are there
		4. If so, send those over
	*/

	//There will be a channel that accepts slices of InstagramData
	var igdChan = make(chan []InstagramData)
	
	//There will be a channel that tells us something bad has happened
	var errChan = make(chan error)

	//For each tag, start a new goroutine that queries Instagram
	for _, tag := range tags {
		go func(t string) {
			res, err := ig.TagMediaRecent(t)
			if err != nil {
				errChan <- err
			}
			igdChan <- *res

			return
		}(tag)
	}

	var igd = []InstagramData{}

	for i := 0; i < len(tags); i++ {
		select {
		case result, ok := <-igdChan:
			if !ok {
				fmt.Println("This channel is closed")
			} else {
				igd = append(igd, result...)
			}
		case badness := <-errChan:
			return &igd, badness
		}
	}

	return &igd, nil


	/*
	
		//Serial version

		//For now, we run the query for all tags and send all of them back
		//Below is the procedural / in-order version
		var igd = []InstagramData{}

		for _, tag := range tags {
			res, err := ig.TagMediaRecent(tag)
			if err != nil {
				return &igd, err
			}

			igd = append(igd, *res...)
		}

		return &igd, nil
	*/
}

func (ig *Instagram) TagMediaRecent(tag string) (*[]InstagramData, error) {
	igr, err := ig.tagMediaRecent(tag, "")
	if err != nil {
		return &igr.Data, err
	}

	if igr.Meta.Code != http.StatusOK {
		return &igr.Data, errors.New(fmt.Sprintf("Instagram returned a %d error", igr.Meta.Code))
	}

	return &igr.Data, err
}

func (ig *Instagram) tagMediaRecent(tagName, maxTagID string) (*InstagramResponse, error) {
	//TODO(james) parse more than just the first tag
	u, err := url.Parse("https://api.instagram.com/v1/tags/" + tagName + "/media/recent")
	if err != nil {
		return &InstagramResponse{}, err
	}

	//Construct our query string. If we've been given a maxTagID, add it
	qs := u.Query()
	if maxTagID != "" {
		qs.Set("max_tag_id", maxTagID)
	}

	return ig.getDecode(u, &qs)
}

//Make the request with the appropriate authorization and decode the response into json
func (ig *Instagram) getDecode(u *url.URL, qs *url.Values) (*InstagramResponse, error) {
	var data InstagramResponse

	//Build the location from your URL data
	location := ig.BuildQuery(u, qs)

	//Store location in the struct so we can access our current URL if we feel like it
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

	return &data, nil
}

//BuildQuery inspects the instagram instance and the list of the last 5000 requests (vaporware)
//to see if we should be wrapping the request with the app's client_id or with a user's info
//If the 5000th request is more than 1 hour old, we can use our own client_id, otherwise
//we need to show the user an error or tell them to login
func (ig *Instagram) BuildQuery(u *url.URL, qs *url.Values) string {
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
