package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-vgo/robotgo"
)

type MousePos struct {
	tc int64
	x  int
	y  int
}

var positions []MousePos
var startTime time.Time
var recording bool
var playing bool

func main() {

	// Define a flag variable to store the value of the -r flag
	var rFlag bool
	var filename string

	// Parse the command line arguments to populate the flag variable
	flag.BoolVar(&rFlag, "r", false, "Record session")
	flag.StringVar(&filename, "p", "", "CSV file to reply.")

	// Call the flag.Parse() function to parse the command line arguments
	flag.Parse()

	if rFlag {
		fmt.Println("Recording...")
		recording = true
	}

	if filename != "" {
		if recording {
			panic("Cannot record and play the same time")
		}
		fmt.Sprintln("Re-playing file: %s", filename)
		playing = true
	}

	positions = make([]MousePos, 0)
	startTime = time.Now()

	if recording {
		// run on separate thread
		go CurrentMousePosition()
	}

	//// HANDLE INTERRUPT SIGNAL //////////////////////////////////////////
	// create a channel to receive signals
	sigCh := make(chan os.Signal, 1)

	// register the channel to receive SIGINT signal (ctrl+c)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGINT)

	// wait for the signal
	<-sigCh

	// run the cleanup function
	cleanup()
	///////////////////////////////////////////////////////////////////
}

func cleanup() {

	fmt.Println("Interrupt SIG received, cleaning up and saving session.")
	if recording {
		saveSplice(positions)
	}

	os.Exit(0)
}

func saveSplice(input []MousePos) {

	// Get the current time
	currentTime := time.Now()

	// Format the current time as a string
	timestamp := currentTime.Format("2006_01_02-150405")

	filename := fmt.Sprintf("%s.csv", timestamp)
	// Open the file for writing
	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, pos := range positions {
		err := writer.Write([]string{strconv.FormatInt(pos.tc, 10), strconv.Itoa(pos.x), strconv.Itoa(pos.y)})
		if err != nil {
			log.Fatal("Error writing record to CSV:", err)
		}

	}

	log.Printf("Saving: %s", filename)

}

func CurrentMousePosition() {
	for {
		robotgo.MicroSleep(10)
		x, y := robotgo.GetMousePos()
		tc := time.Since(startTime).Milliseconds()
		pos := MousePos{tc: tc, x: x, y: y}
		positions = append(positions, pos)
		fmt.Printf("tc:%d x:%d y:%d\n", tc, x, y)
	}
}
