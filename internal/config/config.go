package config

import (
	"strconv"
	"strings"

	"github.com/ilyakaznacheev/cleanenv"
)

// TODO: насыпать проверок того, что конфиг не пустой

type SearchedText struct {
	Backup  string `yaml:"text_backup"`
	Restore string `yaml:"text_restore"`
}

type EmailForSend struct {
	ServerAddress string   `yaml:"server_url"`
	ServerPort    int      `yaml:"server_port"`
	User          string   `yaml:"user"`
	Password      string   `yaml:"password"`
	MailTo        []string `yaml:"mail_to"`
}

func (efs *EmailForSend) Username() string {
	i := strings.Index(efs.User, "@")
	if i == -1 {
		return ""
	}
	return efs.User[:i]
}

func (efs *EmailForSend) Server() string {
	return efs.ServerAddress + ":" + strconv.Itoa(efs.ServerPort)
}

type Config struct {
	ST SearchedText `yaml:"searched_text"`

	EFS EmailForSend `yaml:"email_for_send"`

	DB []struct {
		Name    string `yaml:"name"`
		Path    string `yaml:"path"`
		File    string `yaml:"file"`
		Weekday int    `yaml:"weekday"`
	} `yaml:"db"`
}

func New(path string) (*Config, error) {
	var c Config
	var err error

	if path != "" {
		err = cleanenv.ReadConfig(path, &c)
	} else {
		err = cleanenv.ReadEnv(&c)
	}

	if err != nil {
		return nil, err
	}

	return &c, nil
}
