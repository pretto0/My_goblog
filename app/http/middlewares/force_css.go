package middlewares

import "net/http"

func ForceCSS(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
        next.ServeHTTP(w, r)
    })
}