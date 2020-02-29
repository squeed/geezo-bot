package config

import (
	"io/ioutil"

	"github.com/ghodss/yaml"
)

type Config struct {
	Imap MailConfig `json:"imap"`
	Smtp MailConfig `json:"smtp"`
	Main MainConfig `json:"main"`
}

type MailConfig struct {
	Server   string `json:"server"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type MainConfig struct {
	DbFile  string `json:"dbFile"`
	From    string `json:"from"`
	To      string `json:"to"`
	WorkDir string `json:"workDir"`
}

func ReadConfig(path string) (*Config, error) {
	cb, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	out := Config{}
	err = yaml.Unmarshal(cb, &out)
	if err != nil {
		return nil, err
	}

	return &out, nil
}
