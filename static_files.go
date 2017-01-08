package main

import (
	"fmt"
	"net/http"
	"path/filepath"
)

// DefaultHtmlis a middleware serving injecting content type HTML if file has no extension
func DefaultHtml(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		_, haveType := w.Header()["Content-Type"]
		// inject content-type only if is not set
		if !haveType {
			fileext := filepath.Ext(r.URL.Path)
			if "" == fileext {
				//w.Header().Set("Content-Type", "text/html; charset=utf-8")
			}
		}
		h.ServeHTTP(w, r)
		fmt.Printf("XDEBUG:DefaultHtml\n")
	}

	return http.HandlerFunc(fn)
}
