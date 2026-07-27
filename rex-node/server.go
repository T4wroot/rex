package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Server is the REX WebSocket server
type Server struct {
	cfg       *Config
	allowlist *Allowlist
	executor  *Executor
	streamer  *Streamer
}

// NewServer creates the server
func NewServer(cfg *Config, al *Allowlist) *Server {
	return &Server{
		cfg:       cfg,
		allowlist: al,
		executor:  NewExecutor(al),
		streamer:  NewStreamer(al),
	}
}

// Start begins listening for connections
func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/rex", s.handleConnection)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "rex-node ok")
	})

	addr := fmt.Sprintf(":%d", s.cfg.Port)
	log.Printf("[rex] listening on %s", addr)

	if s.cfg.TLS && s.cfg.CertFile != "" && s.cfg.KeyFile != "" {
		return http.ListenAndServeTLS(addr, s.cfg.CertFile, s.cfg.KeyFile, mux)
	}
	return http.ListenAndServe(addr, mux)
}

// handleConnection upgrades HTTP to WebSocket and handles the REX session
func (s *Server) handleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[rex] upgrade error: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("[rex] new connection from %s", r.RemoteAddr)

	if !s.doHandshake(conn) {
		return
	}

	s.handleSession(conn)
	log.Printf("[rex] connection closed from %s", r.RemoteAddr)
}

// doHandshake validates the token and sends server info
func (s *Server) doHandshake(conn *websocket.Conn) bool {
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	_, data, err := conn.ReadMessage()
	if err != nil {
		log.Printf("[rex] handshake read error: %v", err)
		return false
	}

	var msg map[string]interface{}
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("[rex] handshake invalid JSON: %v", err)
		return false
	}

	if msg["type"] != "handshake" {
		log.Printf("[rex] expected handshake, got: %v", msg["type"])
		return false
	}

	token, _ := msg["token"].(string)
	if token != s.cfg.Token {
		log.Printf("[rex] invalid token — connection rejected")
		conn.WriteJSON(map[string]interface{}{
			"type":   "handshake_ack",
			"status": "error",
			"error":  "invalid token",
		})
		return false
	}

	hostname, _ := os.Hostname()
	conn.SetReadDeadline(time.Time{})

	conn.WriteJSON(map[string]interface{}{
		"type":         "handshake_ack",
		"status":       "ok",
		"node_id":      hostname,
		"os":           runtime.GOOS + "/" + runtime.GOARCH,
		"capabilities": []string{"exec", "stream", "sysinfo"},
		"protocol":     "RXP/1.0",
	})

	log.Printf("[rex] client authenticated — node ready")
	return true
}

// handleSession processes incoming RXP messages in a loop
func (s *Server) handleSession(conn *websocket.Conn) {
	var mu sync.Mutex
	streams := make(map[string]context.CancelFunc)

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("[rex] unexpected close: %v", err)
			}
			for _, cancel := range streams {
				cancel()
			}
			return
		}

		var msg map[string]interface{}
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Printf("[rex] invalid JSON: %v", err)
			continue
		}

		msgType, _ := msg["type"].(string)
		msgID, _ := msg["id"].(string)

		switch msgType {
		case "exec":
			go s.handleExec(conn, &mu, msg, msgID)

		case "stream":
			path, _ := msg["target"].(string)
			tailLines := 50
			if v, ok := msg["lines"].(float64); ok {
				tailLines = int(v)
			}
			ctx, cancel := context.WithCancel(context.Background())
			streams[msgID] = cancel
			go func() {
				s.streamer.StreamFile(ctx, conn, &mu, msgID, path, tailLines)
				delete(streams, msgID)
			}()

		case "stream_stop":
			if cancel, ok := streams[msgID]; ok {
				cancel()
				delete(streams, msgID)
			}

		case "sysinfo":
			go s.handleSysinfo(conn, &mu, msgID)

		case "ping":
			_ = sendJSON(conn, &mu, map[string]interface{}{
				"type": "pong",
				"id":   msgID,
				"ts":   time.Now().UnixMilli(),
			})

		default:
			log.Printf("[rex] unknown message type: %q", msgType)
			_ = sendJSON(conn, &mu, map[string]interface{}{
				"type":  "error",
				"id":    msgID,
				"error": fmt.Sprintf("unknown type: %s", msgType),
			})
		}
	}
}

func (s *Server) handleExec(conn *websocket.Conn, mu sync.Locker, msg map[string]interface{}, id string) {
	command, _ := msg["command"].(string)
	timeout := 30
	if v, ok := msg["timeout"].(float64); ok {
		timeout = int(v)
	}

	log.Printf("[rex] exec: %q", command)
	result := s.executor.Run(command, timeout)

	_ = sendJSON(conn, mu, map[string]interface{}{
		"type":        "exec_result",
		"id":          id,
		"exit_code":   result.ExitCode,
		"stdout":      result.Stdout,
		"stderr":      result.Stderr,
		"duration_ms": result.DurationMs,
		"error":       result.Error,
	})
}

func (s *Server) handleSysinfo(conn *websocket.Conn, mu sync.Locker, id string) {
	info := collectSysinfo()
	info["type"] = "sysinfo_result"
	info["id"] = id
	_ = sendJSON(conn, mu, info)
}

// sendJSON thread-safely sends a JSON message over WebSocket
func sendJSON(conn *websocket.Conn, mu interface{ Lock(); Unlock() }, v interface{}) error {
	mu.Lock()
	defer mu.Unlock()
	return conn.WriteJSON(v)
}
