package instagram

import (
	"net/http"
	"sync"

	"code.google.com/p/goauth2/oauth"
)

//Contains miscellaneous types

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

	mu           *sync.Mutex
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
