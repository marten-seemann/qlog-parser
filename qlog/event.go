package qlog

import (
	"net"
	"time"

	"github.com/francoispqt/gojay"
)

func parseDuration(dec *gojay.Decoder) (time.Duration, error) {
	var val float64
	if err := dec.Float64(&val); err != nil {
		return 0, err
	}
	return time.Duration(val*1e6) * time.Nanosecond, nil
}

type unimplementedEvent struct{}

func (e unimplementedEvent) UnmarshalJSONObject(dec *gojay.Decoder, key string) error { return nil }
func (e unimplementedEvent) NKeys() int                                               { return 0 }

// EventConnectionStarted is the connection_started event
type EventConnectionStarted struct {
	Src, Dest net.UDPAddr
}

// UnmarshalJSONObject unmarshals the connection_started event
func (e *EventConnectionStarted) UnmarshalJSONObject(dec *gojay.Decoder, key string) error {
	switch key {
	case "src_ip":
		var s string
		if err := dec.String(&s); err != nil {
			return err
		}
		e.Src.IP = net.ParseIP(s)
	case "dst_ip":
		var s string
		if err := dec.String(&s); err != nil {
			return err
		}
		e.Src.IP = net.ParseIP(s)
	case "src_port":
		return dec.Int(&e.Src.Port)
	case "dst_port":
		return dec.Int(&e.Dest.Port)
	}
	return nil
}

// NKeys is the number of keys
func (e *EventConnectionStarted) NKeys() int { return 0 }

// EventPacketSent it the packet_sent event
type EventPacketSent struct {
	PacketType  string
	Header      Header
	IsCoalesced bool
	Frames      []Frame
	Trigger     string
}

// UnmarshalJSONObject unmarshals the packet_sent event
func (e *EventPacketSent) UnmarshalJSONObject(dec *gojay.Decoder, key string) error {
	switch key {
	case "frames":
		var frames frames
		if err := dec.Array(&frames); err != nil {
			return err
		}
		e.Frames = make([]Frame, len(frames))
		for i, f := range frames {
			e.Frames[i] = f.Frame
		}
	case "header":
		return dec.Object(&e.Header)
	case "is_coalesced":
		return dec.Bool(&e.IsCoalesced)
	case "packet_type":
		return dec.String(&e.PacketType)
	case "trigger":
		return dec.String(&e.Trigger)
	}
	return nil
}

// NKeys is the number of keys
func (e *EventPacketSent) NKeys() int { return 0 }

// EventPacketReceived it the packet_received event
type EventPacketReceived struct {
	PacketType  string
	Header      Header
	IsCoalesced bool
	Frames      []Frame
	Trigger     string
}

// UnmarshalJSONObject unmarshals the packet_received event
func (e *EventPacketReceived) UnmarshalJSONObject(dec *gojay.Decoder, key string) error {
	switch key {
	case "frames":
		var frames frames
		if err := dec.Array(&frames); err != nil {
			return err
		}
		e.Frames = make([]Frame, len(frames))
		for i, f := range frames {
			e.Frames[i] = f.Frame
		}
	case "header":
		return dec.Object(&e.Header)
	case "is_coalesced":
		return dec.Bool(&e.IsCoalesced)
	case "packet_type":
		return dec.String(&e.PacketType)
	case "trigger":
		return dec.String(&e.Trigger)
	}
	return nil
}

// NKeys is the number of keys
func (e *EventPacketReceived) NKeys() int { return 0 }

type frames []*frame

func (f *frames) UnmarshalJSONArray(dec *gojay.Decoder) error {
	frame := &frame{}
	if err := dec.Object(frame); err != nil {
		return err
	}
	if frame.Frame != nil {
		*f = append(*f, frame)
	}
	return nil
}

// EventPacketLost is the packet_lost event
type EventPacketLost struct {
	PacketType   string
	PacketNumber int64
	Trigger      string
}

// UnmarshalJSONObject unmarshals the packet_lost event
func (e *EventPacketLost) UnmarshalJSONObject(dec *gojay.Decoder, key string) error {
	switch key {
	case "packet_type":
		return dec.String(&e.PacketType)
	case "packet_number":
		return dec.Int64(&e.PacketNumber)
	case "trigger":
		return dec.String(&e.Trigger)
	}
	return nil
}

// NKeys is the number of keys
func (e *EventPacketLost) NKeys() int { return 0 }

// EventMetricsUpdated is the metrics_updated event
type EventMetricsUpdated struct {
	PTOCount         *uint32
	LatestRTT        *time.Duration
	SmoothedRTT      *time.Duration
	MinRTT           *time.Duration
	RTTVariance      *time.Duration
	CongestionWindow uint64
	BytesInFlight    uint64
}

// UnmarshalJSONObject unmarshals the metrics_updated event
func (e *EventMetricsUpdated) UnmarshalJSONObject(dec *gojay.Decoder, key string) error {
	switch key {
	case "pto_count":
		return dec.Uint32Null(&e.PTOCount)
	case "latest_rtt":
		dur, err := parseDuration(dec)
		if err != nil {
			return err
		}
		e.LatestRTT = &dur
	case "smoothed_rtt":
		dur, err := parseDuration(dec)
		if err != nil {
			return err
		}
		e.SmoothedRTT = &dur
	case "min_rtt":
		dur, err := parseDuration(dec)
		if err != nil {
			return err
		}
		e.MinRTT = &dur
	case "rtt_variance":
		dur, err := parseDuration(dec)
		if err != nil {
			return err
		}
		e.RTTVariance = &dur
	case "congestion_window":
		return dec.Uint64(&e.CongestionWindow)
	case "bytes_in_flight":
		return dec.Uint64(&e.BytesInFlight)
	}
	return nil
}

// NKeys is the number of keys
func (e *EventMetricsUpdated) NKeys() int { return 0 }
