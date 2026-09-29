package kpop

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/YourAverageMoron/screwfix/internal/gmail"
	"github.com/YourAverageMoron/screwfix/internal/pdf"
)

type App struct {
	cfg Config
}

func NewApp(cfg Config) *App {
	return &App{
		cfg: cfg,
	}
}

func (app *App) SendEmailsToKindle(ctx context.Context, args Args) (Response, error) {
	client, err := gmail.NewClient(ctx)
	if err != nil {
		return Response{}, err
	}

	opts := &gmail.GetEmailIdOpts{From: args.From, Subject: args.Subject, Before: args.Before, After: args.After}

	ids, err := client.GetEmailIds(ctx, opts)
	if err != nil {
		return Response{}, err
	}


    res := Response{}

	for _, id := range ids {
        file, err := sendEmailToKindle(ctx, client, id, app.cfg.GmailEmail, app.cfg.KindleEmail)
        if err != nil{
            res.Errors = append(res.Errors, err)
        }
        res.Files = append(res.Files, file)
	}

	return res, nil
}

func sendEmailToKindle(ctx context.Context, client *gmail.Client, msgId, gmailEmail, kindleEmail string) (string, error) {
	f, err := os.CreateTemp("", "gmail-html-*.html")
	if err != nil {
		return "", err
	}
	htmlPath := f.Name()
	defer f.Close()
	defer os.Remove(htmlPath)

	msg, err := client.GetMessage(ctx, msgId)
	if err != nil {
		return "", err
	}

	err = msg.WriteHTML(ctx, f)
	if err != nil {
		return "", err
	}

	pdfBytes, err := pdf.RenderFromHtml(htmlPath)
	if err != nil {
		return "", err
	}

    mId := msg.GetId()
	err = client.SendFile(ctx, bytes.NewReader(pdfBytes), fmt.Sprintf("%s.pdf", mId), &gmail.SendOpts{
		From:    gmailEmail,
		To:      kindleEmail,
		Subject: mId,
		Body:    "",
	})

	return mId, err
}
