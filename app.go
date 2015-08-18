package app

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"image"
	_ "image/jpeg"
	_ "image/png"

	"appengine"
	"appengine/datastore"
	"appengine/urlfetch"
)

func init() {
	http.HandleFunc("/static/", StaticHandler)
	http.HandleFunc("/", MainHandler)
}

func StaticHandler(w http.ResponseWriter, r *http.Request) {
	path := "." + r.URL.Path

	if f, err := os.Stat(path); err == nil && !f.IsDir() {
		http.ServeFile(w, r, path)
		return
	}

	http.NotFound(w, r)
}

func MainHandler(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path != "/":
		http.NotFound(w, r)
	case r.Method == "GET":
		handleGet(w, r)
	case r.Method == "POST":
		handlePost(w, r)
	}

	return
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	context := appengine.NewContext(r)
	data, err := ioutil.ReadFile("index.html")
	check(err, context)
	w.Write(data)
}

type Log struct {
	Name     string
	Gender   string
	Party    string
	Email    string
	Birthday string
	Hometown string
}

func handlePost(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	url := r.FormValue("url")
	party := r.FormValue("party")

	log := Log{
		Name:     r.FormValue("name"),
		Email:    r.FormValue("email"),
		Gender:   r.FormValue("gender"),
		Party:    party,
		Birthday: r.FormValue("birthday"),
		Hometown: r.FormValue("hometown"),
	}

	context := appengine.NewContext(r)
	client := urlfetch.Client(context)

	_, err := datastore.Put(context,
		datastore.NewIncompleteKey(context, "log", nil),
		&log)
	check(err, context)

	resp, err := client.Get(url)
	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	check(err, context)

	reader := bytes.NewReader(data)
	profilePicture, _, err := image.Decode(reader)
	check(err, context)

	imagebytes := addLogo(&profilePicture, party, context)
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Length", strconv.Itoa(len(imagebytes)))
	_, err = w.Write(imagebytes)
	check(err, context)
}
