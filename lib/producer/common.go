package producer

import (
	"bufio"
	"compress/bzip2"
	"compress/gzip"
	"errors"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io"
	"math/rand"
	"net"
	"os"
	"plane.watch/lib/tracker"
	"strings"
	"sync"
	"time"
)

const (
	cmdExit = 1

	Avr = iota
	Beast
	Sbs1
)

type (
	Producer struct {
		tracker.FrameSource
		producerType int

		log zerolog.Logger

		out chan tracker.FrameEvent

		cmdChan chan int

		splitter                      bufio.SplitFunc
		beastDelay, keepAliveRepeater bool

		run func()

		stats struct {
			avr, beast, sbs1 prometheus.Counter
		}

		hasFetcher, fetcherConnected bool

		repeater *keepAliveRepeater
	}

	Option func(*Producer)
)

func New(opts ...Option) *Producer {
	p := &Producer{
		FrameSource: tracker.FrameSource{
			OriginIdentifier: "",
			Name:             "",
			Tag:              "",
			RefLat:           nil,
			RefLon:           nil,
		},
		out:     make(chan tracker.FrameEvent, 100),
		cmdChan: make(chan int),
		run: func() {
			println("You did not specify any sources")
			os.Exit(1)
		},
	}
	p.log = log.With().Logger()

	for _, opt := range opts {
		opt(p)
	}

	if "" == p.Name {
		p.Name = p.OriginIdentifier
	}
	if "" == p.Name {
		p.Name = p.Tag
	}
	if "" == p.Name {
		p.Name = producerType(p.producerType)
	}
	p.log = log.With().
		Str("Name", p.Name).
		Str("ProducerType", producerType(p.producerType)).
		Logger()

	if p.keepAliveRepeater {
		p.log.Debug().Msg("Setting up repeater")
		p.repeater = newKeepAliveRepeater()
		go p.repeater.processor(p)
	}

	return p
}

func producerType(in int) string {
	switch in {
	case Avr:
		return "AVR"
	case Beast:
		return "Beast"
	case Sbs1:
		return "SBS1"
	default:
		return "Unknown"
	}
}

// Producer.New(WithFetcher(host, port), WithType(Producer.Avr), WithRefLatLon(lat, lon))

func WithListener(host, port string) Option {
	return func(p *Producer) {
		p.run = func() {
			addr := net.JoinHostPort(host, port)
			ln, err := net.Listen("tcp", addr)
			if err != nil {
				// handle error
				p.log.Error().Err(err).Str("host:port", addr).Msg("Failed to listen")
			}
			for {
				conn, errConn := ln.Accept()
				if errConn != nil {
					// handle error
					p.log.Error().Err(errConn).Msg("Failed to accept a connection")
				}

				go func(c net.Conn) {
					scan := bufio.NewScanner(c)
					scan.Split(p.splitter)
					errRead := p.readFromScanner(scan)
					if nil != errRead {
						p.log.Error().Err(errRead).Msg("No more reading")
					}
				}(conn)
			}
		}
	}
}

func WithSourceTag(tag string) Option {
	return func(p *Producer) {
		p.FrameSource.Tag = tag
	}
}

func WithFetcher(host, port string) Option {
	hp := net.JoinHostPort(host, port)
	return func(p *Producer) {
		p.hasFetcher = true
		p.FrameSource.OriginIdentifier = hp
		p.run = func() {
			p.addInfo("Fetching From Host: %s:%s", host, port)
			p.fetcher(host, port, func(conn net.Conn) error {
				scan := bufio.NewScanner(conn)
				scan.Split(p.splitter)
				return p.readFromScanner(scan)
			})
		}
	}
}

func WithOriginName(name string) Option {
	return func(p *Producer) {
		p.FrameSource.Name = name
	}
}

func WithFiles(filePaths []string) Option {
	return func(p *Producer) {
		p.run = func() {
			p.readFiles(filePaths, func(reader io.Reader, fileName string) error {
				scanner := bufio.NewScanner(reader)
				p.FrameSource.OriginIdentifier = "file://" + fileName
				return p.readFromScanner(scanner)
			})
		}
	}
}

func WithBeastDelay(beastDelay bool) Option {
	return func(p *Producer) {
		p.beastDelay = beastDelay
	}
}

func WithType(producerType int) Option {
	return func(p *Producer) {
		switch producerType {
		case Avr, Sbs1:
			p.producerType = producerType
			p.splitter = bufio.ScanLines
		case Beast:
			p.producerType = producerType
			p.splitter = ScanBeast
		default:
			p.log.Error().Msgf("Unknown Producer Type")
		}
	}
}

func WithPrometheusCounters(avr, beast, sbs1 prometheus.Counter) Option {
	return func(p *Producer) {
		p.stats.avr = avr
		p.stats.beast = beast
		p.stats.sbs1 = sbs1
	}
}

func (p *Producer) readFromScanner(scan *bufio.Scanner) error {
	scan.Split(p.splitter)

	switch p.producerType {
	case Avr:
		return p.avrScanner(scan)
	case Sbs1:
		return p.sbsScanner(scan)
	case Beast:
		return p.beastScanner(scan)
	default:
		return errors.New("unknown Producer type")
	}
}

