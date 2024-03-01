package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/hudsn/bakery/example/handlers"
	"github.com/hudsn/bakery/example/views"
)

func main() {
	isDev := flag.Bool("dev", false, "whether the environment is dev or not")
	port := flag.Int("port", 9001, "set port to listen on")
	flag.Parse()

	bakery := views.New(*isDev) // we register all our template paths and stuff in the Views package, then call it here.
	scriptHandler, err := bakery.MakeScriptHandler("/tepid", *port)
	if err != nil {
		log.Fatal(err)
	}
	sseHandler, err := bakery.MakeSSEHandler()
	if err != nil {
		log.Fatal(err)
	}
	staticHandler := bakery.MakeStaticHandler("/static/")

	router := http.NewServeMux()
	if *isDev {
		router.HandleFunc("/tepid.js", scriptHandler)
		router.HandleFunc("/tepid", sseHandler)
	}

	router.HandleFunc("/home", handlers.HomeHandler(bakery))
	router.HandleFunc("/static/", staticHandler)

	fmt.Printf("Listening on :%d...\n", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), router))

}
