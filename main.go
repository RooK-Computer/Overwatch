package main

import (
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/warthog618/go-gpiocdev"
)

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
				if lastState == 0 { // LOW-Signal (pressed with pull-up)
					if pinConfig.PressedFile != "" {
						f, err := os.OpenFile(pinConfig.PressedFile, os.O_WRONLY|os.O_CREATE, 0644)
						if err != nil {
							log.Printf("Failed to create pressed_file %q for pin %d: %v", pinConfig.PressedFile, pinConfig.PinNumber, err)
						} else {
							_ = f.Close()
						}
					}

					log.Printf("Pin %d is LOW, executing command: %s", pinConfig.PinNumber, pinConfig.Command)
					executeCommand(pinConfig.Command)
				} else {
					if pinConfig.PressedFile != "" {
						err := os.Remove(pinConfig.PressedFile)
						if err != nil && !os.IsNotExist(err) {
							log.Printf("Failed to remove pressed_file %q for pin %d: %v", pinConfig.PressedFile, pinConfig.PinNumber, err)
						}
					}
				}
			}
		}
	}()

	if value, err := line.Value(); err != nil {
		log.Printf("Failed to read initial value for pin %d: %v", pinConfig.PinNumber, err)
	} else if value == 0 {
		// If the button is already pressed when the program starts, no edge event will fire.
		// Enqueue a synthetic event to kick off the existing debounce logic.
		eventQueue <- Event{Type: GPIOEvent, Timestamp: time.Now(), Value: value}
	}

	for {
		<-timer.C
		eventQueue <- Event{Type: TimerEvent, Timestamp: time.Now()}
	}
}

func executeCommand(command string) {
	args := strings.Fields(command)
	if len(args) == 0 {
		log.Printf("Invalid command: %s", command)
		return
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		log.Printf("Failed to execute command %s: %v", command, err)
	}
}
