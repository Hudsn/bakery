# Bakery

Fresh baked `html/template` views delivered fresh to your browser. No other 3rd party dependencies required.

Makes working with Go's `html/template` package easier by helping to reduce setup boilerplate, providing helpers for chaining and rendering templates, and enabling live browser reload in development environments.

## Why though?

As someone interested in building monolithic web applications in Go (HTMX hype train anyone?), I found myself struggling with initial setup and project structure of my Go HTML templates. 

Even when I got an initial setup working I still felt annoyed by the sluggish feedback loop of having to re-run `go run ...` and then refresh the browser just to see how a one line change affected the appearance of the page. Auto-reload tools like `air` and `modd` exist and are fantastic but you shouldn't have to refresh an application for updating static files.

This package is just a version of how I set up my own implementation + a splash of hot browser reloading without the need to restart your program every time.

## Usage Example

A comprehensive example of rendering a page with nested templates can be found in the `example` directory.

A detailed walkthrough of this implementation is included in the next section but if you just want to try it out, you can simply `go run ./example -dev` to play around with the dev environment implementation with hot reload.

Alternatively, remove the `-dev` flag to disable hot reload and only use embedded filesystems for serving cached templates and static files.

## An overly verbose walkthrough

### NOTE / TLDR: 

You absolutely don't have to follow this entire walkthrough if you can read and surmise most of what is being done in the `example/views/views.go`, and `example/main.go` files. However if you happen to want an over-explanation of the example implementation, read on...

### Your file structure
You'll need to choose and eventually specify root directory paths for: 

1. Your template files
2. Your static files (images, css, js, etc...) 

These directories should contain all the files for your templates and static files, respectively.

In our example we use `example/views` for our templates, and `example/static` for static files. Note that these should be specified relative to the root of your project.

### Embed your files

We declare the static file embedding for our `css` and `js` files at `example/static`:

```go
// example/static/static.go

import "embed"

//go:embed style/*.css
//go:embed script/*.js
var StaticFS embed.FS
```

Note that you can implement this embedding any way you want. You just need to ensure that you know how to access them by path. For example, if we wanted to access our `style.css` file, we need to do so from the path `style/style.css`. 

For another example if we hypothetically declared our style embed one level up from that folder with `go:embed static/style/*.css` we'd then access it with `static/style/style.css`. Any approach we want to take is fine, just be congizant of how you'd access it for later.

We declare the template file embedding for our files at `example/views`:

```go
//go:embed layouts/*go.html
//go:embed components/*.go.html
//go:embed components/*/*.go.html
//go:embed pages/*.go.html
var viewFS embed.FS
```

The same guidance for how you embed files applies to templates as well.

### An example config: 

With our decisions around template and static directories out of the way, let's look at what our configuration struct might be. This is used to initialize our `Bakery` struct which is the engine for our rendering and hot reload functionality:

```go
// example/views/views.go
bakery.Config{
    IsDev: true, // whether or not the environment is a dev environment
    WatchExtensions: []string{ 
        ".go.html",
        ".css",
        ".js",
    }, // which extensions we want to watch so that we can trigger browser reload
    TemplateFS:      viewFS, // our template FS that we declared above
    TemplateRootDir: "example/views", // the path to our template directory (relative to our project root), that we decided above
    StaticFS:        static.StaticFS, // our static FS that we declared above
    StaticRootDir:   "example/static",  // the path to our static directory (relative to our project root), that we decided above
}
```

### Initializing our Bakery 

The `bakery.New` function accepts a configuration like the one above as its input, and returns a new `bakery.Bakery` struct for rendering and browser reloading.

Let's write a quick function in our `example/views/views.go` file to generate a slightly different `Bakery` struct depending on whether we're in a development environment:

```go
//example/views/views.go

func New(isDev bool) bakery.Bakery {

	config := bakery.Config{
		IsDev: isDev,
		WatchExtensions: []string{
			".go.html",
			".css",
			".js",
		},
		TemplateFS: viewFS,
		TemplateRootDir: "example/views",
		StaticFS:        static.StaticFS,
		StaticRootDir:   "example/static",
	}

	myBakery := bakery.New(config)

    myBakery.Init()

	return myBakery
}
```

### Adding a template 'recipe'

Now that we have our `Bakery` engine struct, we can add template chains to be rendered using `AddRecipe`. In our example we want to use the following template chain: `layouts/app.go.html` > `home.go.html` > `components/button.go.html` > `components/sub/sub.go.html`

We'll modify our own existing `New(isDev bool)` function from the previous section to enable this:

```go
//example/views/views.go

func New(isDev bool) bakery.Bakery {

	//...

    //Add this line to the existing function
	myBakery.AddRecipe("home", "layouts/app.go.html", "pages/home.go.html", "components/button.go.html", "components/sub/sub.go.html")


    // everything below this comment was already in our function
    myBakery.Init() 

	return myBakery
}
```

Now we can simply call `myBakery.Bake("home", ourDataHere)` and we'll get back an `http.HandlerFunc` that renders the page for us! More on that later.

### Avoiding repetition (optional)

