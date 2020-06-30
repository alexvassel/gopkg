package provider

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/k3a/html2text"
	"github.com/severgroup-tt/gopkg-app/types"
	errors "github.com/severgroup-tt/gopkg-errors"
	"strings"
	"time"
)

var nl2br = strings.NewReplacer("\r\n", "<br/>", "\n", "<br/>")
var html2plain = strings.NewReplacer("\n", "", "\r", "", "<br>", "<br/>", "</br>", "")

// message

type Message struct {
	Subject      string
	to           ContactList
	bodyPlain    string
	bodyHTML     string
	calendarCard *CalendarCard
}

func (m *Message) Prepare(ctx context.Context) error {
	if m.bodyPlain == "" && m.bodyHTML == "" {
		if m.calendarCard != nil {
			m.bodyPlain = "Вам назначена встреча (во вложении)"
		} else {
			return errors.Internal.Err(ctx, "Ошибка при отправке письма").
				WithLogKV("err", "body can't be empty")
		}
	}
	if m.bodyPlain == "" {
		m.bodyPlain = html2text.HTML2Text(html2plain.Replace(m.bodyHTML))
	}
	if m.bodyHTML == "" {
		m.bodyHTML = nl2br.Replace(m.bodyPlain)
	}
	return nil
}

func (m *Message) WithTo(addressNamePair ...string) *Message {
	if m.to == nil {
		m.to = make(ContactList, 0, len(addressNamePair)/2)
	}
	for i := 0; i < len(addressNamePair); i += 2 {
		m.to = append(m.to, &Contact{Address: addressNamePair[i], Name: addressNamePair[i+1]})
	}
	return m
}

func (m *Message) WithToAddress(address ...string) *Message {
	if m.to == nil {
		m.to = make(ContactList, 0, len(address))
	}
	for _, a := range address {
		m.to = append(m.to, &Contact{Address: a})
	}
	return m
}

func (m *Message) WithPlain(body string) *Message {
	m.bodyPlain = body
	return m
}

func (m *Message) WithHTML(body string) *Message {
	m.bodyHTML = body
	return m
}

func (m *Message) WithCalendarCard(card *CalendarCard) *Message {
	m.calendarCard = card
	return m
}

// calendar card

type CalendarCard struct {
	Name        string
	Location    string
	Description string
	IsAllDay    bool
	Start       time.Time
	Finish      time.Time
}

func (i CalendarCard) GetContent() string {
	var start, end string
	if i.IsAllDay {
		start = i.Start.Format("20060102")
		end = i.Finish.Format("20060102")
	} else {
		start = types.DateTimeToYMDTHms(i.Start)
		end = types.DateTimeToYMDTHms(i.Finish)
	}
	return base64.StdEncoding.EncodeToString([]byte(
		fmt.Sprintf(`BEGIN:VCALENDAR
VERSION:2.0
CALSCALE:GREGORIAN
BEGIN:VEVENT
SUMMARY:%s
DTSTART;TZID=Europe/Moscow:%s
DTEND;TZID=Europe/Moscow:%s
LOCATION:%s
DESCRIPTION:%s
STATUS:CONFIRMED
SEQUENCE:3
BEGIN:VALARM
TRIGGER:-PT10M
DESCRIPTION:%s
ACTION:DISPLAY
END:VALARM
END:VEVENT
END:VCALENDAR`, i.Name, start, end, i.Location, i.Description, i.Name)))
}
