//Command to run test version:
//goapp serve app.yaml
//Command to deploy/update application:
//goapp deploy -application golangnode0 -version 0

package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/smtp"
)

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

func sender() {
	from := "golangapplication@gmail.com"
	auth := smtp.PlainAuth("", from, "glob456987dss@#", "smtp.gmail.com")
	err := smtp.SendMail(
		"smtp.gmail.com:587", // server address
		auth,                 // authentication
		from,                 // sender's address
		[]string{"greyson.dean@gmail.com"}, // recipients' address
		[]byte("Hello World!"),             // message body
	)
	if err != nil {
		log.Print(err)
	}
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
		go sender()
		fmt.Fprintf(w, "Successful read command/input from web-interface! Yeah! ")
	}
}

func showInfo(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Information page for test project.")
	fmt.Fprintln(w, "Language - Go;")
	fmt.Fprintln(w, "Platform - Google Application Engine;")
}

/*
func sendEmail(w http.ResponseWriter, r *http.Redirect) {

}
*/

func init() {
	http.HandleFunc("/", startPage)
	http.HandleFunc("/helloworld", helloWorld)
	http.HandleFunc("/showinfo", showInfo)
	//http.HandleFunc("/save", showInfo)

	//Wrong code for App Enine - server cant understand what it need to show
	//http.ListenAndServe(":80", nil)
}

//this func not needed for deploy on Google App Engine, init() func replace main()

func main() {
	//fmt.Println("Hello, test server started on 8080 port.\n - /helloworld - show title page\n - /showinfo - show information about this thing")
	//http.ListenAndServe(":8080", nil)
	go sender()
}
