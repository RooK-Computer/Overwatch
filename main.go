package main

import (
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/warthog618/go-gpiocdev"
)

// Config repräsentiert die Struktur der Konfigurationsdatei
type Config struct {
	Pins []PinConfig `json:"pins"`
}

// PinConfig enthält die Konfiguration für einen einzelnen GPIO-Pin
type PinConfig struct {
	PinNumber int    `json:"pin_number"`
	Command   string `json:"command"`
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <config-file>", os.Args[0])
	}

	configFile := os.Args[1]
	config, err := loadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	chip, err := gpiocdev.NewChip("gpiochip0")
	if err != nil {
		log.Fatalf("Failed to open GPIO chip: %v", err)
	}
	defer chip.Close()

	for _, pinConfig := range config.Pins {
		go monitorPin(chip, pinConfig)
	}

	// Block forever
	select {}
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

func monitorPin(chip *gpiocdev.Chip, pinConfig PinConfig) {
	type EventType int

	const (
		GPIOEvent EventType = iota
		TimerEvent
	)

	type Event struct {
		Type      EventType
		Timestamp time.Time
		Value     int
	}

	eventQueue := make(chan Event, 10)
	timer := time.NewTimer(10 * time.Second)
	if !timer.Stop() {
		<-timer.C
	}

	line, err := chip.RequestLine(pinConfig.PinNumber, gpiocdev.WithPullUp, gpiocdev.WithBothEdges, gpiocdev.WithEventHandler(func(evt gpiocdev.LineEvent) {
		var value int
		if evt.Type == gpiocdev.LineEventRisingEdge {
			value = 1
		} else if evt.Type == gpiocdev.LineEventFallingEdge {
			value = 0
		} else {
			log.Printf("Unknown event type for pin %d", pinConfig.PinNumber)
			return
		}
		eventQueue <- Event{Type: GPIOEvent, Timestamp: time.Now(), Value: value}
	}))
	if err != nil {
		log.Printf("Failed to request line for pin %d: %v", pinConfig.PinNumber, err)
		return
	}
	defer line.Close()

	log.Printf("Monitoring pin %d", pinConfig.PinNumber)

	go func() {
		var lastState int
		for event := range eventQueue {
			switch event.Type {
			case GPIOEvent:
				timer.Reset(50 * time.Millisecond)
				lastState = event.Value
			case TimerEvent:
				if lastState == 0 { // LOW-Signal
					log.Printf("Pin %d is LOW, executing command: %s", pinConfig.PinNumber, pinConfig.Command)
					executeCommand(pinConfig.Command)
				}
			}
		}
	}()

	for {
		<-timer.C
		eventQueue <- Event{Type: TimerEvent, Timestamp: time.Now()}
	}
}

func executeCommand(command string) {
	cmd := exec.Command(command)
	err := cmd.Run()
	if err != nil {
		log.Printf("Failed to execute command %s: %v", command, err)
	}
}
