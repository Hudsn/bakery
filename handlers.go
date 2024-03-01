package bakery

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"os"
)

//go:embed sse_consumer.go.html
var sseConsumer embed.FS

func (b Bakery) MakeStaticHandler(route string) http.HandlerFunc {
	handler := http.FileServer(http.FS(b.staticFS))
	if b.isDevEnv {
		handler = http.FileServer(http.Dir(b.StaticRootDir))
	}

	handler = http.StripPrefix(route, handler)
	return func(w http.ResponseWriter, r *http.Request) {
		if b.isDevEnv {
			w.Header().Set("Cache-Control", "no-cache")
		}
		handler.ServeHTTP(w, r)
	}
}

func (b Bakery) MakeScriptHandler(checkinEndpoint string, port int) (http.HandlerFunc, error) {

	t, err := template.ParseFS(sseConsumer, "sse_consumer.go.html")
	if err != nil {
		return nil, err
	}

	data := ReloaderData{
		checkInEndpoint: checkinEndpoint,
	}

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/javascript")
		t.Execute(w, data)
	}, nil
}

func (b Bakery) MakeSSEHandler() (http.HandlerFunc, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	localFS := os.DirFS(pwd)

	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		toSend := "data: {\"connected\": true}\n\n"
		fmt.Fprint(w, toSend)
		w.(http.Flusher).Flush()

		reloaderChan := make(chan bool)
		go b.watch(r.Context(), localFS, reloaderChan)

		for {
			select {
			case <-r.Context().Done():
				return
			case <-reloaderChan:
				toSend := "data: {\"reload\": true}\n\n"
				fmt.Fprint(w, toSend)
				w.(http.Flusher).Flush()
			}
		}
	}, nil
}

type ReloaderData struct {
	checkInEndpoint string
}

func (rd ReloaderData) GetEndpoint() string {
	return rd.checkInEndpoint
}
func (rd *ReloaderData) SetEndpoint(c string) {
	rd.checkInEndpoint = c
}

func (rd ReloaderData) HasEndpoint() bool {
	return len(rd.checkInEndpoint) > 0
}
