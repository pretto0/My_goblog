package middlewares

import "net/http"



func ForceHTML(h http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "text/html; charset=utf-8")
        h.ServeHTTP(w, r)
    })
}
