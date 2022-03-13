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
	"fmt"
	"strconv"
)

type Session struct {
	Client http.Client
	uid string
	pwd string
	ClassMates []string
}

func newSession(uid string, pwd string) (session *Session) {
	jar, err := cookiejar.New(nil)
	if err != nil { log.Fatal(err) }
	client := http.Client{Jar:jar}
	return &Session{client, uid, pwd, nil}
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
	accountInfo, _ := auth["accountInfo"].(map[string]interface{})
	if err == nil {
		name := fmt.Sprintf("%s %s", accountInfo["nome"].(string), accountInfo["cognome"].(string))
		session.ClassMates = append(session.ClassMates, name)
	}
	return auth["loggedIn"].(bool)
}

func (session *Session) getClassmates() (err error) {
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
			session.ClassMates = append(session.ClassMates, name)
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
	log.Print("/draw GET")
	name := draw(session.ClassMates)
	response := make(map[string]string)
	response["DRAW"] = name
	jsonresponse, _ := json.Marshal(response)
	w.Write(jsonresponse)
}

func (session *Session) getClassHandler (w http.ResponseWriter, r *http.Request) {
	log.Print("/getClass GET")
	class := make(map[int]string)
	for i, mate := range session.ClassMates {
		class[i] = mate
	}
	response := make(map[string]map[int]string)
	response["CLASS"] = class
	response["LEN"] = map[int]string{len(class): "classmates"}
	jsonresponse, _ := json.Marshal(response)
	w.Write(jsonresponse)
}

func (session *Session) reloadClassHandler (w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Print("/reload method not allowed")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	log.Print("/reload POST")
	session.ClassMates = nil
	session.getClassmates()
	response, _ := json.Marshal(map[string]string{"Message": "Data reloaded"})
	w.Write(response)
}

func (session *Session) addMateHandler (w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Print("/addMate method not allowed")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	log.Print("/addMate POST")
	name := r.FormValue("name")
	if name == "" {
		response, _ := json.Marshal(map[string]string{"Message": "Failed request"})
		w.Write(response)
		return
	}
	session.ClassMates = append(session.ClassMates, name)
	response, _  := json.Marshal(map[string]string{"Message": "Success"})
	w.Write(response)
}

func (session *Session) removeMateHandler (w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Print("/removeMate Method not allowed")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	log.Print("/removeMate POST")
	form := r.FormValue("id")
	id, err := strconv.Atoi(form)
	if err != nil || id >= len(session.ClassMates) {
		response, _ := json.Marshal(map[string]string{"Message": "Failed request"})
		w.Write(response)
		return
	}
	session.ClassMates[id] = session.ClassMates[len(session.ClassMates)-1]
	session.ClassMates = session.ClassMates[:len(session.ClassMates)-1]
	response, _  := json.Marshal(map[string]string{"Message": "Success"})
	w.Write(response)
}

func main() {
	uid := os.Args[1]
	pwd := os.Args[2]
	log.Print("Server starting...")
	session := newSession(uid, pwd)
	login := session.loginCVV(uid, pwd)
	if !login { log.Fatal("Server disconnecting...") }
	log.Print("Server connected")
	log.Print("Loading classmates...")
	err := session.getClassmates()
	if err != nil { log.Fatal(err) }
	log.Print("Classmates loaded")
	http.HandleFunc("/draw", session.drawHandler)
	http.HandleFunc("/getClass", session.getClassHandler)
	http.HandleFunc("/addMate", session.addMateHandler)
	http.HandleFunc("/removeMate", session.removeMateHandler)
	http.HandleFunc("/reload", session.reloadClassHandler)
	http.ListenAndServe(":8000", nil)
}
