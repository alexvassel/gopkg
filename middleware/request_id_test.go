package middleware

import (
	"net/http"
	"testing"
)

func TestSetId(t *testing.T) {
	t.Run("SetUuid", func(t *testing.T) {
		var r = &http.Request{}
		var x = r.WithContext(SetUuid(r.Context()))
		y := GetId(x.Context())
		if x.Context().Value(uuidKey) != y {
			t.Error("uuid Keys do not match")
		}
	})
}
