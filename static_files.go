package main

import (
	"fmt"
	"github.com/pressly/chi"
	"net/http"
	"net/url"
	"os"
	"path"
	_ "path/filepath"
	"sort"
	"strings"
)

type fileHandler struct {
	root http.FileSystem
}

type byName []os.FileInfo

func (s byName) Len() int           { return len(s) }
func (s byName) Less(i, j int) bool { return s[i].Name() < s[j].Name() }
func (s byName) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

var htmlReplacer = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
	// "&#34;" is shorter than "&quot;".
	`"`, "&#34;",
	// "&#39;" is shorter than "&apos;" and apos was not in HTML until HTML5.
	"'", "&#39;",
)

// toHTTPError returns a non-specific HTTP error message and status code
// for a given non-nil error value. It's important that toHTTPError does not
// actually return err.Error(), since msg and httpStatus are returned to users,
// and historically Go's ServeContent always returned just "404 Not Found" for
// all errors. We don't want to start leaking information in error messages.
func toHTTPError(err error) (msg string, httpStatus int) {
	if os.IsNotExist(err) {
		return "404 page not found", http.StatusNotFound
	}
	if os.IsPermission(err) {
		return "403 Forbidden", http.StatusForbidden
	}
	// Default:
	return "500 Internal Server Error", http.StatusInternalServerError
}

// localRedirect gives a Moved Permanently response.
// It does not convert relative paths to absolute paths like Redirect does.
func localRedirect(w http.ResponseWriter, r *http.Request, newPath string) {
	if q := r.URL.RawQuery; q != "" {
		newPath += "?" + q
	}
	w.Header().Set("Location", newPath)
	w.WriteHeader(http.StatusMovedPermanently)
}

func dirList(base string, w http.ResponseWriter, f http.File) {
	dirs, err := f.Readdir(-1)
	if err != nil {
		// TODO: log err.Error() to the Server.ErrorLog, once it's possible
		// for a handler to get at its Server via the ResponseWriter. See
		// Issue 12438.
		http.Error(w, "Error reading directory", http.StatusInternalServerError)
		return
	}
	sort.Sort(byName(dirs))

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, "<pre>\n")
	for _, d := range dirs {
		name := d.Name()
		if d.IsDir() {
			name += "/"
		}
		// name may contain '?' or '#', which must be escaped to remain
		// part of the URL path, and not indicate the start of a query
		// string or fragment.
		url := url.URL{Path: base + "/" + name}
		fmt.Fprintf(w, "<a href=\"%s\">%s</a>\n", url.String(), htmlReplacer.Replace(name))
	}
	fmt.Fprintf(w, "</pre>\n")
}

// name is '/'-separated, not filepath.Separator.
func serveFile(w http.ResponseWriter, r *http.Request, fs http.FileSystem, name string, redirect bool) {
	indexPage := "/index"
	fmt.Println("Serving ", name)

	f, err := fs.Open(name)
	if err != nil {
		msg, code := toHTTPError(err)
		http.Error(w, msg, code)
		return
	}
	defer f.Close()

	d, err := f.Stat()
	if err != nil {
		msg, code := toHTTPError(err)
		http.Error(w, msg, code)
		return
	}

	// normalize direcoty url : remove ending '/' of directories, they are pages
	n := len(name)
	if n > 1 && name[n-1:] == "/" {
		fmt.Println("redirect ", name, " to ", name[0:n-1])
		localRedirect(w, r, name[0:n-1])
		return
	}

	// display directory content : indexPage
	if d.IsDir() {
		// try to serve index page
		index := strings.TrimSuffix(name, "/") + indexPage
		ff, err := fs.Open(index)
		if err == nil {
			defer ff.Close()
			dd, err := ff.Stat()
			if err == nil {
				name = index
				d = dd
				f = ff
			}
		}
	}

	// Still a directory? (we didn't find an indexpage file)
	if d.IsDir() {
		dirList(name, w, f)
		return
	}

	// serveContent will check modification time
	http.ServeContent(w, r, d.Name(), d.ModTime(), f)
}

func (f *fileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	upath := r.URL.Path
	if !strings.HasPrefix(upath, "/") {
		upath = "/" + upath
		r.URL.Path = upath
	}

	// cleanup path early
	nPath := path.Clean(upath)
	if nPath != upath {
		localRedirect(w, r, nPath)
		return
	}
	serveFile(w, r, f.root, path.Clean(upath), true)
}

// FileServer returns a handler that serves HTTP requests
// with the contents of the file system rooted at root.
//
// To use the operating system's file system implementation,
// use http.Dir:
//
//     http.Handle("/", http.FileServer(http.Dir("/tmp")))
//
// As a special case, the returned file server redirects any request
// ending in "/index.html" to the same path, without the final
// "index.html".
func FileServer(root http.FileSystem) http.Handler {
	return &fileHandler{root}
}

// StaticServer conveniently sets up a customized version of http.FileServer handler to serve
// static files from a http.FileSystem.
func StaticServer(mx *chi.Mux, path string, root http.FileSystem) {
	if strings.ContainsAny(path, ":*") {
		panic("chi: FileServer does not permit URL parameters.")
	}

	mx.Group(func(r chi.Router) {
		//r.Use(NormalizePath)

		fs := http.StripPrefix(path, FileServer(root))
		r.Get(path+"*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fs.ServeHTTP(w, r)
		}))
	})

}
