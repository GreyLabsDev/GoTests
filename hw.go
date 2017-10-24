//Command to run test version:
//goapp serve app.yaml
//Command to deploy/update application:
//goapp deploy -application golangnode0 -version 0
//Command to test application without deploy
//goapp serve app.yaml

package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
)

var statusContent string = "Default status"

type webPage struct {
	Title string
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

func echo(ID string) {
	url := "http://goappnode" + ID + ".appspot.com" + "/status"

	var jsonStr = []byte(`{"message":"Web echo"}`)
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
		//go sendMail("Hello from test golang webapp!")
		//go sender()
		echo(1)
		fmt.Fprintf(w, "Successful read command/input from web-interface! Yeah! ")
	}
}

func statusServer(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		fmt.Fprintf(w, "Get status - "+statusContent)
	case "POST":
		r.ParseForm()
		fmt.Fprintf(w, "Get data by params in POST - OK")
		statusContent = "POST request - Responded"
	}
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
