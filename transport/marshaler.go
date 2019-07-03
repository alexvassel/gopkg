package transport

import (
	"io"

	gogojsonpb "github.com/gogo/protobuf/jsonpb"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/utrack/clay/v2/transport/httpruntime"
)

const contentType = "application/json"

// NewMarshaller ...
func NewMarshaller() httpruntime.Marshaler {
	return marshaller{
		base: httpruntime.MarshalerPbJSON{
			Marshaler:       &runtime.JSONPb{OrigName: true, EmitDefaults: true, EnumsAsInts: false},
			Unmarshaler:     &runtime.JSONPb{OrigName: true, EmitDefaults: true, EnumsAsInts: false},
			GogoMarshaler:   &gogojsonpb.Marshaler{OrigName: true, EmitDefaults: true, EnumsAsInts: false},
			GogoUnmarshaler: &gogojsonpb.Unmarshaler{AllowUnknownFields: true},
		},
	}
}

type marshaller struct {
	base httpruntime.Marshaler
}

// ContentType ...
func (m marshaller) ContentType() string {
	return contentType
}

// Marshal ...
func (m marshaller) Marshal(w io.Writer, response interface{}) error {
	return m.base.Marshal(w, response)
}

// Unmarshal ...
func (m marshaller) Unmarshal(r io.Reader, dst interface{}) error {
	return m.base.Unmarshal(r, dst)
}
