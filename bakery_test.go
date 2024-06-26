package bakery_test

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/hudsn/bakery"
	testtemplate "github.com/hudsn/bakery/test_files/templates"
)

//go:embed test_files/templates/*.go.html
var testTemplates embed.FS

//go:embed test_files/static/*.txt
var testStatic embed.FS

//go:embed test_files/static2/page.html
var testPageByte []byte

func TestBakery(t *testing.T) {

	tFS, err := fs.Sub(testTemplates, "test_files/templates")
	if err != nil {
		t.Fatal(err)
	}

	_, err = tFS.Open("home.go.html")
	if err != nil {
		t.Log(err)
	}

	sFS, err := fs.Sub(testStatic, "test_files/static")
	if err != nil {
		t.Fatal(err)
	}
	myConfig := bakery.Config{
		IsDev:           false,
		WatchExtensions: []string{".go.html", ".txt"},
		TemplateRootDir: "test_files/templates",
		TemplateFS:      tFS,
	}

	myBakery := bakery.New(myConfig)

	myBakery.AddRecipe("home", "home.go.html")
	myBakery.Init()

	t.Run("check static file", func(t *testing.T) {
		b, err := os.ReadFile("test_files/static/test.txt")
		if err != nil {
			t.Fatal(err)
		}
		wantText := string(b)

		myStaticHandler := bakery.MakeStaticHandler("", "test_files/static", sFS, false)
		testServ := httptest.NewServer(myStaticHandler)
		defer testServ.Close()

		res, err := http.Get(testServ.URL + "/test.txt")
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		if res.StatusCode != 200 {
			t.Errorf("expected status OK, instead got status code %d", res.StatusCode)
			return
		}

		resBytes, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}

		if string(resBytes) != wantText {
			t.Errorf("expected file contents %q, instead got %q", wantText, string(resBytes))
		}

	})
	t.Run("check template render", func(t *testing.T) {

		templateData := testtemplate.HomeData{
			Title: "MY TITLE",
		}

		wantText1 := "<h1>HELLO WORLD!</h1>"
		wantText2 := fmt.Sprintf("<title>%s</title>", templateData.Title)

		testServ := httptest.NewServer(myBakery.Bake("home", templateData))
		defer testServ.Close()

		res, err := http.Get(testServ.URL)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		if res.StatusCode != 200 {
			t.Errorf("expected status OK, instead got status code %d", res.StatusCode)
			return
		}

		resBytes, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}

		if !strings.Contains(string(resBytes), wantText1) {
			t.Errorf("expected response to contain %q,\ninstead got:\n%q", wantText1, string(resBytes))

		}
		if !strings.Contains(string(resBytes), wantText2) {
			t.Errorf("expected response to contain %q,\ninstead got:\n%q", wantText1, string(resBytes))

		}

	})
	t.Run("check individual static page render", func(t *testing.T) {

		wantText1 := "<h1>HELLO WORLD 2!</h1>"

		errHandler := func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		testSingleFileHandler := bakery.MakeSingleFileHandler("test_files/static2/page.html", testPageByte, "text/html", errHandler, false)

		testServ := httptest.NewServer(testSingleFileHandler)
		defer testServ.Close()

		res, err := http.Get(testServ.URL)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		if res.StatusCode != 200 {
			t.Errorf("expected status OK, instead got status code %d", res.StatusCode)
			return
		}

		resBytes, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}

		if !strings.Contains(string(resBytes), wantText1) {
			t.Errorf("expected response to contain %q,\ninstead got:\n%q", wantText1, string(resBytes))
		}

	})
	t.Run("check static page iframe mirroring", func(t *testing.T) {

		wantText1 := `<script defer="true" src="/reloader.js"></script>`
		wantText2 := `<iframe width="100%" height="100%" frameborder="0" src="/imaginary"></iframe>`

		iframeHandler, err := bakery.MakeIframeWatcher("/imaginary", "/reloader.js")
		if err != nil {
			t.Fatal(err)
		}

		testServ := httptest.NewServer(iframeHandler)
		defer testServ.Close()

		res, err := http.Get(testServ.URL)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		if res.StatusCode != 200 {
			t.Errorf("expected status OK, instead got status code %d", res.StatusCode)
			return
		}

		resBytes, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}

		if !strings.Contains(string(resBytes), wantText1) {
			t.Errorf("expected response to contain %q,\ninstead got:\n%q", wantText1, string(resBytes))
		}

		if !strings.Contains(string(resBytes), wantText2) {
			t.Errorf("expected response to contain %q,\ninstead got:\n%q", wantText2, string(resBytes))
		}

	})

}
