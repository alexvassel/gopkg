package request_id

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
)

const ContextName = "request-id"

func SetUuid(ctx context.Context) context.Context {
	b := make([]byte, 16)
	rand.Read(b)

	uuid := fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])

	return context.WithValue(ctx, ContextName, uuid)
}

func GetId(ctx context.Context) string {
	uuid, _ := ctx.Value(ContextName).(string)
	return uuid
}

func SetId(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(SetUuid(r.Context()))
		w.Header().Add("X-Request-Id", GetId(r.Context()))
		next.ServeHTTP(w, r)
	})
}
