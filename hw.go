package main

import (
	"fmt"
	"net/http"
)

func helloWorld(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World!")
}

func startPage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello, test server started on 8080 port.\n - /helloworld - show title page\n - /showinfo - show information about this thing")
}

func showInfo(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Inforamtion page for test project.\nLanguage - Go\nPlatform - Google Application Engine")
}

func init() {

	http.HandleFunc("/", startPage)
	http.HandleFunc("/helloworld", helloWorld)
	http.HandleFunc("/showinfo", showInfo)
	//Wrong code for App Enine - server cant understand what it need to show
	//http.ListenAndServe(":80", nil)
}

/*
func main() {
	fmt.Println("Hello, test server started on 80 port.\n - /helloworld - show title page\n - /showinfo - show information about this thing")
	http.HandleFunc("/", startPage)
	http.HandleFunc("/helloworld", helloWorld)
	http.HandleFunc("/showinfo", showInfo)
	http.ListenAndServe(":80", nil)
}
*/
//goapp serve app.yaml
//goapp deploy -application golangnode0 -version 0
