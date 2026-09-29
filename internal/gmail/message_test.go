package gmail

import (
	"testing"

	"github.com/stretchr/testify/assert"
	gmailapi "google.golang.org/api/gmail/v1"
)

func msg(headers ...[2]string) *Message {
	hs := make([]*gmailapi.MessagePartHeader, 0, len(headers))
	for _, h := range headers {
		hs = append(hs, &gmailapi.MessagePartHeader{Name: h[0], Value: h[1]})
	}
	return &Message{msg: &gmailapi.Message{Payload: &gmailapi.MessagePart{Headers: hs}}}
}

func TestGetId(t *testing.T) {
	tests := []struct {
		name string
		m    *Message
		want string
	}{
		{
			name: "full headers",
			m: msg(
				[2]string{"From", "Sensemaker at The Observer <editor@newsroom.observer.co.uk>"},
				[2]string{"Date", "Tue, 22 Sep 2026 06:01:12 +0000"},
				[2]string{"Subject", "Germany forgets"},
			),
			want: "editor-newsroom-observer-co-uk_2026-09-22_Germany-forgets",
		},
		{
			name: "bare address as from",
			m: msg(
				[2]string{"From", "bob@example.com"},
				[2]string{"Date", "Mon, 1 Jun 2026 10:00:00 +0100"},
				[2]string{"Subject", "Hello"},
			),
			want: "bob-example-com_2026-06-01_Hello",
		},
		{
			name: "punctuation in subject",
			m: msg(
				[2]string{"From", "a@b.io"},
				[2]string{"Date", "Wed, 3 Dec 2025 23:59:59 -0500"},
				[2]string{"Subject", "Re: Fwd: Quarterly report (Q4)!"},
			),
			want: "a-b-io_2025-12-03_Re-Fwd-Quarterly-report-Q4",
		},
		{
			name: "missing headers",
			m:    msg(),
			want: "__",
		},
		{
			name: "unparsable date falls back to raw header",
			m: msg(
				[2]string{"From", "x@y.z"},
				[2]string{"Date", "not a date"},
				[2]string{"Subject", "s"},
			),
			want: "x-y-z_not-a-date_s",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.m.GetId())
		})
	}
}

func TestSanitize(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"Germany forgets", "Germany-forgets"},
		{"  padded  ", "padded"},
		{"Re: Fwd: **urgent**", "Re-Fwd-urgent"},
		{"plain", "plain"},
		{"", ""},
		{"---leading-and-trailing---", "leading-and-trailing"},
		{"under_score-dash", "under_score-dash"},
		{"tabs\tand\nnewlines", "tabs-and-newlines"},
		{"emoji 🎉 drop", "emoji-drop"},
		{" déjà vu ", "d-j-vu"},
	}

	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			assert.Equal(t, tc.want, sanitize(tc.in))
		})
	}
}
