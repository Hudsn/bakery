package bakery

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"os"
)

//go:embed sse_consumer.go.html
var sseConsumer embed.FS

//go:embed static_page_wrapper.go.html
var iframeWrapper embed.FS

func MakeStaticHandler(route string, rootDir string, targetFS fs.FS, isDevEnv bool) http.HandlerFunc {
	handler := http.FileServer(http.FS(targetFS))
	if isDevEnv {
		handler = http.FileServer(http.Dir(rootDir))
	}

	handler = http.StripPrefix(route, handler)
	return func(w http.ResponseWriter, r *http.Request) {
		if isDevEnv {
			w.Header().Set("Cache-Control", "no-cache")
		}
		handler.ServeHTTP(w, r)
	}
}

// creates an html page embedding the page at the target route within an iframe so we can enable hot reloading.
// for dev use only; mostly useful for editing static pages that don't use templating (generic error pages / content pages) .
func MakeIframeWatcher(watchHtmlRoute string, reloadScriptPath string) (http.HandlerFunc, error) {

	type IframeData struct {
		TargetRoute      string
		ReloadScriptPath string
	}

	t, err := template.ParseFS(iframeWrapper, "static_page_wrapper.go.html")
	if err != nil {
		return nil, err
	}

	data := IframeData{
		TargetRoute:      watchHtmlRoute,
		ReloadScriptPath: reloadScriptPath,
	}

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("Cache-Control", "no-cache")
		err := t.Execute(w, data)
		if err != nil {
			fmt.Println(err)
		}
	}, nil
}

func MakeSingleFileHandler(filePath string, embedBytes []byte, contentType string, errHandler http.HandlerFunc, isDevEnv bool) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		var resBuffer *bytes.Buffer
		if isDevEnv {
			w.Header().Set("Cache-Control", "no-cache")
			b, err := os.ReadFile(filePath)
			if err != nil {
				errHandler.ServeHTTP(w, r)
				return
			}
			resBuffer = bytes.NewBuffer(b)
		} else {
			resBuffer = bytes.NewBuffer(embedBytes)
		}
		w.Header().Set("Content-Type", contentType)
		resBuffer.WriteTo(w)
	}
}

func MakeScriptHandler(checkinEndpoint string, port int) (http.HandlerFunc, error) {

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

func MakeSSEHandler(b Bakery) (http.HandlerFunc, error) {
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
