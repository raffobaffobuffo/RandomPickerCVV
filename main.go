package main

import (
	"log"
	"net/http"
	"net/url"
	"io/ioutil"
	"net/http/cookiejar"
	"encoding/json"
	"math/rand"
	"time"
	"os"
)

type Session struct {
	Client http.Client
	uid string
	pwd string
}

func newSession(uid string, pwd string) (session *Session) {
	jar, err := cookiejar.New(nil)
	if err != nil { log.Fatal(err) }
	client := http.Client{Jar:jar}
	return &Session{client, uid, pwd}
}

func (session *Session) loginCVV(uid string, pwd string) (status bool) {
	req, err := session.Client.PostForm("https://web.spaggiari.eu/auth-p7/app/default/AuthApi4.php?a=aLoginPwd",
		url.Values{"uid": {uid}, "pwd": {pwd}})
	if err != nil {
		return false
	}
	b, _ := ioutil.ReadAll(req.Body)
	req.Body.Close()
	var mapRes map[string]interface{}
	json.Unmarshal(b, &mapRes)
	data, _ := mapRes["data"].(map[string]interface{})
	auth, _ := data["auth"].(map[string]interface{})
	return auth["loggedIn"].(bool)
}

func (session *Session) getClassmates() (classmates []string, err error) {
	res, err := session.Client.Get("https://web.spaggiari.eu/sps/app/default/SocMsgApi.php?a=acGetRubrica")
	if err != nil {
		return
	}
	b, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()
	var response map[string]interface{}
	json.Unmarshal(b, &response)
	targets, _ := response["OAS"].(map[string]interface{})
	persone, _ := targets["targets"].(map[string]interface{})
	mates, _ := persone["persone"].(map[string]interface{})
	for k := range mates {
		info, _ := mates[k].(map[string]interface{})
		if info["role"] != "doc" {
			name := info["persona"].(string)
			classmates = append(classmates, name)
		}
	}
	return
}

func draw(classmates []string) (name string) {
	rand.Seed(time.Now().Unix())
	name = classmates[rand.Intn(len(classmates))]
	return
}

func (session *Session) drawHandler (w http.ResponseWriter, r *http.Request) {
	log.Print("/ GET")
	classmates, err := session.getClassmates()
	if err != nil {
		response := make(map[string]string)
		response["Error"] = "Data not founde"
		jsonresponse, _ := json.Marshal(response)
		w.Write(jsonresponse)
	}
	name := draw(classmates)
	response := make(map[string]string)
	response["DRAW"] = name
	jsonresponse, _ := json.Marshal(response)
	w.Write(jsonresponse)
}

func main() {
	uid := os.Args[1]
	pwd := os.Args[2]
	log.Print("Server starting...")
	session := newSession(uid, pwd)
	login := session.loginCVV(uid, pwd)
	if !login { log.Fatal("Server disconnecting...") }
	log.Print("Server connected")
	http.HandleFunc("/", session.drawHandler)
	http.ListenAndServe(":8000", nil)
}
