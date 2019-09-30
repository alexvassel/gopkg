package middleware

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/go-chi/chi/middleware"
	"net/http"
	"os"
	"strings"
)

var prefix string

func SetRequestId(ctx context.Context) context.Context {
	id := middleware.NextRequestID()
	return context.WithValue(ctx, middleware.RequestIDKey, fmt.Sprintf("%s-%06d", prefix, id))
}

func GetRequestId(ctx context.Context) string {
	return middleware.GetReqID(ctx)
}

func NewRequestIdMiddleware() func(next http.Handler) http.Handler {
	return middleware.RequestID
}

func init() {
	hostname, err := os.Hostname()
	if hostname == "" || err != nil {
		hostname = "localhost"
	}
	var buf [12]byte
	var b64 string
	for len(b64) < 10 {
		_, _ = rand.Read(buf[:])
		b64 = base64.StdEncoding.EncodeToString(buf[:])
		b64 = strings.NewReplacer("+", "", "/", "").Replace(b64)
	}

	prefix = fmt.Sprintf("%s/%s", hostname, b64[0:10])
}
