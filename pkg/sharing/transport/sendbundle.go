package transport

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/quic-go/quic-go"
)

func BuildOffer(absPaths []string) (OfferMsg, []FileMeta, error) {
	metas := make([]FileMeta, 0, len(absPaths))
	var total int64
	for _, abs := range absPaths {
		st, err := os.Stat(abs)
		if err != nil {
			return OfferMsg{}, nil, err
		}
		if st.IsDir() {
			continue
		}
		metas = append(metas, FileMeta{
			Rel:  filepath.Base(abs),
			Size: st.Size(),
		})
		total += st.Size()
	}
	return OfferMsg{Files: metas, Total: total}, metas, nil
}

func SendOfferAndFiles(ctx context.Context, ctrl *quic.Stream, conn *quic.Conn, absPaths []string) error {
	offer, metas, err := BuildOffer(absPaths)
	if err != nil {
		return err
	}
	if err := WriteJSON(ctx, ctrl, offer); err != nil {
		return err
	}

	// Wait for server to accept the offer
	var response map[string]string
	if err := ReadJSON(ctx, ctrl, &response); err != nil {
		return fmt.Errorf("failed to read server response: %w", err)
	}
	if response["status"] != "accepted" {
		return fmt.Errorf("server rejected offer: %s", response["status"])
	}

	// Send the files
	if err := SendFiles(ctx, conn, absPaths, metas); err != nil {
		return err
	}

	// Wait for completion confirmation
	if err := ReadJSON(ctx, ctrl, &response); err != nil {
		return fmt.Errorf("failed to read completion response: %w", err)
	}
	if response["status"] != "complete" {
		return fmt.Errorf("unexpected completion status: %s", response["status"])
	}

	// Send acknowledgment
	if err := WriteJSON(ctx, ctrl, map[string]string{"status": "acknowledged"}); err != nil {
		return fmt.Errorf("failed to send acknowledgment: %w", err)
	}

	return nil
}
