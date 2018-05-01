package chix

import (
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi"
)

// FileServer conveniently sets up a http.FileServer r to serve static files
// from a http.FileSystem.
// @see https://github.com/go-chi/chi/issues/155
// @see https://github.com/go-chi/chi/issues/184
func FileServer(r *chi.Mux, pattern string, path string) {
	if strings.ContainsAny(pattern, "{}*") {
		panic("FileServer does not permit URL parameters.")
	}

	root := http.Dir(path)
	fs := http.StripPrefix(pattern, http.FileServer(root))

	if pattern != "/" && pattern[len(pattern)-1] != '/' {
		r.Get(pattern, http.RedirectHandler(pattern+"/", http.StatusMovedPermanently).ServeHTTP)
		pattern += "/"
	}

	r.Get(pattern+"*", http.HandlerFunc(func(w http.ResponseWriter, rq *http.Request) {
		if _, err := os.Stat(path + rq.RequestURI); os.IsNotExist(err) {
			r.NotFoundHandler().ServeHTTP(w, rq)
		} else {
			fs.ServeHTTP(w, rq)
		}
	}))
}
