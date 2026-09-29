package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/YourAverageMoron/screwfix/internal/kpop"
	"github.com/spf13/viper"
)

type Config struct {
	KindleEmail string `mapstructure:"kindle_email"`
	GmailEmail  string `mapstructure:"gmail_email"`
}

func loadConfig(c any) error {
	v := viper.New()
	v.SetConfigName("kpop")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("$HOME/.config/screwfix")
	v.SetEnvPrefix("KPOP")
	v.AutomaticEnv()
	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return fmt.Errorf("read kpop config: %w", err)
		}
	}
	if err := v.Unmarshal(&c); err != nil {
		return err
	}

	return nil
}

func parseFlags() (kpop.Args, error) {
	subject := flag.String("subject", "", "email subject to search for")
	from := flag.String("from", "", "sender email to search for")
	before := flag.String("before", "", "only emails before this date (YYYY-MM-DD)")
	after := flag.String("after", "", "only emails after this date (YYYY-MM-DD)")
	flag.Parse()
	a := kpop.Args{
		Subject: *subject,
		From:    *from,
	}
	if *before != "" {
		parsedB, err := time.Parse("2006-01-02", *before)
		if err != nil {
			return kpop.Args{}, fmt.Errorf("bad -before date %q: %v", *before, err)
		}
		a.Before = parsedB
	}
	if *after != "" {
		parsedA, err := time.Parse("2006-01-02", *after)
		if err != nil {
			return kpop.Args{}, fmt.Errorf("bad -before date %q: %v", *before, err)
		}
		a.After = parsedA
	}
	return a, nil
}

func main() {
	ctx := context.Background()
	cfg := &kpop.Config{}
	err := loadConfig(cfg)
	if err != nil {
		log.Fatal(err)
	}
	app := kpop.NewApp(*cfg)
	args, err := parseFlags()
	if err != nil {
		log.Fatal(err)
	}
	res, err := app.SendEmailsToKindle(ctx, args)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res)
}
