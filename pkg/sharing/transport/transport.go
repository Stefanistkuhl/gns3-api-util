package transport

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"time"
)

const bufferSize = 1024 * 1024 // 1MB buffer for file transfers

// SendFile sends a file to the specified address
func SendFile(ctx context.Context, path string, addr string) error {
	// Connect to the receiver
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect to receiver: %w", err)
	}
	defer conn.Close()

	// Open the file
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Get file info for size
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	// Send file size (8 bytes)
	err = binary.Write(conn, binary.BigEndian, fileInfo.Size())
	if err != nil {
		return fmt.Errorf("failed to send file size: %w", err)
	}

	// Send file name
	fileName := filepath.Base(path)
	err = binary.Write(conn, binary.BigEndian, int32(len(fileName)))
	if err != nil {
		return fmt.Errorf("failed to send file name length: %w", err)
	}
	_, err = conn.Write([]byte(fileName))
	if err != nil {
		return fmt.Errorf("failed to send file name: %w", err)
	}

	// Send file data
	_, err = io.CopyN(conn, file, fileInfo.Size())
	if err != nil {
		return fmt.Errorf("failed to send file data: %w", err)
	}

	return nil
}

// ReceiveFile starts a server to receive a file
func ReceiveFile(ctx context.Context, port int, outputDir string) (string, error) {
	// Create a TCP listener
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return "", fmt.Errorf("failed to start listener: %w", err)
	}
	defer listener.Close()

	// Handle incoming connection
	conn, err := listener.Accept()
	if err != nil {
		return "", fmt.Errorf("failed to accept connection: %w", err)
	}
	defer conn.Close()

	// Read file size (8 bytes)
	var fileSize int64
	err = binary.Read(conn, binary.BigEndian, &fileSize)
	if err != nil {
		return "", fmt.Errorf("failed to read file size: %w", err)
	}

	// Read file name length (4 bytes)
	var nameLen int32
	err = binary.Read(conn, binary.BigEndian, &nameLen)
	if err != nil {
		return "", fmt.Errorf("failed to read file name length: %w", err)
	}

	// Read file name
	fileNameBuf := make([]byte, nameLen)
	_, err = io.ReadFull(conn, fileNameBuf)
	if err != nil {
		return "", fmt.Errorf("failed to read file name: %w", err)
	}
	fileName := string(fileNameBuf)

	// Create output file
	outputPath := filepath.Join(outputDir, fileName)
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return "", fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	// Receive file data
	_, err = io.CopyN(outputFile, conn, fileSize)
	if err != nil {
		return "", fmt.Errorf("failed to receive file data: %w", err)
	}

	abspath, err := filepath.Abs(outputPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	return abspath, nil
}

// GeneratePort generates a random port number in the dynamic/private range
func GeneratePort() int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(16383) + 49152 // 49152-65535
}

// GetLocalIP returns the first non-loopback IP address of the machine
func GetLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", fmt.Errorf("no non-loopback IP address found")
}
