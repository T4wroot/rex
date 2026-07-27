package main

import (
	"bufio"
	"context"
	"log"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

// StreamMessage is sent per log line
type StreamMessage struct {
	Type string `json:"type"`
	ID   string `json:"id"`
	Line string `json:"line,omitempty"`
}

// Streamer handles live file tailing over WebSocket
type Streamer struct {
	allowlist *Allowlist
}

// NewStreamer creates a new Streamer
func NewStreamer(al *Allowlist) *Streamer {
	return &Streamer{allowlist: al}
}

// StreamFile tails a log file and sends lines to the WebSocket connection.
// Stops when ctx is cancelled or connection closes.
func (s *Streamer) StreamFile(ctx context.Context, conn *websocket.Conn, mu interface{ Lock(); Unlock() }, id, path string, tailLines int) {
	if !s.allowlist.IsPathAllowed(path) {
		_ = sendJSON(conn, mu, map[string]interface{}{
			"type":  "stream_error",
			"id":    id,
			"error": "path not allowed by rex policy",
		})
		return
	}

	f, err := os.Open(path)
	if err != nil {
		_ = sendJSON(conn, mu, map[string]interface{}{
			"type":  "stream_error",
			"id":    id,
			"error": err.Error(),
		})
		return
	}
	defer f.Close()

	if tailLines > 0 {
		seekToTail(f, tailLines)
	} else {
		f.Seek(0, 2)
	}

	scanner := bufio.NewScanner(f)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	log.Printf("[rex] streaming %q (id=%s)", path, id)

	for {
		select {
		case <-ctx.Done():
			_ = sendJSON(conn, mu, map[string]interface{}{
				"type": "stream_end",
				"id":   id,
			})
			return
		case <-ticker.C:
			for scanner.Scan() {
				line := scanner.Text()
				if err := sendJSON(conn, mu, StreamMessage{
					Type: "stream_line",
					ID:   id,
					Line: line,
				}); err != nil {
					log.Printf("[rex] stream send error: %v", err)
					return
				}
			}
		}
	}
}

// seekToTail moves file pointer so next read returns approximately last N lines
func seekToTail(f *os.File, n int) {
	const chunkSize = 4096
	stat, err := f.Stat()
	if err != nil {
		return
	}
	size := stat.Size()
	if size == 0 {
		return
	}
	buf := make([]byte, chunkSize)
	lineCount := 0
	pos := size
	for pos > 0 && lineCount <= n {
		readSize := int64(chunkSize)
		if pos < readSize {
			readSize = pos
		}
		pos -= readSize
		f.Seek(pos, 0)
		read, _ := f.Read(buf[:readSize])
		for i := read - 1; i >= 0; i-- {
			if buf[i] == '\n' {
				lineCount++
				if lineCount > n {
					f.Seek(pos+int64(i)+1, 0)
					return
				}
			}
		}
	}
	f.Seek(0, 0)
}
