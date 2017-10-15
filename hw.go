//Command to run test version:
//goapp serve app.yaml
//Command to deploy/update application:
//goapp deploy -application golangnode0 -version 0

package main

import (
	"fmt"
	"net/http"
)

func helloWorld(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World!")
}

func startPage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello, test application started.")
	fmt.Fprintln(w, " - /helloworld - show title page")
	fmt.Fprintln(w, " - /showinfo - show information about this thing")
}

func showInfo(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Inforamtion page for test project.")
	fmt.Fprintf(w, "Language - Go")
	fmt.Fprintf(w, "Platform - Google Application Engine")
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
	fmt.Println("Hello, test server started on 8080 port.\n - /helloworld - show title page\n - /showinfo - show information about this thing")
	http.HandleFunc("/", startPage)
	http.HandleFunc("/helloworld", helloWorld)
	http.HandleFunc("/showinfo", showInfo)
	http.ListenAndServe(":8080", nil)
}
*/
