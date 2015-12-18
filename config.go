package main

import (
	"gopkg.in/gcfg.v1"
	"io/ioutil"
	"log"
)

type Config struct {
	IRC struct {
		UseTLS        bool
		SkipTLSVerify bool
		Server        string
		Port          int
		Nick          string
		Channel       string
	}
	Mattermost struct {
		Server       string
		Team         string
		Login        string
		Password     string
		Channel      string
		ShowJoinPart bool
	}
	General struct {
		GiphyAPIKey string
	}
}

func NewConfig(cfgfile string) *Config {
	var cfg Config
	content, err := ioutil.ReadFile(cfgfile)
	if err != nil {
		log.Fatal(err)
	}
	err = gcfg.ReadStringInto(&cfg, string(content))
	if err != nil {
		log.Fatal("Failed to parse "+cfgfile+":", err)
	}
	return &cfg
}
