package app

import (
	"bytes"
	"image"
	"io"
	"io/ioutil"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	_ "image/jpeg"
	_ "image/png"

	fb "github.com/huandu/facebook"

	"appengine"
	"appengine/datastore"
	"appengine/urlfetch"
)

var clientID = "526791527487217"
var FbApp = fb.New(clientID, APPSECRET)

var aboutParams = fb.Params{
	"method":       fb.GET,
	"relative_url": "me?fields=name,email,gender,age_range,address,location",
}

var photoParams = fb.Params{
	"method":       fb.GET,
	"relative_url": "me/picture?width=320&height=320&redirect=false",
}

func init() {
	http.HandleFunc("/static/", StaticHandler)
	http.HandleFunc("/", MainHandler)
	http.HandleFunc("/web/", WebHandler)

	fb.Debug = fb.DEBUG_ALL
	FbApp.EnableAppsecretProof = true
	rand.Seed(time.Now().UTC().UnixNano())
}

func WebHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	code := query.Get("code")

	if code == "" {
		http.Error(w, "Cannot get code from facebook.", 505)
		return
	}

	context := appengine.NewContext(r)
	client := urlfetch.Client(context)
	fb.SetHttpClient(client)

	redirectURL := "http://" + r.Host + r.URL.Path
	accessResp, err := fb.Get("/v2.4/oauth/access_token", fb.Params{
		"code":          code,
		"redirect_uri":  redirectURL,
		"client_id":     clientID,
		"client_secret": APPSECRET,
	})
	check(err, context)

	var accessToken string
	accessResp.DecodeField("access_token", &accessToken)

	paths := strings.Split(r.URL.Path, "/")
	party := paths[len(paths)-1]
	photoID := UploadPhoto(accessToken, party, context)
	redirectUrl := "https://facebook.com/photo.php?fbid=" + photoID + "&makeprofile=1&prof"
	http.Redirect(w, r, redirectUrl, 303)
}

func UploadPhoto(accessToken string, party string, context appengine.Context) string {
	session := FbApp.Session(accessToken)
	session.HttpClient = urlfetch.Client(context)
	err := session.Validate()
	check(err, context)

	results, err := session.BatchApi(aboutParams, photoParams)
	check(err, context)

	aboutBatch, err := results[0].Batch()
	check(err, context)
	photoBatch, err := results[1].Batch()
	check(err, context)

	aboutResp := aboutBatch.Result
	photoResp := photoBatch.Result

	SaveAboutUser(&aboutResp, party, context)

	profilePicture := GetUserPhoto(&photoResp, context)

	imagebytes := addLogo(profilePicture, party, context)
	form, mime := CreateImageForm(&imagebytes, context)

	url := "https://graph.facebook.com/v2.4/me/photos" +
		"?access_token=" + accessToken +
		"&no_story=true" +
		"&appsecret_proof=" + session.AppsecretProof()

	uploadResquest, _ := http.NewRequest("POST", url, form)
	uploadResquest.Header.Set("Content-Type", mime)
	uploadResp, _ := session.Request(uploadResquest)
	check(err, context)

	var photoID string
	uploadResp.DecodeField("id", &photoID)

	return photoID
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

	case r.URL.Path == "/privacy":
		context := appengine.NewContext(r)
		data, err := ioutil.ReadFile("privacy.html")
		check(err, context)
		w.Write(data)

	case r.URL.Path != "/":
		http.NotFound(w, r)

	default:
		handleMain(w, r)
	}

	return
}

func handleMain(w http.ResponseWriter, r *http.Request) {
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
	AgeRange int
	Location string
}

func SaveAboutUser(aboutResp *fb.Result, party string, context appengine.Context) {
	var log Log
	aboutResp.Decode(&log)

	var ageRange map[string]int
	var location map[string]string
	aboutResp.DecodeField("location", &location)
	aboutResp.DecodeField("age_range", &ageRange)

	log.AgeRange = ageRange["min"]
	log.Location = location["name"]
	log.Party = party

	_, err := datastore.Put(context,
		datastore.NewIncompleteKey(context, "log", nil),
		&log)
	check(err, context)
}

func GetUserPhoto(photoResp *fb.Result, context appengine.Context) *image.Image {
	var dataField fb.Result
	photoResp.DecodeField("data", &dataField)

	var url string
	dataField.DecodeField("url", &url)

	client := urlfetch.Client(context)
	resp, err := client.Get(url)
	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	check(err, context)

	reader := bytes.NewReader(data)
	profilePicture, _, err := image.Decode(reader)
	check(err, context)

	return &profilePicture
}

func CreateImageForm(imageBytes *[]byte, context appengine.Context) (*bytes.Buffer, string) {
	var formBuffer bytes.Buffer
	multiWriter := multipart.NewWriter(&formBuffer)

	imageField, err := multiWriter.CreateFormFile("source", "election.png")
	check(err, context)

	imageBuffer := bytes.NewBuffer(*imageBytes)
	_, err = io.Copy(imageField, imageBuffer)
	check(err, context)

	multiWriter.Close()
	return &formBuffer, multiWriter.FormDataContentType()
}
