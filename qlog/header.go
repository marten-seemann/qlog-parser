package qlog

import (
	"github.com/francoispqt/gojay"
)

// A Header is a QUIC packet header
type Header struct {
	SrcConnID, DestConnID string
	PacketNumber          int64
	PayloadLength         uint64
	PacketSize            uint64
	Version               string
}

// UnmarshalJSONObject unmarshals the Header.
func (h *Header) UnmarshalJSONObject(dec *gojay.Decoder, key string) error {
	switch key {
	case "packet_number":
		return dec.Int64(&h.PacketNumber)
	case "payload_length":
		return dec.Uint64((&h.PayloadLength))
	case "packet_size":
		return dec.Uint64(&h.PacketSize)
	case "version":
		return dec.String(&h.Version)
	case "scid":
		return dec.String(&h.SrcConnID)
	case "dcid":
		return dec.String(&h.DestConnID)
	}
	return nil
}

// NKeys returns the number of keys.
func (h *Header) NKeys() int { return 0 }
