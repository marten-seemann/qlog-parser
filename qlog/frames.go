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
