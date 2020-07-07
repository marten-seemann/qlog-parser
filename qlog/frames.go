package qlog

import "github.com/francoispqt/gojay"

// Frame is a qlog frame
type Frame interface {
	UnmarshalJSONObject(*gojay.Decoder, string) error
}

// CryptoFrame is a CRYPTO frame
type CryptoFrame struct {
	Offset uint64
	Length uint64
}

// UnmarshalJSONObject unmarshals the CRYPTO frmae
func (f *CryptoFrame) UnmarshalJSONObject(dec *gojay.Decoder, key string) error {
	switch key {
	case "offset":
		return dec.Uint64(&f.Offset)
	case "length":
		return dec.Uint64(&f.Length)
	}
	return nil
}

// ConnectionCloseFrame is a CONNECTION_CLOSE frame
type ConnectionCloseFrame struct {
	ErrorSpace   string
	RawErrorCode uint64
	Reason       string
}

// UnmarshalJSONObject unmarshals the CONNECTION_CLOSE frame
func (f *ConnectionCloseFrame) UnmarshalJSONObject(dec *gojay.Decoder, key string) error {
	switch key {
	case "error_space":
		return dec.String(&f.ErrorSpace)
	case "raw_error_code":
		return dec.Uint64(&f.RawErrorCode)
	case "reason":
		return dec.String(&f.Reason)
	}
	return nil
}

// NewConnectionIDFrame is a NEW_CONNECTION_ID frame
type NewConnectionIDFrame struct {
	SequenceNumber      uint64
	RetirePriorTo       uint64
	StatelessResetToken string
	ConnectionID        string
}

// UnmarshalJSONObject unmarshals the NEW_CONNECTION_ID frame
func (f *NewConnectionIDFrame) UnmarshalJSONObject(dec *gojay.Decoder, key string) error {
	switch key {
	case "sequence_number":
		return dec.Uint64(&f.SequenceNumber)
	case "retire_prior_to":
		return dec.Uint64(&f.RetirePriorTo)
	case "connection_id":
		return dec.String(&f.ConnectionID)
	case "stateless_reset_token":
		return dec.String(&f.StatelessResetToken)
	}
	return nil
}

// RetireConnectionIDFrame is RETIRE_CONNECTION_ID frame
type RetireConnectionIDFrame struct {
	SequenceNumber uint64
}

// UnmarshalJSONObject unmarshals the RETIRE_CONNECTION_ID frame
func (f *RetireConnectionIDFrame) UnmarshalJSONObject(dec *gojay.Decoder, key string) error {
	if key == "sequence_number" {
		return dec.Uint64(&f.SequenceNumber)
	}
	return nil
}
