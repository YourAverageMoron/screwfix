package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/YourAverageMoron/screwfix/internal/gmail"
	"github.com/YourAverageMoron/screwfix/internal/pdf"
)

func main() {
	subject := flag.String("subject", "", "email subject to search for")
	from := flag.String("from", "", "sender email to search for")
	before := flag.String("before", "", "only emails before this date (YYYY-MM-DD)")
	after := flag.String("after", "", "only emails after this date (YYYY-MM-DD)")
	out := flag.String("out", "email-example.pdf", "output PDF path")
	id := flag.String("id", "", "id of message")
	flag.Parse()

	ctx := context.Background()
	client, err := gmail.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}

	msgId := *id
	if msgId == "" {
		opts := &gmail.GetEmailIdOpts{From: *from, Subject: *subject}
		if *before != "" {
			t, err := time.Parse("2006-01-02", *before)
			if err != nil {
				log.Fatalf("bad -before date %q: %v", *before, err)
			}
			opts.Before = t
		}
		if *after != "" {
			t, err := time.Parse("2006-01-02", *after)
			if err != nil {
				log.Fatalf("bad -after date %q: %v", *after, err)
			}
			opts.After = t
		}
		ids, err := client.GetEmailIds(ctx, opts)
		if err != nil {
			log.Fatal(err)
		}
		msgId = ids[0]
	}

	f, err := os.CreateTemp("", "gmail-pdf-*.html")
	if err != nil {
		log.Fatal(err)
	}
	htmlPath := f.Name()
	defer os.Remove(htmlPath)
	if err := client.ReadToHtmlFile(ctx, msgId, f); err != nil {
		f.Close()
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}

	if err := pdf.RenderFromHtml(htmlPath, *out); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Wrote %s\n", *out)

    pfdF, err := os.Open(*out)
    if err != nil {
        log.Fatal(err)
    }

    err = client.SendFile(ctx, &gmail.SendOpts{
        From:    "ryannffc21@gmail.com",
        To:      "youraveragemoron@kindle.com",
        Subject: "example",
        Body:    "",
    }, pfdF)
    if err != nil {
        log.Fatal(err)
    }
	fmt.Printf("Sent Email %s\n", *out)
}
