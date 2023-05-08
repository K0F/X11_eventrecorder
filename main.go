package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-vgo/robotgo"
	hook "github.com/robotn/gohook"
)

type MousePos struct {
	tc   int64
	kind int
	x    int
	y    int
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
	flag.StringVar(&filename, "p", "", "CSV file to re-play.")

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

	startTime = time.Now()
	positions = make([]MousePos, 0)

	if recording {

		// run on separate thread
		go CurrentMousePosition()
		go low()
	}

	if playing {

		positions, err := readSplice(filename)
		if err != nil {
			log.Fatalln("Error reading file.")
		}

		go Replay(positions)

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

func low() {
	evChan := hook.Start()
	defer hook.End()

	for ev := range evChan {

		tc := time.Since(startTime).Milliseconds()

		if ev.Clicks == 1 && ev.Kind == hook.MouseDown {
			positions = append(positions, MousePos{tc: tc, kind: int(ev.Kind), x: int(ev.X), y: int(ev.Y)})
			fmt.Printf("tc:%d kind:%d X:%d Y:%d\n", tc, ev.Kind, ev.X, ev.Y)
		}
	}
}

func cleanup() {

	fmt.Println("Interrupt SIG received, cleaning up and saving session.")
	if recording {
		saveSplice(positions)
	}

	os.Exit(0)
}

func readSplice(filename string) ([]MousePos, error) {

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	var pos []MousePos

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		tc, err := strconv.ParseInt(record[0], 10, 64)
		if err != nil {
			return nil, err
		}
		kind, err := strconv.Atoi(record[1])
		if err != nil {
			return nil, err
		}

		x, err := strconv.Atoi(record[2])
		if err != nil {
			return nil, err
		}
		y, err := strconv.Atoi(record[3])
		if err != nil {
			return nil, err
		}

		p := MousePos{
			tc:   tc,
			kind: kind,
			x:    x,
			y:    y,
		}

		pos = append(pos, p)
	}

	return pos, nil
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

	for _, pos := range input {
		err := writer.Write([]string{strconv.FormatInt(pos.tc, 10), strconv.Itoa(pos.kind), strconv.Itoa(pos.x), strconv.Itoa(pos.y)})
		if err != nil {
			log.Fatal("Error writing record to CSV:", err)
		}

	}

	log.Printf("Saving: %s", filename)

}

var cnt = 0

func Replay(pos []MousePos) {
	for {
		x := pos[cnt].x
		y := pos[cnt].y

		//FIXME: precise timing using tc?
		robotgo.MicroSleep(10)
		if pos[cnt].kind == 0 {
			robotgo.Move(x, y)
		} else if pos[cnt].kind == 8 {
			robotgo.Click()
		}

		cnt++

		if cnt >= len(pos) {
			log.Printf("Replay finished, looping.")
			cnt = 0
		}
	}

}

func CurrentMousePosition() {
	for {
		robotgo.MicroSleep(10)
		x, y := robotgo.GetMousePos()
		tc := time.Since(startTime).Milliseconds()
		kind := 0
		pos := MousePos{tc: tc, kind: 0, x: x, y: y}

		positions = append(positions, pos)
		fmt.Printf("tc:%d kind:%d x:%d y:%d\n", tc, kind, x, y)
	}
}
