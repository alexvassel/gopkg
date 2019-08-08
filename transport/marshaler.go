package transport

import (
	"io"

	gogojsonpb "github.com/gogo/protobuf/jsonpb"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/utrack/clay/v2/transport/httpruntime"
)

const contentType = "application/json"

// NewMarshaller ...
func NewMarshaller(options MarshallerOptions) httpruntime.Marshaler {
	return marshaller{
		base: httpruntime.MarshalerPbJSON{
			Marshaler:       &runtime.JSONPb{OrigName: true, EmitDefaults: options.emitDefaults, EnumsAsInts: false},
			Unmarshaler:     &runtime.JSONPb{OrigName: true, EmitDefaults: options.emitDefaults, EnumsAsInts: false},
			GogoMarshaler:   &gogojsonpb.Marshaler{OrigName: true, EmitDefaults: options.emitDefaults, EnumsAsInts: false},
			GogoUnmarshaler: &gogojsonpb.Unmarshaler{AllowUnknownFields: true},
		},
	}
}

func NewMarshallerOptions() MarshallerOptions {
	return MarshallerOptions{
		emitDefaults: true,
	}
}

func (m MarshallerOptions) WithEmitDefaults(val bool) {
	m.emitDefaults = val
}

type MarshallerOptions struct {
	emitDefaults bool
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
