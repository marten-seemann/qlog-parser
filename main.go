package main

import (
	"bufio"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	_ "net/http/pprof"

	"github.com/marten-seemann/qlog-parser/qlog"
)

const numWorkers = 128

func main() {
	role := flag.String("role", "", "client / server / <emtpy>")
	flag.Parse()

	if len(flag.Args()) < 1 {
		log.Fatalf("No qlog directory given.")
	}

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	qlogDir := flag.Arg(0)
	if err := process(qlogDir, *role); err != nil {
		log.Fatalf("Processing failed: %s", err)
	}
}

func process(dir string, role string) error {
	var onlyClient, onlyServer bool
	if role == "client" {
		onlyClient = true
	}
	if role == "server" {
		onlyServer = true
	}

	sem := make(chan struct{}, numWorkers)
	var wg sync.WaitGroup
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() ||
			(!strings.HasSuffix(info.Name(), "qlog.gz") && !strings.HasSuffix(info.Name(), ".qlog")) {
			return nil
		}
		if onlyClient && !strings.Contains(info.Name(), "client") {
			return nil
		}
		if onlyServer && !strings.Contains(info.Name(), "server") {
			return nil
		}

		sem <- struct{}{}
		wg.Add(1)

		go func() {
			defer func() {
				<-sem
				wg.Done()
			}()
			if err := processQlog(path); err != nil {
				if err == io.EOF {
					err = nil
				}
				log.Printf("Parsing %s failed: %s\n", path, err)
			}
		}()
		return nil
	})
	if err != nil {
		return err
	}
	wg.Wait()
	return nil
}

func processQlog(path string) error {
	// log.Printf("Processing %s\n", path)
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	var src io.Reader
	if strings.HasSuffix(path, ".gz") {
		gz, err := gzip.NewReader(bufio.NewReader(f))
		if err != nil {
			return err
		}
		src = gz
	} else {
		src = bufio.NewReader(f)
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
	return d.Decode(src)
}
