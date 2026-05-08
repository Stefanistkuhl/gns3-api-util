package handlers

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/0xveya/gns3util/pkg/web/helpers"
)

func writeRangeObject(w http.ResponseWriter, r *http.Request, obj *DownloadObject) error {
	rangeHeader := r.Header.Get("Range")
	if rangeHeader == "" {
		return nil
	}

	ranges, err := parseRange(rangeHeader, obj.Size)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidRange, err)
	}

	if len(ranges) != 1 {
		w.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", obj.Size))
		return ErrMultipleRangesNotSupported
	}

	ra := ranges[0]

	srcFile, err := os.Open(obj.Path) // #nosec G304
	if err != nil {
		return fmt.Errorf("%w: %w", ErrOpenFile, err)
	}
	defer srcFile.Close()

	if _, err := srcFile.Seek(ra.start, io.SeekStart); err != nil {
		return fmt.Errorf("%w: %w", ErrSeekFile, err)
	}

	w.Header().Set("Content-Range",
		fmt.Sprintf("bytes %d-%d/%d", ra.start, ra.start+ra.length-1, obj.Size),
	)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", ra.length))
	w.WriteHeader(http.StatusPartialContent)

	if _, err := io.CopyN(w, srcFile, ra.length); err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("%w: %w", ErrCopyFile, err)
	}

	return nil
}

func (f *FilestoreHandlers) writeRangeObjectError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidRange):
		helpers.WriteAPIError(w, "invalid range header", helpers.ErrCodeInvalidInput, err.Error(), http.StatusBadRequest)

	case errors.Is(err, ErrMultipleRangesNotSupported):
		helpers.WriteAPIError(w, "multiple ranges not supported", helpers.ErrCodeInternal, err.Error(), http.StatusRequestedRangeNotSatisfiable)

	case errors.Is(err, ErrOpenFile),
		errors.Is(err, ErrSeekFile),
		errors.Is(err, ErrCopyFile):
		f.Logger.Error("file IO error", "err", err)
		helpers.WriteAPIError(w, "file processing error", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)

	default:
		f.Logger.Error("unexpected error", "err", err)
		helpers.WriteAPIError(w, "internal error", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
	}
}
