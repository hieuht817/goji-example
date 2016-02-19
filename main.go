package main

import (
	"fmt"
	"net/http"

	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
	"regexp"
	"github.com/zenazn/goji/web/middleware"
	"io"
	"github.com/goji/param"
	"strconv"
	"time"
)

func hello(c web.C, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %s!", c.URLParams["name"])
}

func main() {
	goji.Get("/", Root)
	goji.Get("/greets", http.RedirectHandler("/", 301))
	goji.Post("/greets", NewGreet)
	goji.Get("/users/:name", GetUser)
	goji.Get(regexp.MustCompile(`^/greets/(?P<id>\d+)$`), GetGreet)

	// Middleware
	goji.Use(PlainText)

	// prefix path
	admin := web.New()
	goji.Handle("/admin/*", admin)

	admin.Use(middleware.SubRouter)

	admin.Use(SuperSecure)

	admin.Get("/", AdminRoot)
	admin.Get("/finances", AdminFinances)


	goji.Get("/admin", http.RedirectHandler("/admin/", 301))
	// custom 404 handler
	goji.NotFound(NotFound)

	// some long request
	goji.Get("/waitforit", WaitForIt)

	// Serve():
	// binding to a socket (auto support for systemd and Einhorn
	// support graceful shutdown on SIGINT
	// for both development and production
	goji.Serve()
}

// Root route (GET "/"). Print a list of greets.
func Root(w http.ResponseWriter, r *http.Request) {
	// can use template
	io.WriteString(w, "Gritter\n======\n\n")
	for i := len(Greets) - 1; i >= 0; i-- {
		Greets[i].Write(w)
	}
}

// NewGreet creates a new greet (POST "/greets"). Creates a greet and redirects
// to the created greet.
//
// To post a new greet, try this at a shell:
// $ now=$(date +'%Y-%m-%dT%H:%M:%SZ')
// $ curl -i -d "user=carl&message=Hello+World&time=$now" localhost:8000/greets
func NewGreet(w http.ResponseWriter, r *http.Request) {
	var greet Greet

	r.ParseForm()
	err := param.Parse(r.Form, &greet)

	if err != nil || len(greet.Message) > 140 {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	Greets = append(Greets, greet)
	url := fmt.Sprintf("/greets/%d", len(Greets)-1)
	http.Redirect(w, r, url, http.StatusCreated)
}

// GetUser finds a given user and her greets (GET "/user/:name")
func GetUser(c web.C, w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Gritter\n======\n\n")
	handle := c.URLParams["name"]
	user, ok := Users[handle]
	if !ok {
		http.Error(w, http.StatusText(404), 404)
		return
	}

	user.Write(w, handle)
	io.WriteString(w, "\nGreets:\n")
	for i := len(Greets) - 1; i >=0; i-- {
		if Greets[i].User == handle {
			Greets[i].Write(w)
		}
	}
}

// GetGreet finds a particular greet by ID (GET "/greets/\d+"). Does no bounds
// checking, so will probably panic.
func GetGreet(c web.C, w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(c.URLParams["id"])
	if err != nil {
		http.Error(w, http.StatusText(404), 404)
		return
	}
	greet := Greets[id]

	io.WriteString(w, "Gritter\n======\n\n")
	greet.Write(w)
}

// WaitForIt is a particularly slow handler (GET "/waitforit"). Try loading this
// endpoint and initiating a graceful shutdown (Ctrl-C) or Einhorn reload. The
// old server will stop accepting new connections and will attempt to kill
// outstanding idle (keep-alive) connections, but will patiently stick around
// for this endpoint to finish. How kind of it!
func WaitForIt(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "This is going to be legend... (wait for it)\n")
	if fl, ok := w.(http.Flusher); ok {
		fl.Flush()
	}
	time.Sleep(15 * time.Second)
	io.WriteString(w, "...dary! Legendary!\n")
}

// AdminRoot is root (GET "/admin/root"). Much secrete. Very administrate. Wow.
func AdminRoot(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Gritter\n======\n\nSuper secrete admin page!\n")
}

// AdminFinances would answer the question 'How are you doing?'
// (GET "/admin/finances")
func AdminFinances(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Gritter\n======\n\nWe're broke! :(\n")
}

// NotFound is a 404 handler.
func NotFound(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Umm... have you tried turning if off and on again?", 404)
}