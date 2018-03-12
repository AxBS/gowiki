package main

import (
	"errors"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
)

type Page struct {
	Title string
	Body  []byte
}

func (p *Page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err //Returning nil instead of the missing page, and the error value
	}
	return &Page{Title: title, Body: body}, nil //Returing the page and the error value(nil in this case)
}

//
// func handler(w http.ResponseWriter, r *http.Request) {
// 	//An http.Request is a data structure that represents the client HTTP request. r.URL.Path is the path component
// 	// of the request URL. The trailing [1:] means "create a sub-slice Path from de 1st charactes ot he end", droppinf the leading "/" from the path name.
// 	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
// }

//Handler that allows users to view a wiki page. It will handle URLs prefixed with "/view/"
func viewHandler(w http.ResponseWriter, r *http.Request, title string) {

	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

/*
We can make out code more efficient  by rendering both templates once at program initialization, parsing all templates into a single *Template.
Then we can use the ExecuteTemplate method to render a specific template

The function template.Must is a convenience wrapper that pancs when passed a non-nil error value, and otherwise return the *Template unaltered. A panic
is appropriate here; if the templates can't be loaded the only sensible thing to do is exit the program
*/
var templates = template.Must(template.ParseFiles("edit.html", "view.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {

	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

/*
The page title (provided in the URL) and the form's only field, Body, are stored in a new Page.
The save() method is then called to write the data to a file, and the client is redirected to the /view/ page.
*/
func saveHandler(w http.ResponseWriter, r *http.Request, title string) {

	body := r.FormValue("body")
	/*
		The value returned by FormValue is of type string.
		We must convert that value to []byte before it will fit into the Page struct. We use []byte(body) to perform the conversion.
	*/
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {

	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Here we will extract the page title from Request,
		// and call the provided handler 'fn'
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

/*
In order to mitigate the security flaw that we have when we let the user  use an arbitrary path to be read/written on the server (by the user).
We can write a function to validate the title with a regular expression:
The function regexp.MustCompite will parse and compile the regular expression, and return a regexp.Regexp.
MustCompule is distinct from Compile in that it will panic i the expression compilation fails, while Compile returns an error as a second parameter.


If the title is valid, it will be returned along with a nil error value. If the title is invalid, the function will write a "404 Not Found" error to the HTTP
conection, and return an error to the handler.
*/
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

//Now we will write a function that uses validPath to validate path and extract the page title:
func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return "", errors.New("Invalid Page Title")
	}
	return m[2], nil // The title is the second subexpression
}

func main() {

	//This is a call to http.HandleFunc, which tells the http package to handle all requests to the
	// web root ("/") with handler
	//http.HandleFunc("/", handler)

	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	//http.HandleFunc("/", handler)

	//Then it calls http.ListenAndServe, specifying that it should listen on port 8080 on any interface (":8080").
	// LIstenAndServe ALWAYS returns an ERROR, since it only returns when an unexpected error occurs. In order to log that error,
	// we wrap the function call with log.Fatal
	log.Fatal(http.ListenAndServe(":8080", nil))

	//The function handler is of the type http.HandleFunc. It takes an http.ResponseWriter and an http.Request as its arguments
	//An http.ResponseWriter value assembles the HTTP server's response; by writing to it, we send data to the HTTP CLIENT
}
