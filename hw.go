//Command to test application without deploy:
//goapp serve app.yaml
//Command to deploy/update application:
//goapp deploy -application golangnode0 -version 0

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"

	"appengine"
	"appengine/urlfetch"
)

var statusContent string = "Default status"

type webPage struct {
	Title string
}

type nodeStats struct {
	NodeID           int    `json:"ID"`
	HasTask          bool   `json:"hasTask"`
	TaskStatus       string `json:"taskStatus"`
	TaskResult       string `json:"taskResult"`
	TaskFragmentBody string `json:"taskFragmentBody"`
	TaskBody         string `json:"taskBody"`
}

type echoMessage struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type gmailUser struct {
	name string
	pswd string
}

func sendMail(msg string) {
	mailUser := gmailUser{
		"golangapplication@gmail.com",
		"",
	}
	auth := smtp.PlainAuth("",
		mailUser.name,
		mailUser.pswd,
		"smtp.gmail.com",
	)
	err := smtp.SendMail(
		"smtp.gmail.com:587",
		auth,
		mailUser.name,
		[]string{"rec@mail.com"},
		[]byte(msg),
	)
	if err != nil {
		log.Fatal(err)
	}
}

//wrong func for Google App Engine deployment. Need to use appengine libs...=(
func echo() {

	url := "http://golangappnode1.appspot.com/status"

	var jsonStr = []byte(`{"msg":"Hello!"}`)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	statusContent = string(body)

}

func helloWorld(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World!")
}

func startPage(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		templatePage, _ := template.ParseFiles("start.html")
		templatePage.Execute(w, &webPage{"simplePage"})
	case "POST":
		r.ParseForm()
		//go echo()
		//fmt.Fprintf(w, "Successful read command/input from web-interface! Input contains - "+r.FormValue("nodeId")+" "+r.FormValue("echoContent"))
	}
}

func statusServer(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		fmt.Fprintf(w, "Get status - "+statusContent)
	case "POST":
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		newStr := buf.String()

		thisNodeStats := nodeStats{
			1,
			false,
			"not running",
			"empty",
			"empty fragment",
			"empty",
		}

		jsonNodeStats, err1 := json.Marshal(thisNodeStats)
		if err1 != nil {
			panic(err1)
		}

		fmt.Fprintf(w, "Get data by params in POST - OK "+string(jsonNodeStats))
		//statusContent = "POST request handled, " + "Node id: " + string(nodeSends.id) + ", Echo content: " + nodeSends.content
		statusContent = "POST request handled, " + newStr
	}
}

func testEcho(w http.ResponseWriter, r *http.Request) {
	msg := echoMessage{
		"Message is",
		"",
	}

	r.ParseForm()
	c := appengine.NewContext(r)
	msg.Content = r.FormValue("echoContent")

	jsonMessage, err2 := json.Marshal(msg)
	if err2 != nil {
		panic(err2)
	}

	//jsonStr := []byte(`{"message":"` + r.FormValue("echoContent") + `"}`)
	jsonStr := []byte(jsonMessage)
	buf := bytes.NewBuffer(jsonStr)
	client := http.Client{Transport: &urlfetch.Transport{Context: c}}
	resp, err := client.Post("http://goappnode"+r.FormValue("nodeId")+".appspot.com/status", "application/octet-stream", buf)
	if err != nil {
		statusContent = err.Error()
		fmt.Println(err)
	}
	respBody, _ := ioutil.ReadAll(resp.Body)
	statusContent = "Response from node - " + string(respBody)
}

func showInfo(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Information page for test project.")
	fmt.Fprintln(w, "Language - Go;")
	fmt.Fprintln(w, "Platform - Google Application Engine;")
}

func init() {
	http.HandleFunc("/", startPage)
	http.HandleFunc("/helloworld", helloWorld)
	http.HandleFunc("/showinfo", showInfo)
	http.HandleFunc("/status", statusServer)
	http.HandleFunc("/echo", testEcho)

	//Wrong code for App Enine - server cant understand what it need to show
	//http.ListenAndServe(":80", nil)
}

//this func not needed for deploy on Google App Engine, init() func replace main()
/*
func main() {
	//fmt.Println("Hello, test server started on 8080 port.\n - /helloworld - show title page\n - /showinfo - show information about this thing")
	//http.ListenAndServe(":8080", nil)
	go sender()
}*/
