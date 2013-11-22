package instagram

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"time"
)

//API docs: http://instagram.com/developer
const (
	IgBaseURL IgUrl = IgUrl("https://api.instagram.com/v1/")
)

func NewInstagram(clientId string) (*Instagram, error) {
	return &Instagram{
		ClientId: clientId,
	}, nil
}

type Instagram struct {
	ClientId string
	*sync.Mutex
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
	MediaCount   string `json:"media_count"`
	Name         string
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
type InstagramResponse2 struct {
	Meta struct {
		Code int
		URL  string `json:"-"` //This is not part of Instagram's output
	}
	Data       InstagramData
	Pagination struct {
		MaxTagID string `json:"max_tag_id"`
		MinTagID string `json:"min_tag_id"`
		NextURL  string `json:"next_url"`
	}
}
*/

//Query via the public API / without any user-specific authentication, just your app's client_id
func (ig *Instagram) QueryPublic(location IgUrl) (*InstagramResponse, error) {
	var data InstagramResponse

	//Make sure we put our client_id into the query
	location = location.addClientId(ig.ClientId)

	//Store location in the struct so we can access our current URL if we feel like it
	data.Meta.URL = location.String()

	//Actually hit the Instagram servers
	resp, err := http.Get(location.String())
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &data, err
	}

	//Parse the returned JSON
	if err := json.Unmarshal(body, &data); err != nil {
		return &data, err
	}

	return &data, nil
}

func (igd *InstagramData) Created() time.Time {
	intVal, err := strconv.ParseInt(igd.CreatedTime, 10, 64)
	if err != nil {
		intVal = int64(0)
	}

	return time.Unix(intVal, 0)
}
