package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/lstoll/rwnewsengine"
	"go.pedge.io/env"
)

var c *rwnewsengine.Config

func main() {
	c = &rwnewsengine.Config{}
	if err := env.Populate(c); err != nil {
		log.Fatal(err)
	}

	rwnewsengine.HTTPSetup(c)
	rwnewsengine.HTTPServe(c)
}
