package p5

import "net/http"

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

var routes = Routes{
	Route{
		"GetCarForm",
		"GET",
		"/getCarForm",
		CarFormAPI,
	},
	Route{
		"PostCarForm",
		"POST",
		"/postCarForm",
		CarFormAPI,
	},

}