// WithReferenceLatLon sets up the reference lat/lon for decoding surface position messages
func WithReferenceLatLon(lat, lon float64) Option {
	return func(p *Producer) {
		p.log.Debug().Float64("lat", lat).Float64("lon", lon).Msg("With Reference Lat/Lon")
		p.RefLat = &lat
		p.RefLon = &lon
	}
}
func WithKeepAliveRepeater() Option {
	return func(p *Producer) {
		p.keepAliveRepeater = true
	}
}

func (p *Producer) String() string {
	return p.Name
}

func (p *Producer) Listen() chan tracker.FrameEvent {
	go p.run()
	return p.out
}

func (p *Producer) addFrame(f tracker.Frame, s *tracker.FrameSource) {
	fe := tracker.NewFrameEvent(f, s)
	if p.keepAliveRepeater {
		// update the repeater for this listFrames
		p.repeater.chanFrame <- fe
	}
	p.AddEvent(fe)
}

func (p *Producer) addDebug(sfmt string, v ...interface{}) {
	p.log.Debug().Str("section", p.Name).Msgf(sfmt, v...)
}

func (p *Producer) addInfo(sfmt string, v ...interface{}) {
	p.log.Info().Str("section", p.Name).Msgf(sfmt, v...)
}

func (p *Producer) addError(err error) {
	p.log.Error().Str("section", p.Name).Err(err).Send()
}

func (p *Producer) HealthCheck() bool {
	if p.hasFetcher {
		return p.fetcherConnected
	}
	return true
}

func (p *Producer) HealthCheckName() string {
	return p.Name
}

func (p *Producer) Stop() {
	p.cmdChan <- cmdExit
}

func (p *Producer) AddEvent(e tracker.FrameEvent) {
	defer func() {
		if r := recover(); nil != r {
			p.log.Error().Interface("recover", r).Msg("Failed to add event")
		}
	}()
	p.out <- e
}

func (p *Producer) Cleanup() {
	defer func() { recover() }()
	close(p.out)
}

func (p *Producer) readFiles(dataFiles []string, read func(io.Reader, string) error) {
	var err error
	var inFile *os.File
	var gzipFile *gzip.Reader
	go func() {
		for _, inFileName := range dataFiles {
			log.Debug().Str("FileName", inFileName).Msg("Loading contents...")
			p.FrameSource.OriginIdentifier = "file://" + inFileName
			inFile, err = os.Open(inFileName)
			if err != nil {
				p.addError(fmt.Errorf("failed to open file {%s}: %s", inFileName, err))
				continue
			}

			isGzip := strings.ToLower(inFileName[len(inFileName)-2:]) == "gz"
			isBzip2 := strings.ToLower(inFileName[len(inFileName)-3:]) == "bz2"
			log.Debug().
				Str("FileName", inFileName).
				Bool("Is Gzip", isGzip).
				Bool("Is Bzip2", isBzip2).
				Bool("Is Plain", !isBzip2 && !isGzip).
				Msg("Format")

			if isGzip {
				gzipFile, err = gzip.NewReader(inFile)
				if nil != err {
					log.Error().Err(err).Str("file", inFileName).Msg("Failed to open file")
				}
				err = read(gzipFile, inFileName)
			} else if isBzip2 {
				bzip2File := bzip2.NewReader(inFile)
				err = read(bzip2File, inFileName)
			} else {
				err = read(inFile, inFileName)
			}
			if nil != err {
				p.addError(err)
			}
			_ = inFile.Close()
			log.Debug().
				Str("FileName", inFileName).
				Msg("Finished with file")
		}
		log.Debug().Msg("Done loading contents from files")
		p.Cleanup()
	}()

	go func() {
		for cmd := range p.cmdChan {
			switch cmd {
			case cmdExit:
				return
			}
		}
	}()
}

func (p *Producer) fetcher(host, port string, read func(net.Conn) error) {
	var conn net.Conn
	var wLock sync.RWMutex
	working := true

	isWorking := func() bool {
		wLock.RLock()
		defer wLock.RUnlock()
		return working
	}

	go func() {
		var backOff = time.Second
		var err error
		for isWorking() {
			p.addDebug("Connecting...")
			wLock.Lock()
			conn, err = net.Dial("tcp", net.JoinHostPort(host, port))
			wLock.Unlock()
			if nil != err {
				p.fetcherConnected = false
				p.addError(err)
				time.Sleep(backOff)
				backOff = backOff*2 + ((time.Duration(rand.Intn(20)) * time.Millisecond * 100) - time.Second)
				if backOff > time.Minute {
					backOff = time.Minute
				}
				continue
			}
			p.addDebug("Connected!")
			backOff = time.Second
			p.fetcherConnected = true

			if err = read(conn); nil != err {
				p.addError(err)
			}
		}
		p.addDebug("Done with Producer %s", p)
		p.Cleanup()
	}()

	go func() {
		for cmd := range p.cmdChan {
			switch cmd {
			case cmdExit:
				p.addDebug("Got Cmd Exit")
				wLock.Lock()
				working = false
				if nil != conn {
					_ = conn.Close()
				}
				wLock.Unlock()
				return
			}
		}
	}()
}
