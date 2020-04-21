package main

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	_ "net/http/pprof"

	"github.com/marten-seemann/qlog-parser/qlog"
)

const numWorkers = 128

func main() {
	if len(os.Args) <= 1 {
		log.Fatalf("No qlog directory given.")
	}

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	qlogDir := os.Args[1]
	if err := process(qlogDir); err != nil {
		log.Fatalf("Processing failed: %s", err)
	}
}

func process(dir string) error {
	sem := make(chan struct{}, numWorkers)
	var done bool
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if done {
			return nil
		}
		if info.IsDir() || !strings.HasSuffix(info.Name(), "qlog.gz") {
			return nil
		}

		sem <- struct{}{}

		go func() {
			// done = true
			err = processQlog(path)
			if err == io.EOF {
				err = nil
			}
			if err != nil {
				log.Printf("Parsing %s failed: %s\n", path, err)
			}
			<-sem
		}()
		return nil
	})
}

func processQlog(path string) error {
	// log.Printf("Processing %s\n", path)
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	eventChan := make(chan qlog.Event, 100)
	go func() {
		for e := range eventChan {
			var frames []qlog.Frame
			switch ev := e.Details.(type) {
			case *qlog.EventPacketReceived:
				frames = ev.Frames
			case *qlog.EventPacketSent:
				frames = ev.Frames
			case *qlog.EventConnectionStarted:
				// fmt.Printf("Connection from %s to %s.\n", ev.Src.String(), ev.Dest.String())
				continue
			default:
				continue
			}
			if len(frames) == 0 {
				continue
			}
			for _, f := range frames {
				if ccf, ok := f.(*qlog.ConnectionCloseFrame); ok {
					if (ccf.ErrorSpace == "application" && ccf.RawErrorCode == 0) ||
						(ccf.ErrorSpace == "transport" && ccf.RawErrorCode == 0xc) ||
						(ccf.ErrorSpace == "transport" && ccf.RawErrorCode == 0x15a) ||
						(ccf.ErrorSpace == "transport" && ccf.RawErrorCode == 0x12a) {
						continue
					}
					fmt.Printf("%s: %#v\n", path, ccf)
				}
			}
		}
	}()
	d := qlog.NewDecoder(eventChan)
	return d.Decode(bufio.NewReader(gz))
}
