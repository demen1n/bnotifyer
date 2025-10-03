package config

import (
	"errors"
	"strconv"
	"strings"

	"github.com/ilyakaznacheev/cleanenv"
)

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

type DBConfig struct {
	Name    string `yaml:"name"`
	Path    string `yaml:"path"`
	File    string `yaml:"file"`
	Weekday int    `yaml:"weekday"`
}

type Config struct {
	ST  SearchedText `yaml:"searched_text"`
	EFS EmailForSend `yaml:"email_for_send"`
	DB  []DBConfig   `yaml:"db"`
}

func (c *Config) Validate() error {
	if c.ST.Backup == "" {
		return errors.New("backup search text is empty")
	}
	if c.ST.Restore == "" {
		return errors.New("restore search text is empty")
	}

	if c.EFS.ServerAddress == "" {
		return errors.New("email server address is empty")
	}
	if c.EFS.ServerPort == 0 {
		return errors.New("email server port is not set")
	}
	if c.EFS.User == "" {
		return errors.New("email user is empty")
	}
	if c.EFS.Password == "" {
		return errors.New("email password is empty")
	}
	if len(c.EFS.MailTo) == 0 {
		return errors.New("no email recipients specified")
	}

	if len(c.DB) == 0 {
		return errors.New("no databases configured")
	}

	for i, db := range c.DB {
		if db.Name == "" {
			return errors.New("database name is empty at index " + strconv.Itoa(i))
		}
		if db.Path == "" {
			return errors.New("database path is empty for " + db.Name)
		}
		if db.File == "" {
			return errors.New("database file is empty for " + db.Name)
		}
		if db.Weekday < -1 || db.Weekday > 6 {
			return errors.New("invalid weekday value for " + db.Name + " (must be -1 to 6)")
		}
	}

	return nil
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

	if err := c.Validate(); err != nil {
		return nil, err
	}

	return &c, nil
}
