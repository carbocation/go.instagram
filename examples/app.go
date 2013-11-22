package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/carbocation/go.instagram"
	"github.com/gorilla/mux"
)

type tplData struct {
	Json  *instagram.InstagramResponse
	Query string
}

var ClientId string = "f38ffd3ce6284722a4dd6af96514a147"
var ig *instagram.Instagram

func init() {
	err := error(nil)
	ig, err = instagram.NewInstagram(ClientId)
	if err != nil {
		log.Println(err)
	}
}

func tag(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)

	url := instagram.UrlTagsMediaRecent(v["tag"], "", "")
	result, err := ig.QueryPublic(url)
	if err != nil {
		log.Println("ERROR:", err)
		return
	}

	data := tplData{Json: result, Query: url.String()}

	log.Println(url)
	template.Must(template.New("searchTags").Parse(tpl.all)).Execute(w, data)

	return
}

func tagInfo(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)

	url := instagram.UrlTags(v["tag"])
	result, err := ig.QueryPublic(url)
	if err != nil {
		log.Println("ERROR:", err)
		log.Printf("%+v\n", result)
		return
	}

	data := tplData{Json: result, Query: url.String()}

	log.Println(url)
	template.Must(template.New("tagInfo").Parse(tpl.all)).Execute(w, data)

	return
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/tag/{tag}", tag)
	r.HandleFunc("/tagInfo/{tag}", tagInfo)

	fmt.Println("Try http://localhost:9999/tags/selfie")
	http.Handle("/", r)
	http.ListenAndServe(":9999", nil)
}

var tpl = struct {
	all string
}{
	all: `{{define "searchTags"}}
<html>
<head>
<title>
{{.Query}}
</title>
</head>
<body>
<h1>You searched for {{.Query}}</h1>
<br />
<br />
{{range .Json.Data}}
	{{template "parseData" .}}
	<br />
	<br />
{{end}}
</body>
</html>
{{end}}

{{define "tagInfo"}}
<html>
<head>
<title>
{{.Query}}
</title>
</head>
<body>
<h1>You searched for {{.Query}}</h1>
<br />
<br />
{{.Json}}
</body>
</html>
{{end}}

{{define "parseData"}}
<div>
	By {{.User.Username}} on {{.Created}}
	<br />
	{{ if .Location }}
		In {{.Location.Name}}<br />
		Lat: {{.Location.Latitude}}<br />
		Long: {{.Location.Longitude}}<br />
		<br />
	{{end}}
	Tags: 
	{{range .Tags}}#{{.}}, {{end}}
	<br />
	{{with .Images.LowResolution}}
		<img src="{{.URL}}" width={{.Width}} height={{.Height}}>
	{{end}}
	<br />
</div>
{{end}}

{{define "welcome"}}
<html>
<body>
<h1>Welcome</h1>
<a href="/tags/golang%20gopher">Try an example search for tags containing 'golang' and 'gopher'</a>
</body>
</html>
{{end}}

{{define "error"}}
<h1>Error</h1>
{{.Error}}
<br />
This is usually because they blocked a query with adult language.
{{end}}`,
}
