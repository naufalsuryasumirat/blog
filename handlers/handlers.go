package handlers

import (
	"log"
	"os"
)

var domain string

func chk(err error) {
    if err != nil {
        log.Panic(err)
    }
}

func init() {
    if os.Getenv("BLOG_MODE") == "prod" {
        domain = os.Getenv("BLOG_DOMAIN_PROD")
    } else {
        domain = os.Getenv("BLOG_DOMAIN_DEV")
    }

    if len(domain) == 0 {
        log.Panic("Blog domain empty, exiting application")
    }
}
