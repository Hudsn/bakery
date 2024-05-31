package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/hudsn/bakery"
	"github.com/hudsn/bakery/internal/example/handlers"
	"github.com/hudsn/bakery/internal/example/otherstatic"
	"github.com/hudsn/bakery/internal/example/static"
	"github.com/hudsn/bakery/internal/example/views"
)

func main() {
	isDev := flag.Bool("dev", false, "whether the environment is dev or not")
	port := flag.Int("port", 9001, "set port to listen on")
	flag.Parse()

	myBakery := views.New(*isDev) // we register all our template paths and stuff in the Views folder/package, then call it here.
	scriptHandler, err := bakery.MakeScriptHandler("/tepid", *port)
	if err != nil {
		log.Fatal(err)
	}
	sseHandler, err := bakery.MakeSSEHandler(myBakery)
	if err != nil {
		log.Fatal(err)
	}
	staticHandler := bakery.MakeStaticHandler("/static/", "internal/example/static", static.StaticFS, *isDev)
	singleFileHandler := bakery.MakeSingleFileHandler("internal/example/otherstatic/page.html", otherstatic.PageStatic, "text/html", handlers.DefaultErrorHandler, *isDev)
	iframeHandlerHomePage, err := bakery.MakeIframeWatcher("/", "/tepid.js")
	if err != nil {
		log.Fatal(err)
	}

	router := http.NewServeMux()

	if *isDev {
		router.HandleFunc("/tepid.js", scriptHandler)
		router.HandleFunc("/tepid", sseHandler)
		router.HandleFunc("/index/dev", iframeHandlerHomePage)
	}

	router.HandleFunc("/", singleFileHandler)
	router.HandleFunc("/home", handlers.HomeHandler(myBakery))
	router.HandleFunc("/static/", staticHandler)

	fmt.Printf("Listening on :%d...\n", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), router))

}
