package transport

import (
	"context"
	"encoding/json"
	"time"

	"github.com/quic-go/quic-go"
)

type FileMeta struct {
	Rel    string `json:"rel"`
	Size   int64  `json:"size"`
	Sha256 string `json:"sha256,omitempty"`
}

type OfferMsg struct {
	Files []FileMeta `json:"files"`
	Total int64      `json:"total"`
}

type Hello struct {
	Label string `json:"label"`
	FP    string `json:"fp"`
}

type SASMsg struct {
	Nonce []byte `json:"nonce"`
}

func WriteJSON(ctx context.Context, s *quic.Stream, v any) error {
	if dl, ok := ctx.Deadline(); ok {
		_ = s.SetWriteDeadline(dl)
	} else {
		_ = s.SetWriteDeadline(time.Now().Add(5 * time.Second))
	}
	enc := json.NewEncoder(s)
	return enc.Encode(v)
}

func ReadJSON(ctx context.Context, s *quic.Stream, v any) error {
	if dl, ok := ctx.Deadline(); ok {
		_ = s.SetReadDeadline(dl)
	} else {
		_ = s.SetReadDeadline(time.Now().Add(5 * time.Second))
	}
	dec := json.NewDecoder(s)
	return dec.Decode(v)
}
