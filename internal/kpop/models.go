package kpop

import "time"

type Config struct {
	KindleEmail string `mapstructure:"kindle_email"`
	GmailEmail  string `mapstructure:"gmail_email"`
}

type Args struct {
	Subject string
	From    string
	Before  time.Time
	After   time.Time
}

type Response struct {
	Files []string
    Errors []error
}
