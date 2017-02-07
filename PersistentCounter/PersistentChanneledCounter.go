package PersistentCounter

import (
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"
)

type PersistentChanneledCounter struct {
	file              *os.File
	persistFrequency  time.Duration
	persistSeconds    uint64
	inputChannel      chan uint64
	shutdownChannel   chan os.Signal
	outputChannel     chan uint64
	previousTimeStamp time.Time
	buffer            *RingBuffer
	stringData        []string
}

func NewPersistentChanneledCounter(fileName string, persistFrequency time.Duration, persistSeconds uint64) (
	pCounter *PersistentChanneledCounter, err error) {

	newFile := false
	// Check if file exists first
	if _, err = os.Stat(fileName); err != nil {
		log.Printf("File %s does not exists and will be created automatically", fileName)
		_, err := os.Create(fileName)
		if err != nil {
			return nil, fmt.Errorf("Unable to create file %s", fileName)
		}
		newFile = true
	}
	// File exists, check write permission
	file, err := os.OpenFile(fileName, os.O_RDWR, 0666)
	if err != nil {
		return nil, fmt.Errorf("Unable to open file %s for reading and writing", fileName)
	}

	pCounter = new(PersistentChanneledCounter)
	pCounter.persistFrequency = persistFrequency
	pCounter.persistSeconds = persistSeconds
	pCounter.file = file
	pCounter.buffer = NewRingBuffer(persistSeconds)
	if !newFile {
		err = pCounter.loadBuffer()
	}
	pCounter.inputChannel = make(chan uint64, persistSeconds*1200)
	pCounter.outputChannel = make(chan uint64)
	pCounter.shutdownChannel = make(chan os.Signal)
	pCounter.stringData = make([]string, persistSeconds)

	signal.Notify(pCounter.shutdownChannel, os.Interrupt)

	//Counts requests per seconds and perform responses
	go pCounter.counterServiceWorker()

	return pCounter, nil
}

func (p *PersistentChanneledCounter) loadBuffer() (err error) {
	log.Println("Trying to restore previous counter")
	fileSize, err := p.file.Stat()
	if err != nil {
		return errors.New("Unable to get file information")
	}

	buffer := make([]byte, fileSize.Size())
	p.file.Seek(0, 0)
	n, err := p.file.Read(buffer)
	log.Printf("%d bytes readed\r\n", n)
	if err != nil {
		return errors.New("Unable to restore buffer from file")
	}

	stringData := strings.Split(string(buffer), ",")
	if len(stringData) == 0 {
		return errors.New("Corrupted file")
	}

	var result []uint64
	for _, s := range stringData {
		d, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return errors.New("Unable to parse file")
		}
		result = append(result, d)
	}
	p.buffer.CopyDataFrom(result)
	return nil
}

func (p *PersistentChanneledCounter) saveBuffer() {
	p.file.Truncate(0)
	p.file.Seek(0, 0)
	for i := 0; i < len(p.buffer.items); i++ {
		p.stringData[i] = strconv.Itoa(int(p.buffer.items[i]))
	}
	result := strings.Join(p.stringData, ",")
	p.file.Write([]byte(result))
}

func (p *PersistentChanneledCounter) counterPersistWorker(dataIn <-chan uint64) {
	timeSave := time.NewTicker(p.persistFrequency)

	for {
		select {
		case data := <-dataIn:
			p.buffer.AddItem(data)
		case <-timeSave.C:
			p.saveBuffer()
		case <-p.shutdownChannel:
			log.Println("Saving buffer end exiting")
			p.saveBuffer()
			os.Exit(0)
		}
	}
}

func (p *PersistentChanneledCounter) counterServiceWorker() {
	// Next second counter
	timeNext := time.NewTicker(5 * time.Millisecond)
	var currentCounter uint64 = 0
	currentTime := time.Now()
	bufferMirror := NewRingBuffer(p.persistSeconds)
	bufferMirror.CopyDataFrom(p.buffer.items)

	persistentChannel := make(chan uint64, p.persistSeconds)

	//Persists data to file every <persistFrequency>
	go p.counterPersistWorker(persistentChannel)

	for {
		select {
		case <-timeNext.C:
			elapsedSeconds := math.Floor(time.Since(currentTime).Seconds())
			if elapsedSeconds > 0 {
				currentTime = time.Now()
				bufferMirror.AddItem(currentCounter)
				persistentChannel <- currentCounter
				log.Printf("Current req/sec saved: %d", currentCounter)
				currentCounter = 0
			}
		case <-p.inputChannel:
			currentCounter++
		case <-p.outputChannel:
			result := bufferMirror.SummElements()
			p.outputChannel <- result
		}
	}
}

func (p *PersistentChanneledCounter) IncreaseCounter() {
	p.inputChannel <- 0
}

func (p *PersistentChanneledCounter) GetTotals() (result uint64) {
	p.outputChannel <- 0
	return <-p.outputChannel
}
