package main

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"strings"
)

// Config represents the config file structre
type Config struct {
	Pins []PinConfig `json:"pins"`
}

// PinConfig is the config for an individual pin
type PinConfig struct {
	PinNumber   int      `json:"pin_number"`
	Command     string   `json:"command"`
	PressedFile string   `json:"pressed_file,omitempty"`
	CommandArgs []string `json:"-"`
}

func (p *PinConfig) Execute() error {
	if len(p.CommandArgs) == 0 {
		return errors.New("Invalid command: " + p.Command)
	}

	//nolint: gosec
	cmd := exec.Command(p.CommandArgs[0], p.CommandArgs[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	return err
}

func loadConfig(filename string) (*Config, error) {
	//nolint: gosec
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	var config Config

	errDecode := json.NewDecoder(file).Decode(&config)
	if errDecode != nil {
		return nil, errDecode
	}

	// So that we don't have to do this everytime the button is pressed
	for i, pinConfig := range config.Pins {
		config.Pins[i].CommandArgs = strings.Fields(pinConfig.Command)
	}

	return &config, nil
}
