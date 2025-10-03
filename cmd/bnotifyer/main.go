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
	defer func() {
		if cerr := f.Close(); cerr != nil {
			log.Printf("Error closing log file: %v\n", cerr)
		}
	}()
	log.SetOutput(f)

	log.Println("=== Starting bnotifyer ===")

	cfg, err := config.New("config.yml")
	if err != nil {
		log.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	log.Printf("Loaded config: %d databases configured\n", len(cfg.DB))

	analyzer.Do(cfg)

	log.Println("=== Finished bnotifyer ===")
}
