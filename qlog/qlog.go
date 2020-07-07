package qlog

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/francoispqt/gojay"
)

// A Decoder decodes a qlog
type Decoder struct {
	eventChan chan<- Event
}

// NewDecoder creates a new decoder
func NewDecoder(c chan<- Event) *Decoder {
	return &Decoder{eventChan: c}
}

// Decode decodes the qlog
func (d *Decoder) Decode(r io.Reader) error {
	decoder := gojay.BorrowDecoder(r)
	defer decoder.Release()
	return decoder.Object(&topLevel{decoder: d})
}

type topLevel struct {
	decoder *Decoder
}

func (t topLevel) UnmarshalJSONObject(dec *gojay.Decoder, key string) error {
	switch key {
	case "traces":
		traces := &traces{decoder: t.decoder}
		if err := dec.Array(traces); err != nil {
			return err
		}
		if len(traces.traces) != 1 {
			return fmt.Errorf("Expected one Trace. Got %d", len(traces.traces))
		}
	}
	return nil
}

func (t topLevel) NKeys() int { return 0 }

type traces struct {
	traces  []Trace
	decoder *Decoder
}

func (t *traces) UnmarshalJSONArray(dec *gojay.Decoder) error {
	trace := &Trace{decoder: t.decoder}
	if err := dec.Object(trace); err != nil {
		return err
	}
	t.traces = append(t.traces, *trace)
	return nil
}

// A Trace is a qlog trac
type Trace struct {
	VantagePoint string

	commonFields commonFields
	eventFields  []string

	decoder *Decoder
}

// UnmarshalJSONObject unmarshals the trace
func (t *Trace) UnmarshalJSONObject(dec *gojay.Decoder, key string) error {
	events := events{trace: t}
	switch key {
	case "vantage_point":
		var vp vantagePoint
		if err := dec.Object(&vp); err != nil {
			return err
		}
		t.VantagePoint = vp.Type
		return nil
	case "event_fields":
		return dec.SliceString(&t.eventFields)
	case "common_fields":
		return dec.Object(&t.commonFields)
	case "events":
		return dec.Array(&events)
	}
	return nil
}

// NKeys is the number of keys
func (t Trace) NKeys() int { return 0 }

type vantagePoint struct {
	Type string
}

func (p vantagePoint) UnmarshalJSONObject(dec *gojay.Decoder, key string) error {
	if key == "type" {
		dec.String(&p.Type)
	}
	return nil
}

func (p vantagePoint) NKeys() int { return 0 }

type commonFields struct {
	ODCID         string
	ReferenceTime time.Time
}

func (f *commonFields) UnmarshalJSONObject(dec *gojay.Decoder, key string) error {
	switch key {
	case "ODCID":
		dec.AddString(&f.ODCID)
	case "reference_time":
		var s float64
		dec.AddFloat64(&s)
		f.ReferenceTime = time.Unix(0, int64(1e6*s))
	}
	return nil
}

func (f commonFields) NKeys() int { return 0 }

type events struct {
	trace *Trace
}

func (e *events) UnmarshalJSONArray(dec *gojay.Decoder) error {
	ev := Event{trace: e.trace}
	if err := dec.Array(&ev); err != nil {
		return err
	}
	e.trace.decoder.eventChan <- ev
	return nil
}

// Event is a qlog event
type Event struct {
	trace     *Trace
	counter   int
	eventName string

	Time     time.Time
	Category string
	Details  interface{}
}

// UnmarshalJSONArray unmarshals
func (e *Event) UnmarshalJSONArray(dec *gojay.Decoder) error {
	err := e.unmarshalJSONArray(dec)
	e.counter++
	return err
}

func (e *Event) unmarshalJSONArray(dec *gojay.Decoder) error {
	if e.counter >= len(e.trace.eventFields) {
		return errors.New("too many fields")
	}
	switch name := e.trace.eventFields[e.counter]; name {
	case "relative_time":
		var d float64
		if err := dec.Float64(&d); err != nil {
			return err
		}
		e.Time = e.trace.commonFields.ReferenceTime.Add(time.Duration(1e6*d) * time.Nanosecond)
	case "category":
		return dec.String(&e.Category)
	case "event":
		return dec.String(&e.eventName)
	case "data":
		var ev gojay.UnmarshalerJSONObject
		switch e.eventName {
		case "connection_started":
			ev = &EventConnectionStarted{}
		case "packet_sent":
			ev = &EventPacketSent{}
		case "packet_received":
			ev = &EventPacketReceived{}
		case "packet_lost":
			ev = &EventPacketLost{}
		case "":
			return errors.New("found data before event name")
		default:
			return dec.Object(&unimplementedEvent{})
		}
		if err := dec.Object(ev); err != nil {
			return err
		}
		e.Details = ev
	default:
		return fmt.Errorf("unknown event field: %s", name)
	}
	return nil
}

func (e *Event) String() string {
	return fmt.Sprintf("{Time: %s, Category: %s, Name: %s}", e.Time, e.Category, e.eventName)
}

type frame struct {
	processedFirstKey bool

	Frame Frame
}

func (f *frame) UnmarshalJSONObject(dec *gojay.Decoder, key string) error {
	if f.Frame != nil {
		switch f.Frame.(type) {
		case *CryptoFrame, *ConnectionCloseFrame, *NewConnectionIDFrame, *RetireConnectionIDFrame:
			return f.Frame.UnmarshalJSONObject(dec, key)
		default:
			panic("unexpected frame type")
		}
	}
	switch key {
	case "frame_type":
		if f.processedFirstKey {
			return errors.New("expected frame_type to be the first key of a frame")
		}
		var t string
		if err := dec.String(&t); err != nil {
			return err
		}
		switch t {
		case "crypto":
			f.Frame = &CryptoFrame{}
		case "connection_close":
			f.Frame = &ConnectionCloseFrame{}
		case "new_connection_id":
			f.Frame = &NewConnectionIDFrame{}
		case "retire_connection_id":
			f.Frame = &RetireConnectionIDFrame{}
		}
	}
	f.processedFirstKey = true
	return nil
}

func (f *frame) NKeys() int { return 0 }
