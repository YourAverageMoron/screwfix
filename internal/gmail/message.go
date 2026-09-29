package gmail

import (
	"context"
	"io"
	"net/mail"
	"regexp"
	"strings"

	gmailapi "google.golang.org/api/gmail/v1"
)

type Message struct {
	msg *gmailapi.Message
}

func (m *Message) WriteHTML(ctx context.Context, w io.Writer) error {
	body, err := messageHTML(m.msg)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(body))
	return err
}

// GetId returns a filename-safe identifier of the form from_date_subject.
func (m *Message) GetId() string {
	from := header(m.msg, "From")
	if a, err := mail.ParseAddress(from); err == nil {
		from = a.Address
	}

	date := header(m.msg, "Date")
	if t, err := mail.ParseDate(date); err == nil {
		date = t.Format("2006-01-02")
	}

	return sanitize(from) + "_" + sanitize(date) + "_" + sanitize(header(m.msg, "Subject"))
}

var unsafeChars = regexp.MustCompile(`[^a-zA-Z0-9_-]+`)

// sanitize replaces runs of characters that are unsafe in filenames with a single dash.
func sanitize(s string) string {
	return strings.Trim(unsafeChars.ReplaceAllString(s, "-"), "-")
}
