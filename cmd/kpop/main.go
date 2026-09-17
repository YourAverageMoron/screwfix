package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"html"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	pdf "github.com/YourAverageMoron/screwfix/internal/pdfer"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

func main() {
	subject := flag.String("subject", "", "email subject to search for")
	from := flag.String("from", "", "sender email to search for")
	out := flag.String("out", "email-example.pdf", "output PDF path")
    id := flag.String("id", "", "id of message")
	flag.Parse()

	ctx := context.Background()
	client, err := gmailClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	svc, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatal(err)
	}

	msg, err := latestMessage(svc, *from, *subject, *id)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found: %q from %s (%s)\n", header(msg, "Subject"), header(msg, "From"), header(msg, "Date"))

	body, err := messageHTML(msg)
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.CreateTemp("", "gmail-pdf-*.html")
	if err != nil {
		log.Fatal(err)
	}
	htmlPath := f.Name()
	defer os.Remove(htmlPath)
	if _, err := f.WriteString(body); err != nil {
		f.Close()
		log.Fatal(err)
	}
	f.Close()

	if err := pdf.RenderFromHtml(htmlPath, *out); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Wrote %s\n", *out)
}

func latestMessage(svc *gmail.Service, from, subject, id string) (*gmail.Message, error) {
    if id != "" {
	    return svc.Users.Messages.Get("me", id).Format("full").Do()
    }

	var q []string
	if from != "" {
		q = append(q, "from:"+from)
	}
	if subject != "" {
		q = append(q, `subject:"`+subject+`"`)
	}
	if len(q) == 0 {
		q = append(q, "in:inbox")
	}
	list, err := svc.Users.Messages.List("me").Q(strings.Join(q, " ")).MaxResults(1).Do()
	if err != nil {
		return nil, err
	}
	if len(list.Messages) == 0 {
		return nil, fmt.Errorf("no message matching: %s", strings.Join(q, " "))
	}
	return svc.Users.Messages.Get("me", list.Messages[0].Id).Format("full").Do()
}

func header(m *gmail.Message, name string) string {
	for _, h := range m.Payload.Headers {
		if strings.EqualFold(h.Name, name) {
			return h.Value
		}
	}
	return ""
}

func findPart(p *gmail.MessagePart, mime string) *gmail.MessagePart {
	if p == nil {
		return nil
	}
	if p.MimeType == mime && p.Body != nil && p.Body.Data != "" {
		return p
	}
	for _, c := range p.Parts {
		if r := findPart(c, mime); r != nil {
			return r
		}
	}
	return nil
}

func decodeB64url(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(strings.TrimRight(s, "="))
}

func messageHTML(m *gmail.Message) (string, error) {
	if p := findPart(m.Payload, "text/html"); p != nil {
		b, err := decodeB64url(p.Body.Data)
		return string(b), err
	}
	if p := findPart(m.Payload, "text/plain"); p != nil {
		b, err := decodeB64url(p.Body.Data)
		if err != nil {
			return "", err
		}
		return "<pre>" + html.EscapeString(string(b)) + "</pre>", nil
	}
	return "", fmt.Errorf("no text body found in message")
}



func gmailClient(ctx context.Context) (*http.Client, error) {
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		return nil, fmt.Errorf("read credentials.json (Google OAuth client, Desktop type): %w", err)
	}
	cfg, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
	if err != nil {
		return nil, err
	}
	tok, err := loadToken("token.json")
	if err != nil {
		tok, err = webToken(ctx, cfg)
		if err != nil {
			return nil, err
		}
		if err := saveToken("token.json", tok); err != nil {
			return nil, err
		}
	}
	return cfg.Client(ctx, tok), nil
}

func webToken(ctx context.Context, cfg *oauth2.Config) (*oauth2.Token, error) {
	host, port := "localhost", "0"
	if u, err := url.Parse(cfg.RedirectURL); err != nil || (u.Host != "localhost" && u.Host != "127.0.0.1") {
		return nil, fmt.Errorf("redirect URI %q is not loopback; add http://localhost to the OAuth client's authorized redirect URIs (Desktop app type)", cfg.RedirectURL)
	} else if p := u.Port(); p != "" {
		host, port = u.Host, p
	}

	l, err := net.Listen("tcp", net.JoinHostPort(host, port))
	if err != nil {
		return nil, err
	}
	defer l.Close()
	if _, p, err := net.SplitHostPort(l.Addr().String()); err == nil {
		cfg.RedirectURL = "http://" + host + ":" + p
	}

	sb := make([]byte, 8)
	if _, err := rand.Read(sb); err != nil {
		return nil, err
	}
	state := hex.EncodeToString(sb)

	codeCh := make(chan string, 1)
	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("state") != state {
			http.Error(w, "bad state", http.StatusBadRequest)
			return
		}
		if c := q.Get("code"); c != "" {
			fmt.Fprintln(w, "Authorized — you can close this tab.")
			select {
			case codeCh <- c:
			default:
			}
		}
	})}
	errCh := make(chan error, 1)
	go func() { errCh <- srv.Serve(l) }()
	defer srv.Close()

	authURL := cfg.AuthCodeURL(state, oauth2.AccessTypeOffline)
	fmt.Println("Opening browser for Google sign-in. If it does not open, visit:")
	fmt.Println(authURL)
	openBrowser(authURL)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	select {
	case code := <-codeCh:
		return cfg.Exchange(ctx, code)
	case err := <-errCh:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func openBrowser(u string) {
	switch runtime.GOOS {
	case "darwin":
		exec.Command("open", u).Start()
	case "windows":
		exec.Command("cmd", "/c", "start", u).Start()
	default:
		exec.Command("xdg-open", u).Start()
	}
}

func loadToken(path string) (*oauth2.Token, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	return t, json.Unmarshal(b, t)
}

func saveToken(path string, t *oauth2.Token) error {
	b, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o600)
}
