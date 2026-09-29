package pdf

import (
	"context"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

func RenderFromHtml(htmlPath string) ([]byte, error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()
	ctx, cancel = context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	var buf []byte
	if err := chromedp.Run(ctx,
		chromedp.Navigate("file://"+htmlPath),
		waitImages,
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			buf, _, err = page.PrintToPDF().
				WithPrintBackground(true).
				WithDisplayHeaderFooter(false).
				Do(ctx)
			return err
		}),
	); err != nil {
		return nil, err
	}
	return buf, nil
}

var waitImages = chromedp.ActionFunc(func(ctx context.Context) error {
	for i := 0; i < 60; i++ {
		var done bool
		if err := chromedp.Evaluate(`Array.from(document.images).every(i => i.complete)`, &done).Do(ctx); err != nil {
			return err
		}
		if done {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(500 * time.Millisecond):
		}
	}
	return nil // ponytail: after 30s print anyway; upgrade path: per-image error surfaced in log
})
