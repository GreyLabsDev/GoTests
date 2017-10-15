package main

import (
	"fmt"
	"net/http"
	"os"
)

func helloWorld(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World")
}

func startPage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello, test server started on 8080 port.\n - /helloworld - show title page\n - /exit - turn off this server")
}

func stopServer(w http.ResponseWriter, r *http.Request) {
	os.Exit(3)
}

func main() {
	fmt.Println("Hello, test server started on 8080 port.\n - /helloworld - show title page\n - /exit - turn off this server")
	http.HandleFunc("/", startPage)
	http.HandleFunc("/helloworld", helloWorld)
	http.HandleFunc("/exit", stopServer)
	http.ListenAndServe(":8080", nil)
}
