package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/carbocation/go-instagram"
	"github.com/gorilla/mux"
)

func init() {
	/*
		In my settings.go file, I have the following:
		igCfg := &instagram.Cfg{
			ClientID:     "", //Fill in as appropriate
			ClientSecret: "",
			RedirectURL:  "",
		}
	*/
	instagram.Initialize(igCfg)
}

func hello(w http.ResponseWriter, r *http.Request) {
	ig, err := instagram.NewInstagram()
	if err != nil {
		log.Println(err)
		return
	}

	url := ig.Authenticate("")
	log.Println(url)
	http.Redirect(w, r, url, http.StatusFound)
	
	return
}

func RedirectHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("inside redirect")
	rbody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(string(rbody))
	fmt.Fprintf(w, "%s", string(rbody))
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", hello)
	r.HandleFunc("/redirect/instagram", RedirectHandler)
	http.Handle("/", r)
	http.ListenAndServe(":9999", nil)
}
