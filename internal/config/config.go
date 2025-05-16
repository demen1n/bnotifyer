package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"strconv"
	"strings"
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

type Database struct {
	Name    string `yaml:"name"`
	Path    string `yaml:"path"`
	File    string `yaml:"file"`
	Weekday int    `yaml:"weekday"`
}

type Config struct {
	BackupFolder string       `yaml:"backup_folder"`
	ST           SearchedText `yaml:"searched_text"`
	EFS          EmailForSend `yaml:"email_for_send"`
	DB           []Database   `yaml:"db"`
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
