package main

import (
	"log"
	"net/http"
	"os"
	"./p5"
)

func main() {


	router := p5.NewRouter()
	if len(os.Args) > 1 {
		log.Fatal(http.ListenAndServe(":"+os.Args[1], router))
	} else {
		log.Fatal(http.ListenAndServe(":8080", router))
	}


}






