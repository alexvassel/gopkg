package provider

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPrepareMessage(t *testing.T) {
	t.Run("Plain to HTML", func(t *testing.T) {
		msg := &Message{
			bodyPlain: `Hello!
Second, string

Bye`,
		}
		err := msg.Prepare(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, "Hello!<br/>Second, string<br/><br/>Bye", msg.bodyHTML)
	})

	t.Run("HTML to plain", func(t *testing.T) {
		msg := &Message{
			bodyHTML: `<h1>Hello!</h1>
<p>Second, string</p>
Some other string<br/>

Second other string<br>


Bye`,
		}
		err := msg.Prepare(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, "Hello!\r\n\r\nSecond, string\r\n\r\nSome other string\r\nSecond other string\r\nBye", msg.bodyPlain)
	})
}
