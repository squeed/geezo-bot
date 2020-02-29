package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/pkg/errors"

	"github.com/squeed/geezo-bot/pkg/app"
	"github.com/squeed/geezo-bot/pkg/config"
)

func main() {
	err := exec()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func exec() error {
	var configPath string
	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.Parse()
	if configPath == "" {
		return errors.New("--config is required")
	}

	conf, err := config.ReadConfig(configPath)
	if err != nil {
		return err
	}

	a, err := app.Init(conf)
	if err != nil {
		return err
	}

	defer a.Close()

	if err := a.ReadMail(); err != nil {
		return err
	}

	if err := a.DownloadImages(); err != nil {
		return err
	}

	if err := a.SendMessages(); err != nil {
		return err
	}

	return nil
}
