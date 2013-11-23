package instagram

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"time"

	"fmt"
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
		ID        float64 `json:",string"`
		Latitude  float64
		Longitude float64
		Name      string
	}
	MediaCount   float64 `json:"media_count"`
	Name         string
	Tags         []string
	Type         string
	User         InstagramUser
	UsersInPhoto []InstagramUser
}

type InstagramResponse struct {
	Meta struct {
		ErrorType    string `json:"error_type"`   //Only if errors occurred
		ErrorMessage string `json:error_message"` //Only if errors occurred
		Code         int
		URL          string `json:"-"` //This is not part of Instagram's output
	}
	DataInterface interface{}     `json:"data"` //Because Instagram has 2 different types of "Data"
	Data          []InstagramData `json:"-"`    //Because Instagram has 2 different types of "Data"
	Pagination    struct {
		MaxTagID  string `json:"max_tag_id"` //Deprecated?
		MinTagID  string `json:"min_tag_id"` //Deprecated?
		NextURL   string `json:"next_url"`
		NextMaxID string `json:"next_max_id"`
	}
}

//Query via the public API / without any user-specific authentication, just your app's client_id
func (ig *Instagram) QueryPublic(location IgUrl) (*InstagramResponse, error) {
	var response InstagramResponse

	//Make sure we put our client_id into the query
	location = location.addClientId(ig.ClientId)

	//Store location in the struct so we can access our current URL if we feel like it
	response.Meta.URL = location.String()

	//Actually hit the Instagram servers
	query, err := http.Get(location.String())
	defer query.Body.Close()
	body, err := ioutil.ReadAll(query.Body)
	if err != nil {
		return &response, err
	}

	//Parse the returned JSON
	if err := json.Unmarshal(body, &response); err != nil {
		return &response, err
	}

	//fmt.Println("TypeOf response", reflect.TypeOf(response.DataInterface))

	//Now for the data portion, re-encapsulate it and plan to re-parse:
	di, err := json.Marshal(response.DataInterface)
	if err != nil {
		fmt.Println("Error in marshaling?", err)
	}

	di2, err := json.Marshal(response.DataInterface)
	if err != nil {
		fmt.Println("Error in marshaling?", err)
	}

	//fmt.Println(string(di))

	darray := []InstagramData{}
	//Try to unmarshal the data as []InstagramData first
	if err := json.Unmarshal(di, &darray); err != nil {
		fmt.Println("Error as array:", err)
		fmt.Println(response.DataInterface)
		dcontainer := InstagramData{}
		if err := json.Unmarshal(di2, &dcontainer); err != nil {
			fmt.Println("Error as object:", err)
			return &response, err
		} else {
			//fmt.Println("response.Data IS InstagramData{}")
			response.Data = []InstagramData{dcontainer}
		}
	} else {
		//fmt.Println("response.Data IS []InstagramData{}{}")
		response.Data = darray
	}

	return &response, nil
}

func (igd *InstagramData) Created() time.Time {
	intVal, err := strconv.ParseInt(igd.CreatedTime, 10, 64)
	if err != nil {
		intVal = int64(0)
	}

	return time.Unix(intVal, 0)
}
