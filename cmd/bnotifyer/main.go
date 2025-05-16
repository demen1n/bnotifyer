package main

import (
	"bnotifyer/internal/analyzer"
	"bnotifyer/internal/config"
	"log"
	"os"
)

func main() {
	f, err := os.OpenFile("info.log", os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	log.SetOutput(f)

	// load config
	var cfg *config.Config
	cfg, err = config.New("config.yml")
	if err != nil {
		log.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	if len(cfg.ST.Backup) == 0 {
		log.Printf("no backup config\n")
		os.Exit(1)
	}

	if len(cfg.ST.Restore) == 0 {
		log.Printf("no restore config\n")
		os.Exit(1)
	}

	analyzer.Do(cfg)
}
