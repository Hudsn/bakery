package handlers

import (
	"net/http"

	"github.com/hudsn/bakery"
	"github.com/hudsn/bakery/internal/example/views/layouts"
)

func DefaultAppData(b bakery.Bakery) layouts.AppLayoutData {
	retVal := layouts.AppLayoutData{}
	retVal.SetIsDev(b.IsDev())
	retVal.SetTitle("My App!")

	return retVal
}

func HomeHandler(b bakery.Bakery) http.HandlerFunc {
	data := DefaultAppData(b)
	data.SetTitle("Home")

	return func(w http.ResponseWriter, r *http.Request) {
		b.Bake("home", data).ServeHTTP(w, r)
	}
}

func DefaultErrorHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}
