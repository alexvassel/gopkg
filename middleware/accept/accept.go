package accept

import (
	"net/http"
	"strings"

	"github.com/go-openapi/runtime/middleware/header"
)

func AddAccept(offer string, excludePaths []string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, path := range excludePaths {
				if strings.Contains(r.URL.Path, path) {
					next.ServeHTTP(w, r)
					return
				}
			}
			specs := header.ParseAccept(r.Header, "Accept")
			for _, spec := range specs {
				if spec.Value == offer {
					next.ServeHTTP(w, r)
					return
				}
			}
			http.Error(w, "Not allowed Content-Type", http.StatusNotAcceptable)
		})
	}
}
