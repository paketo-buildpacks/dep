package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/ZiCog/shiny-thing/foo"
)

func main() {
	foo.Do()
	http.HandleFunc("/", hello)
	fmt.Println("listening...")
	port := "8080"
	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic(err)
	}
}

func hello(res http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(res, "Hello, World!")
}
