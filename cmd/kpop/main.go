package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/YourAverageMoron/screwfix/internal/gmail"
	"github.com/YourAverageMoron/screwfix/internal/pdf"
)

func main() {
	subject := flag.String("subject", "", "email subject to search for")
	from := flag.String("from", "", "sender email to search for")
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
		msgId, err = client.GetEmailId(ctx, *from, *subject)
		if err != nil {
			log.Fatal(err)
		}
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

    err = client.SendFile(ctx, "ryanNFFC21_7219Zm@kindle.com", "example", "example body", pfdF)
    if err != nil {
        log.Fatal(err)
    }
	fmt.Printf("Sent Email %s\n", *out)
}
