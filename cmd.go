package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"

	"github.com/alexcesaro/log"
	"github.com/alexcesaro/log/golog"
	"github.com/conformal/go-flags"
)
import _ "net/http/pprof"

// version gets replaced during build
var version string = "dev"

// Options contains the flag options
type Options struct {
	Bind    string `long:"bind" description:"Host and port to listen on." default:"0.0.0.0:6697"`
	Pprof   int    `long:"pprof" description:"Enable pprof http server for profiling."`
	Verbose []bool `short:"v" long:"verbose" description:"Show verbose logging."`
	Version bool   `long:"version"`
}

var logLevels = []log.Level{
	log.Warning,
	log.Info,
	log.Debug,
}

func fail(code int, format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(code)
}

func main() {
	options := Options{}
	parser := flags.NewParser(&options, flags.Default)
	_, err := parser.Parse()
	if err != nil {
		os.Exit(1)
		return
	}

	if options.Version {
		fmt.Println(version)
		os.Exit(0)
	}

	if options.Pprof != 0 {
		go func() {
			fmt.Println(http.ListenAndServe(fmt.Sprintf("localhost:%d", options.Pprof), nil))
		}()
	}

	// Figure out the log level
	numVerbose := len(options.Verbose)
	if numVerbose > len(logLevels) {
		numVerbose = len(logLevels) - 1
	}

	logLevel := logLevels[numVerbose]
	SetLogger(golog.New(os.Stderr, logLevel))

	host, err := net.Listen("tcp", options.Bind)
	if err != nil {
		fail(4, "Failed to listen on socket: %v\n", err)
	}
	defer host.Close()

	s := &Server{}
	s.Start(host)

	fmt.Printf("Listening for connections on %v\n", host.Addr().String())

	// Construct interrupt handler
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	<-sig // Wait for ^C signal
	fmt.Fprintln(os.Stderr, "Interrupt signal detected, shutting down.")
	os.Exit(0)
}
