//Command to run test version:
//goapp serve app.yaml
//Command to deploy/update application:
//goapp deploy -application golangnode0 -version 0

package main

import (
	"fmt"
	"html/template"
	"net/http"
)

type webPage struct {
	Title string
}

func helloWorld(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World!")
}

func startPage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, test application started.")
	fmt.Fprintln(w, "\n - /helloworld - show title page")
	fmt.Fprintln(w, "\n - /showinfo - show information about this thing")
}

func showInfo(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Information page for test project.")
	fmt.Fprintln(w, "Language - Go;")
	fmt.Fprintln(w, "Platform - Google Application Engine;")
}

func pageRender(w http.ResponseWriter, r *http.Request) {
	templatePage, _ := template.ParseFiles("index.html")
	templatePage.Execute(w, &webPage{"simplePage"})
}

/*
func sendEmail(w http.ResponseWriter, r *http.Redirect) {

}
*/

func init() {
	//fs := http.FileServer(http.Dir("/home/greyson_dean/testGoApp/files"))

	http.HandleFunc("/", startPage)
	http.HandleFunc("/helloworld", helloWorld)
	http.HandleFunc("/showinfo", showInfo)
	//http.Handle("/files/", http.StripPrefix("/files", fs))
	//Wrong code for App Enine - server cant understand what it need to show
	//http.ListenAndServe(":80", nil)
}

//this func not needed for deploy on Google App Engine, init() func replace main()
/*
func main() {
	fmt.Println("Hello, test server started on 8080 port.\n - /helloworld - show title page\n - /showinfo - show information about this thing")
	init()
	http.ListenAndServe(":8080", nil)
}
*/
