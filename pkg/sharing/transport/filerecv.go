package transport

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/quic-go/quic-go"
)

func ReceiveFiles(ctx context.Context, conn *quic.Conn, dstDir string, expected int) error {
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		return err
	}
	for range make([]struct{}, expected) {
		rs, err := conn.AcceptUniStream(ctx)
		if err != nil {
			return err
		}
		if err := recvOne(rs, dstDir); err != nil {
			return err
		}
	}
	return nil
}

func recvOne(rs *quic.ReceiveStream, dstDir string) error {
	var nb [2]byte
	if _, err := io.ReadFull(rs, nb[:]); err != nil {
		return err
	}
	nlen := int(binary.BigEndian.Uint16(nb[:]))
	if nlen <= 0 || nlen > 4096 {
		return fmt.Errorf("invalid name length")
	}
	name := make([]byte, nlen)
	if _, err := io.ReadFull(rs, name); err != nil {
		return err
	}
	var sb [8]byte
	if _, err := io.ReadFull(rs, sb[:]); err != nil {
		return err
	}
	size := int64(binary.BigEndian.Uint64(sb[:]))

	out := filepath.Join(dstDir, string(name))
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return err
	}
	f, err := os.Create(out)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = os.Remove(out)
		}
	}()

	if _, err := io.CopyN(f, rs, size); err != nil {
		return err
	}
	return nil
}
