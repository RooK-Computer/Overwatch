package main

import (
	"log"
	"os"
	"strconv"
	"sync/atomic"
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
	var lastState atomic.Int32
	// Notification channel pattern
	notify := make(chan struct{}, 1)

	timer := time.NewTimer(10 * time.Second)
	if !timer.Stop() {
		<-timer.C
	}

	// Doing this here once, so we can avoid printf
	pinNum := strconv.Itoa(pinConfig.PinNumber)

	line, err := chip.RequestLine(pinConfig.PinNumber, gpiocdev.WithPullUp, gpiocdev.WithBothEdges,
		gpiocdev.WithEventHandler(func(evt gpiocdev.LineEvent) {
			var value int32
			switch evt.Type {
			case gpiocdev.LineEventRisingEdge:
				value = 1
			case gpiocdev.LineEventFallingEdge:
				value = 0
			default:
				log.Print("Unknown event type for pin: " + pinNum)
				return
			}

			lastState.Store(value)

			select {
			case notify <- struct{}{}:
			default:
			}
		}))
	if err != nil {
		log.Print("Failed to request line for pin:" + pinNum)
		return
	}
	defer line.Close()

	if value, err := line.Value(); err != nil {
		log.Print("Failed to read initial value for pin: " + pinNum + " " + err.Error())
	} else if value == 0 {
		lastState.Store(0)
		select {
		case notify <- struct{}{}:
		default:
		}
	}

	for {
		select {
		case <-notify:
			timer.Reset(50 * time.Millisecond)

		case <-timer.C:
			// LOW-signal (pressed with pull-up)
			if lastState.Load() == 0 {
				if pinConfig.PressedFile != "" {
					f, err := os.OpenFile(pinConfig.PressedFile, os.O_WRONLY|os.O_CREATE, 0644)
					if err != nil {
						log.Print("Failed to create pressed_file for pin: " + pinNum + " " + err.Error())
					} else {
						f.Close()
					}
				}

				errExec := pinConfig.Execute()
				if errExec != nil {
					log.Print("Failed to execute command for pin: " + pinNum + " " + errExec.Error())
				}
			} else {
				if pinConfig.PressedFile != "" {
					if err := os.Remove(pinConfig.PressedFile); err != nil && !os.IsNotExist(err) {
						log.Print("Failed to remove pressed_file for pin: " + pinNum + " " + err.Error())
					}
				}
			}
		}
	}
}
