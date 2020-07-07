package qlog

import (
	"net"

	"github.com/francoispqt/gojay"
)

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
