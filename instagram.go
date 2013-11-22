package instagram

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"code.google.com/p/goauth2/oauth"
)

//API docs: http://instagram.com/developer

const (
	AuthorizationURL string = "https://api.instagram.com/oauth/authorize"
	AccessURL        string = "https://api.instagram.com/oauth/access_token"
	IgBaseURL        IgUrl  = IgUrl("https://api.instagram.com/v1/")
)

// OLD

func (igd *InstagramData) Created() time.Time {
	intVal, err := strconv.ParseInt(igd.CreatedTime, 10, 64)
	if err != nil {
		intVal = int64(0)
	}

	return time.Unix(intVal, 0)
}

//TagsMediaRecent accepts a slice of strings with the tag names (no hashtag needed) and
// then returns at least 30 results for that tag, and nil (or an error message).
func (ig *Instagram) TagsMediaRecent(tags []string) (*[]InstagramData, error) {
	simultaneous := len(tags)

	//Buffered channel of results
	results := make(chan *InstagramResponse, simultaneous) //Contains initial JSON responses from Instagram
	blocker := make(chan []InstagramData)                  //Contains the data satisfying consumers' keepCriterion

	for _, tag := range tags {
		//Generate the URL for each tag and pull down the instagram JSON response
		go ig.producer(results, ig.tagsMediaRecentURL(tag))
	}

	//The consumer needs to know how to decide whether each datum should be kept, and
	// this anonymous function will be applied by the consumer to make the decision.
	// Specifically here, we reject any photo that doesn't have ALL of the requested tags.
	keepCriterion := func(igDatum InstagramData) bool {
		hasTags := false
		for _, tag := range tags {
			hasTags = false
			for _, possessedTag := range igDatum.Tags {
				//Criterion is that the datum must actually have one of the tags we requested.
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

	//Interpret the JSON responses from each of the tags and pull down additional results, as needed,
	// until keepCriterion is met. Then, send the data back on chan blocker when the consumer is satisfied.
	go ig.consumer(results, blocker, simultaneous, keepCriterion)

	//Wait until the consumer is satisfied
	output := <-blocker

	//return &[]InstagramData{}, nil
	return &output, nil
}

func (ig *Instagram) LocationSearch(lat, long string) (*[]InstagramData, error) {
	simultaneous := 1

	//Buffered channel of results
	results := make(chan *InstagramResponse, simultaneous)
	blocker := make(chan []InstagramData)

	//How shall the consumer decide whether each datum should be kept?
	keepCriterion := func(igDatum InstagramData) bool { return true }

	go ig.producer(results, ig.locationSearchURL(lat, long))
	go ig.consumer(results, blocker, simultaneous, keepCriterion)

	//Wait until the consumer is satisfied
	output := <-blocker

	//return &[]InstagramData{}, nil
	return &output, nil
}

//Take whatever URL we get as 'work' and send back an InstagramResult object
func (ig *Instagram) producer(results chan *InstagramResponse, work string) {
	//Todo: Error handling
	r, err := ig.getDecode(work)
	if err != nil {
		r = &InstagramResponse{}
	}

	results <- r
	return
}

//TODO: make an error channel

//Interpret the JSON responses from each of the tags and pull down additional results, as needed,
// until keepCriterion is met. Then, send the data back on chan done when the consumer is satisfied.
// Timeout at 10 seconds per query, run N simultaneously.
func (ig *Instagram) consumer(results chan *InstagramResponse, done chan []InstagramData, simultaneous int, keepCriterion func(InstagramData) bool) {
	var igData []InstagramData

	timeout := make(chan bool)
	go func() {
		//TODO: Make timeout a config'able setting
		time.Sleep(10 * time.Second)
		timeout <- true
	}()

	i := 0
	//Pull down results forever until we've hit some satisfaction criterion
	for {
		i++
		select {

		case <-timeout:
			fmt.Printf("We timed out after %d attempts. Unlocking.\n", (i - 1))

			//Drain the channel
			go drain(results, simultaneous+1) //+1 for timeout case
			done <- igData                    //Semaphore{}
			fmt.Printf("Unlocked\n")
			return

		case res, ok := <-results:
			if !ok {
				fmt.Printf("I is %d but the channel is closed.", i)
				done <- igData //Semaphore{}
				fmt.Printf("Unlocked\n")
				return
			}
			fmt.Printf("%s is overall #%d\n", res, i)

			//Append
			for _, datum := range res.Data {
				if keepCriterion(datum) {
					igData = append(igData, datum)
				}
			}

			if !ig.satisfied(igData, i) {
				//If it's not happy after this result, the consumer
				// instructs a producer to start on something new
				//job := URL(random(300))
				fmt.Printf("Consumer is not satisfied after job #%d. Fetching next: %d\n", i, res.Pagination.NextURL)
				if res.Pagination.NextURL == "" {
					fmt.Printf("Consumer is not satisfied after job #%d but no next page was provided.", i)
					//1 fewer goroutine is running at the same time
					simultaneous = simultaneous - 1
					if simultaneous == 0 {
						go drain(results, simultaneous)
						//Nothing worked.
						done <- igData
						fmt.Printf("Unlocked\n")
						return
					}
				} else {
					go ig.producer(results, res.Pagination.NextURL)
				}
			} else {
				fmt.Printf("Consumer is satisfied after job #%d. Unlocking.\n", i)

				//Drain the channel
				go drain(results, simultaneous)
				done <- igData //Semaphore{}
				fmt.Printf("Unlocked\n")
				return
			}
		}
	}
}

func (ig *Instagram) satisfied(igData []InstagramData, i int) bool {
	//We quit after getting 30 results or after making 5 queries
	if len(igData) > 30 || i > 5 {
		return true
	}

	return false
}

func drain(results chan *InstagramResponse, simultaneous int) {
	for i := 1; i < simultaneous; i++ {
		select {
		case _, ok := <-results:
			if !ok {
				//Channel is closed. Quit.
				return
			}
			fmt.Printf("Drained a value\n")
		}
	}

	fmt.Printf("Everything is drained. Closing the channel.\n")
	//Note: Closing this gives deadlocks
	//Not closing gives memory leaks
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
func (ig *Instagram) tagsMediaRecentURL(tag string) string {
	//TODO -- error handling
	u, _ := url.Parse("https://api.instagram.com/v1/tags/" + tag + "/media/recent")

	//Set a higher count limit
	params := u.Query()
	params.Set("count", "100")

	ul := ig.BuildQuery(u, &params)

	fmt.Println(ul)

	return ul
}

//Generate a URL string from a location name
func (ig *Instagram) locationSearchURL(lat, long string) string {
	//TODO -- error handling
	u, _ := url.Parse("https://api.instagram.com/v1/locations/search")

	//Set a higher count limit
	params := u.Query()
	params.Set("lat", lat)
	params.Set("lng", long)

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
