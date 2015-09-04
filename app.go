package app

import (
	"bytes"
	"image"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	_ "image/jpeg"
	_ "image/png"

	fb "github.com/huandu/facebook"

	"appengine"
	"appengine/datastore"
	"appengine/urlfetch"
)

var FbApp = fb.New("526791527487217", "e314e5fc761425d59ea9e2666c63e108")

func init() {
	http.HandleFunc("/static/", StaticHandler)
	http.HandleFunc("/", MainHandler)

	fb.Debug = fb.DEBUG_ALL
	FbApp.EnableAppsecretProof = true
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
	AgeRange string
	Hometown string
}

func handlePost(w http.ResponseWriter, r *http.Request) {
	context := appengine.NewContext(r)
	client := urlfetch.Client(context)

	r.ParseForm()
	access_token := r.FormValue("access_token")
	party := r.FormValue("party")
	context.Infof("party = %s", party)

	session := FbApp.Session(access_token)
	session.HttpClient = client
	err := session.Validate()
	check(err, context)

	aboutResp, err := session.Get("/me", fb.Params{
		"fields": "name,email,gender,age_range,hometown",
	})
	check(err, context)

	var log Log
	aboutResp.Decode(&log)

	var ageRange map[string]string
	aboutResp.DecodeField("age_range", &ageRange)
	log.AgeRange = ageRange["min"]

	context.Infof("Name = %s", log.Name)
	context.Infof("Email = %s", log.Email)
	context.Infof("Gender = %s", log.Gender)
	context.Infof("AgeRange = %s", log.AgeRange)
	context.Infof("Hometown = %s", log.Hometown)

	_, err = datastore.Put(context,
		datastore.NewIncompleteKey(context, "log", nil),
		&log)
	check(err, context)

	photoResp, err := session.Get("/me/picture?width=320&height=320&redirect=false", nil)
	check(err, context)

	var dataField fb.Result
	photoResp.DecodeField("data", &dataField)

	var url string
	dataField.DecodeField("url", &url)

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
