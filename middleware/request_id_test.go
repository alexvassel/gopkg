package middleware

import (
	"net/http"
	"testing"
)

func TestSetId(t *testing.T) {
	t.Run("SetUuid", func(t *testing.T) {
		var r = &http.Request{}
		var x = r.WithContext(SetRequestUuid(r.Context()))
		y := GetRequestId(x.Context())
		if x.Context().Value(requestIDContextName) != y {
			t.Error("uuid Keys do not match")
		}
	})
}