If we have template files in the beginning of our template chain that are common, like a hypothetical chain `mylayout.go.html` > `header.go.html` > `footer.go.html` and you don't want to rewrite them every single time you use `AddRecipe`, you can use `AddRecipeUsingTemplateList(name string, existingList []string, additionalPaths ...string)` instead.

It would look something like this:

```go
//example/views/views.go
var appTemplateChain := []string{
    "mylayout.go.html",
    "header.go.html",
    "footer.go.html",
}

func NewView() bakery.Bakery {} {
    // all our setup code we already covered

    //add an imaginary "about page"
    myBakery.AddRecipeUsingTemplateList("about", appTemplateChain, "pages/about.go.html")

    // init and return below
}
```
### Making our handlers

Now that we have our Bakery constructor and our `home` page ready to be rendered, let's use it in our `main.go` file.

First we get some input from our flags, and pass our `isDev` bool to our bakery constructor function that we already defined above:

```go
//example/main.go

isDev := flag.Bool("dev", false, "whether the environment is dev or not")
port := flag.Int("port", 9001, "set port to listen on")
flag.Parse()

bakery := views.New(*isDev) // we registered all our template paths and in the Views package, then call it here. 

// continued below
```

Next, we'll create our handlers for hot reloading and serving static files:

```go
//example/main.go

//...

// this will be our handler for hosting the js file that listens for updates and triggers browser reload. 
// *NOTE* the argument you pass here should be the route you assign to your "SSE Handler". in this case we've chosen "/tepid". 
scriptHandler, err := bakery.MakeScriptHandler("/tepid", *port)
if err != nil {
    log.Fatal(err)
}

//this is the handler that sends triggers for the browser to reload.
sseHandler, err := bakery.MakeSSEHandler()
if err != nil {
    log.Fatal(err)
}

// this is the handler that will serve as the root for our static files.
// the argument should be the URL path where you wish to serve static files from.
// *NOTE* this handler already strips route prefixes.
staticHandler := bakery.MakeStaticHandler("/static/")
```

Finally, we'll create our router and wire up our generated handlers: 

```go
//...
router := http.NewServeMux()

//we only want to make our reloader endpoints available if we're in a dev environment.
if *isDev {
    router.HandleFunc("/tepid.js", scriptHandler)
    router.HandleFunc("/tepid", sseHandler)
}
router.HandleFunc("/static/", staticHandler)

fmt.Printf("Listening on :%d...\n", *port)
log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), router))

```
Note that the `/tepid` route for our `sseHandler` is the same route that we passed to `MakeScriptHandler` above, and that our `/static/` route is the same as the argument we passed to `MakeStaticHandler`.

Also make aa mental note of the route we're assigning to our `scriptHandler`, which is `/tepid.js`, since we'll need it in our next section...

### Adding our reloader to our templates

Since we want our reloader JS file to be active on all of our pages for our layout, we'll need to add it to our parent layout template:

```handlebars
<!-- example/views/layouts/app.go.html -->

<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.GetTitle}}</title>
    <link rel="stylesheet" href="/static/style/style.css"/>
</head>
<body>
    <h1>My app!</h1>
    {{template "page" .}}

    <!-- THIS PART IS WHAT WE MOSTLY CARE ABOUT FOR THE EXAMPLE -->
    {{if .GetIsDev}}
    <script defer="true" src="/tepid.js"></script>
    {{end}}
</body>
</html>
```

We only render the script tag for our `/tepid.js` reloader listener if we're in a dev environment.

Additionally, since our `home` page will be injected into this layout in the `{{template "page" .}}` slot, it will automatically have the reloader script enabled too. 

Side note - if you're wondering where these "getter" functions like `.GetIsDev` are coming from, we've defined them in `example/views/layouts/`. But note that you can define your own data to pass to templates however you want:

```go
//example/vews/layouts/app.go

type AppLayoutData struct {
	title string
	isDev bool
}

func (a *AppLayoutData) SetTitle(t string) {
	a.title = t
}

func (a AppLayoutData) GetTitle() string {
	return a.title
}

func (a *AppLayoutData) SetIsDev(b bool) {
	a.isDev = b
}

func (a AppLayoutData) GetIsDev() bool {
	return a.isDev
}
```

### Serving the templates

Finally, since that we understand generally where our reloader script lives and how it gets put into our template, let's quickly make the handler for our `home` page, and add it to our router:

First let's make the handler:

```go
//example/handlers/handlers.go

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
```

The `DefaultAppData` helper is just a way for us to get some sane default values for the data we'll pass to our template via our template Bakery's `Bake` function we talked about earlier.

All we really did here in our `HomeHandler` is set the page title to `Home` and then let our `Bake` function take care of the rest.

With our `HomeHandler` functional, let's give it a path in our router and call it done. Here's what our `main.go` looks like at the end:

```go
//example/main.go

func main() {
	isDev := flag.Bool("dev", false, "whether the environment is dev or not")
	port := flag.Int("port", 9001, "set port to listen on")
	flag.Parse()

	bakery := views.New(*isDev) 
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
	router.HandleFunc("/static/", staticHandler)

    //HERE - we added our home handler that we defined above.
	router.HandleFunc("/home", handlers.HomeHandler(bakery))

	fmt.Printf("Listening on :%d...\n", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), router))
}
```
