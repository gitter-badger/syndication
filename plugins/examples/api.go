package main

import "net/http"

func subscriptionsHandler(c Context, r http.Request, w *http.Response) {
}

func Initialize(r *Registrar) {
	r.registerEndpoint("/fever/subscriptions", "GET", subscriptionsHandler)
}

func Shutdown() {

}
