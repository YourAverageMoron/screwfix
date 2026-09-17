package main

import (
	"encoding/base64"
	"strings"
	"testing"

	"google.golang.org/api/gmail/v1"
)

func TestFindPartNested(t *testing.T) {
	root := &gmail.MessagePart{MimeType: "multipart/alternative", Parts: []*gmail.MessagePart{
		{MimeType: "text/plain", Body: &gmail.MessagePartBody{Data: "eA"}},
		{MimeType: "multipart/related", Parts: []*gmail.MessagePart{
			{MimeType: "text/html", Body: &gmail.MessagePartBody{Data: "aGk="}},
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
	m := &gmail.Message{Payload: &gmail.MessagePart{
		MimeType: "text/plain",
		Body:     &gmail.MessagePartBody{Data: base64.RawURLEncoding.EncodeToString([]byte("a<b"))},
	}}
	s, err := messageHTML(m)
	if err != nil || !strings.Contains(s, "a&lt;b") {
		t.Fatalf("got %q err %v", s, err)
	}
}
