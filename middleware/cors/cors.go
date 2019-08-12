package cors

import "net/http"

func AddCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, PATCH, DELETE")
		w.Header().Add("Access-Control-Allow-Headers",
			"Accept, X-Token, X-Compress, Content-Type, Content-Length, Accept-Encoding")

		if r.Method != "OPTIONS" {
			next.ServeHTTP(w, r)
		}
	})
}
