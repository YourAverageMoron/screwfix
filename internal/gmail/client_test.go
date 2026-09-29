package gmail

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"

	gmailapi "google.golang.org/api/gmail/v1"
)

func TestFindPartNested(t *testing.T) {
	root := &gmailapi.MessagePart{MimeType: "multipart/alternative", Parts: []*gmailapi.MessagePart{
		{MimeType: "text/plain", Body: &gmailapi.MessagePartBody{Data: "eA"}},
		{MimeType: "multipart/related", Parts: []*gmailapi.MessagePart{
			{MimeType: "text/html", Body: &gmailapi.MessagePartBody{Data: "aGk="}},
		}},
	}}
	p := findPart(root, "text/html")
	if p == nil || p.Body.Data != "aGk=" {
		t.Fatalf("expected nested html part, got %+v", p)
	}
	if findPart(root, "text/html") == nil || findPart(root, "image/png") != nil {
		t.Fatal("findPart search broken")
	}
}

func TestDecodeB64url(t *testing.T) {
	cases := map[string]string{
		base64.URLEncoding.EncodeToString([]byte("hello")):  "hello",
		base64.RawURLEncoding.EncodeToString([]byte("<b>")): "<b>",
	}
	for in, want := range cases {
		got, err := decodeB64url(in)
		if err != nil {
			t.Fatalf("%s: %v", in, err)
		}
		if string(got) != want {
			t.Fatalf("decode %q = %q, want %q", in, got, want)
		}
	}
}

func TestMessageHTMLPlainFallback(t *testing.T) {
	m := &gmailapi.Message{Payload: &gmailapi.MessagePart{
		MimeType: "text/plain",
		Body:     &gmailapi.MessagePartBody{Data: base64.RawURLEncoding.EncodeToString([]byte("a<b"))},
	}}
	s, err := messageHTML(m)
	if err != nil || !strings.Contains(s, "a&lt;b") {
		t.Fatalf("got %q err %v", s, err)
	}
}

func TestBuildMimeAttachmentRoundtrip(t *testing.T) {
	att := []byte("%PDF-1.4 \xff\xfe\x00 binary \n payload")
	msg := buildMime("a@x.com", "b@y.com", "s", "body", "file.pdf", att)
	b64 := msg[strings.LastIndex(msg, "base64\r\n\r\n")+len("base64\r\n\r\n"):]
	b64 = b64[:strings.Index(b64, "\r\n--")]
	got, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(b64, "\r\n", ""))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, att) {
		t.Fatalf("attachment corrupted: %d bytes in, %d out", len(att), len(got))
	}
}
