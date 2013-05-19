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

type Semaphore struct{}

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
	simultaneous int
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

/*
func (ig *Instagram) TagsMediaRecent(tags []string) (*[]InstagramData, error) {
	//The plan:
	//
	//1. Fire off a goroutine for each tag
	//2. Collect each response
	//3. See if there are any results where all tags are there
	//4. If so, send those over

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

	//We know that there are only
	for _, _ = range tags {
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

	//This will store the output that has all tags
	saved := make([]InstagramData, 0)
	addedPhotos := map[string]bool{}

NextItem:
	//For each photo returned, check all original tags
	for _, photo := range igd {
		//Check all tags that were in our query
		for _, tag := range tags {
			//Until proven otherwise, we don't think this tag was in the photo
			hasThisTag := false

			//Was that tag found in the photo?
			for _, appliedTag := range photo.Tags {
				if appliedTag == tag {
					//Yes! Now we assume it had every tag unless proven otherwise
					hasThisTag = true
					break
				}
				//On the last tag, we only get here if the last tag was not among the v.Tags
				hasThisTag = false
			}

			if !hasThisTag {
				continue NextItem
			}
		}

		//Congrats, nothing kicked you out. This post survives.
		if !addedPhotos[photo.ID] {
			//Unless you've already seen this exact photo
			saved = append(saved, photo)
			addedPhotos[photo.ID] = true
		}
	}

	return &saved, nil
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
*/

func (ig *Instagram) TagsMediaRecent(tags []string) (*[]InstagramData, error) {
	simultaneous := len(tags)
	
	//Buffered channel of results
	results := make(chan *InstagramResponse, simultaneous)
	blocker := make(chan []InstagramData)
	
	//How shall the consumer decide whether each datum should be kept?
	keepCriterion := func(igDatum InstagramData) bool{
		hasTags := false
		for _, tag := range tags {
			hasTags = false
			for _, possessedTag := range igDatum.Tags {
				if possessedTag == tag {
					hasTags = true
					break
				}
			}
			
			if !hasTags {
				//If we get here and hasTags is still false, it means this tag wasn't matched
				return false
			}
		}
		
		return true
	}

	for _, tag := range tags {
		//Generate the URL for each tag
		
		go ig.producer(results, ig.tagURL(tag))
	}
	
	go ig.consumer(results, blocker, simultaneous, keepCriterion)

	//Wait until the consumer is satisfied
	output := <-blocker
	
	//return &[]InstagramData{}, nil
	return &output, nil
}

//Take whatever URL we get as 'work' and send back an InstagramResult object
func (ig *Instagram) producer(results chan *InstagramResponse, work string) {	
	//Todo: Error handling
	r, _ := ig.getDecode(work)
	
	results <- r
	return
}

func (ig *Instagram) consumer(results chan *InstagramResponse, done chan []InstagramData, simultaneous int, keepCriterion func(InstagramData)bool ) {
	var igData []InstagramData
	
	i := 0
	//Pull down results forever until we've hit some satisfaction criterion
	for {
		i++
		select {
		case res, ok := <-results:
			if !ok {
				fmt.Printf("I is %d but the channel is closed.", i)
				done <- igData //Semaphore{}
				return
			}
			fmt.Printf("%s is overall #%d\n", res, i)
			
			//Append
			for _, datum := range res.Data {
				if keepCriterion(datum) {
					igData = append(igData, datum)
				}
			}

			if !ig.satisfied(igData) {
				//If it's not happy after this result, the consumer
				// instructs a producer to start on something new
				//job := URL(random(300))
				fmt.Printf("Consumer is not satisfied after job #%d. Fetching %d\n", i, res.Pagination.NextURL)
				go ig.producer(results, res.Pagination.NextURL)
			} else {
				fmt.Printf("Consumer is satisfied after job #%d. Unlocking.\n", i)

				//Drain the channel
				drain(results, simultaneous)
				done <- igData //Semaphore{}
				return
			}
		}
	}
}

func (ig *Instagram) satisfied(igData []InstagramData) bool {
	if len(igData) > 30 {
		return true
	}

	return false
}

func drain(results chan *InstagramResponse, simultaneous int) {
	for i := 1; i < simultaneous; i++ {
		select {
		case res, ok := <-results:
			if !ok {
				break
			}
			fmt.Printf("Drained %s\n", res)
			//default:
			//	fmt.Println("Ostensibly no results to drain")
			//	close(results)
		}
	}
	
	fmt.Printf("Everything is drained. Closing the channel.\n")
	close(results)
}

//Make the request with the appropriate authorization and decode the response into json
func (ig *Instagram) getDecode(location string) (*InstagramResponse, error) {
	var data InstagramResponse

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

//Generate a URL string from a tag name
func (ig *Instagram) tagURL(tag string) string {
	//TODO -- error handling
	u, _ := url.Parse("https://api.instagram.com/v1/tags/" + tag + "/media/recent")
	
	//Set a higher count limit
	params := u.Query()
	params.Set("count", "25")
	
	ul := ig.BuildQuery(u, &params)
	
	fmt.Println(ul)
	
	return ul
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
