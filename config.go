package main

import (
	"encoding/json"
	"os"
)

// Config repräsentiert die Struktur der Konfigurationsdatei
type Config struct {
	Pins []PinConfig `json:"pins"`
}

// PinConfig enthält die Konfiguration für einen einzelnen GPIO-Pin
type PinConfig struct {
	PinNumber   int    `json:"pin_number"`
	Command     string `json:"command"`
	PressedFile string `json:"pressed_file,omitempty"`
}

func loadConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
