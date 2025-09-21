package transport

import (
	"context"
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
	return SendFiles(ctx, conn, absPaths, metas)
}
