package p5

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

var FIRST_PEER_ADDRESS="localhost:6686"			//first peer's hard-coded address!

func init() {
		log.Println("REGISTRERATION SERVER's API Init method is triggered!")

}


func CarFormAPI(w http.ResponseWriter, r *http.Request) {
		log.Println("GetCarForm method is triggered!")

	//if r.URL.Path != "/getCarForm" {
	//	http.Error(w, "404 not found.", http.StatusNotFound)
	//	return
	//}

	switch r.Method {
	case "GET":
		log.Println("GET CarForm triggered!")

		dir, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("PWD:",dir)

		http.ServeFile(w, r, "CarForm.html")
	case "POST":
		log.Println("POST CarForm triggered!")
		// Call ParseForm() to parse the raw query and update r.PostForm and r.Form.
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		fmt.Fprintf(w, "HTTP Post sent to Registeration Server! PostForm = %v\n", r.PostForm)
		plate := r.FormValue("plate")
		mileage := r.FormValue("mileage")
		fmt.Fprintf(w, "plate = %s\n", plate)
		fmt.Fprintf(w, "mileage = %s\n", mileage)

		//TODO :
		//TransactonId
		//PublicKey
		//CarId
		//Km
		//Plate

	default:
		fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
	}
}







