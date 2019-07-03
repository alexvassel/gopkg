package transport

import (
	"github.com/severgroup-tt/gopkg-errors/transport"

	"github.com/utrack/clay/v2/transport/httpruntime"
)

// Override
func Override() {
	OverrideMarshaler()
	OverrideErrorRenderer()
}

// OverrideMarshaler
func OverrideMarshaler() {
	jsonMarshaller := NewMarshaller()
	httpruntime.OverrideMarshaler(jsonMarshaller.ContentType(), jsonMarshaller)
}

// OverrideErrorRenderer
func OverrideErrorRenderer() {
	httpruntime.SetError = transport.ErrorRenderer
	httpruntime.TransformUnmarshalerError = TransformUnmarshalerError
}
