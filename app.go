package app

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"appengine"
	"appengine/urlfetch"

	"image"

	_ "image/jpeg"
	_ "image/png"
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
	data, err := ioutil.ReadFile("index.html")
	check(err)

	w.Write(data)
}

func handlePost(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	url := r.FormValue("url")

	context := appengine.NewContext(r)
	client := urlfetch.Client(context)

	resp, err := client.Get(url)
	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	check(err)

	reader := bytes.NewReader(data)
	profilePicture, _, err := image.Decode(reader)
	check(err)

	imagebytes := OverlayLogo(profilePicture, "NLD")

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Length", strconv.Itoa(len(imagebytes)))
	_, err = w.Write(imagebytes)
	check(err)
}
