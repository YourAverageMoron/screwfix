package gmail

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	gmailapi "google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type Gmail interface {
	GetMessage(ctx context.Context, id string) (*Message, error)
	GetEmailIds(ctx context.Context, opts *GetEmailIdOpts) ([]string, error)
	SendFile(ctx context.Context, r io.Reader, attachmentName string, opts *SendOpts) error
}

type Client struct {
	svc *gmailapi.Service
}

var _ Gmail = (*Client)(nil)

func NewClient(ctx context.Context) (*Client, error) {
	client, err := oauthClient(ctx)
	if err != nil {
		return nil, err
	}
	svc, err := gmailapi.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}
	return &Client{svc: svc}, nil
}

type GetEmailIdOpts struct {
	From    string
	Subject string
	Before  time.Time
	After   time.Time
}

func (c *Client) GetEmailIds(ctx context.Context, opts *GetEmailIdOpts) ([]string, error) {
	if opts == nil {
		opts = &GetEmailIdOpts{}
	}
	var q []string
	if opts.From != "" {
		q = append(q, "from:"+opts.From)
	}
	if opts.Subject != "" {
		q = append(q, `subject:"`+opts.Subject+`"`)
	}
	// Gmail only supports date granularity; switch to an internal timestamp filter if time-of-day precision needed
	if !opts.After.IsZero() {
		q = append(q, "after:"+opts.After.Format("2006/01/02"))
	}
	if !opts.Before.IsZero() {
		q = append(q, "before:"+opts.Before.Format("2006/01/02"))
	}
	if len(q) == 0 {
		q = append(q, "in:inbox")
	}
	list, err := c.svc.Users.Messages.List("me").Q(strings.Join(q, " ")).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}
	if len(list.Messages) == 0 {
		return nil, fmt.Errorf("no message matching: %s", strings.Join(q, " "))
	}
	ids := make([]string, len(list.Messages))
	for i, m := range list.Messages {
		ids[i] = m.Id
	}
	return ids, nil
}

func (c *Client) GetMessage(ctx context.Context, id string) (*Message, error) {
	msg, err := c.svc.Users.Messages.Get("me", id).Format("full").Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get msg: %w", err)
	}
	return &Message{msg: msg}, nil
}

type SendOpts struct {
	From    string
	To      string
	Subject string
	Body    string
}

func (c *Client) SendFile(ctx context.Context, r io.Reader, attachmentName string, opts *SendOpts) error {
	attachment, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("read attachment: %w", err)
	}
	_, err = c.svc.Users.Messages.Send(opts.From, &gmailapi.Message{
		Raw: base64.RawURLEncoding.EncodeToString([]byte(buildMime(opts.From, opts.To, opts.Subject, opts.Body, attachmentName, attachment))),
	}).Context(ctx).Do()
	return err
}

// ponytail: fixed boundary — fine unless body/attachment can contain "sfx-attach"; then use mime/multipart.Writer
func buildMime(from, to, subject, body, filename string, attachment []byte) string {
	const bnd = "sfx-attach"
	mimeType := mime.TypeByExtension(filepath.Ext(filename))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: multipart/mixed; boundary=%s\r\n\r\n", from, to, subject, bnd)
	fmt.Fprintf(&b, "--%s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s\r\n\r\n", bnd, body)
	fmt.Fprintf(&b, "--%s\r\nContent-Type: %s; name=%q\r\nContent-Disposition: attachment; filename=%q\r\nContent-Transfer-Encoding: base64\r\n\r\n", bnd, mimeType, filename, filename)
	enc := base64.StdEncoding.EncodeToString(attachment)
	for len(enc) > 76 {
		b.WriteString(enc[:76] + "\r\n")
		enc = enc[76:]
	}
	b.WriteString(enc + "\r\n")
	fmt.Fprintf(&b, "--%s--\r\n", bnd)
	return b.String()
}

func header(m *gmailapi.Message, name string) string {
	for _, h := range m.Payload.Headers {
		if strings.EqualFold(h.Name, name) {
			return h.Value
		}
	}
	return ""
}

func findPart(p *gmailapi.MessagePart, mime string) *gmailapi.MessagePart {
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

func messageHTML(m *gmailapi.Message) (string, error) {
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

func oauthClient(ctx context.Context) (*http.Client, error) {
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		return nil, fmt.Errorf("read credentials.json (Google OAuth client, Desktop type): %w", err)
	}
	cfg, err := google.ConfigFromJSON(b, gmailapi.GmailReadonlyScope, gmailapi.GmailSendScope)
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
