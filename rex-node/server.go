package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
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

func NewServer(cfg *Config, al *Allowlist) *Server {
	return &Server{
		cfg:       cfg,
		allowlist: al,
		executor:  NewExecutor(al),
		streamer:  NewStreamer(al),
	}
}

func (s *Server) Start() error {
	go s.StartTCP()

	mux := http.NewServeMux()
	mux.HandleFunc("/rex", s.handleConnection)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "rex-node ok")
	})

	addr := fmt.Sprintf(":%d", s.cfg.Port)
	log.Printf("[rex] WebSocket listening on %s", addr)

	if s.cfg.TLS && s.cfg.CertFile != "" && s.cfg.KeyFile != "" {
		return http.ListenAndServeTLS(addr, s.cfg.CertFile, s.cfg.KeyFile, mux)
	}
	return http.ListenAndServe(addr, mux)
}

func (s *Server) StartTCP() {
	tcpAddr := fmt.Sprintf(":%d", s.cfg.TCPPort)
	listener, err := net.Listen("tcp", tcpAddr)
	if err != nil {
		log.Printf("[rex-tcp] failed to listen on %s: %v", tcpAddr, err)
		return
	}
	defer listener.Close()

	log.Printf("[rex-tcp] RXP/2.0 Binary Server listening on %s", tcpAddr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("[rex-tcp] accept error: %v", err)
			continue
		}
		go s.handleTCPConn(conn)
	}
}

func (s *Server) handleTCPConn(conn net.Conn) {
	defer conn.Close()
	remoteAddr := conn.RemoteAddr().String()
	log.Printf("[rex-tcp] client connected: %s", remoteAddr)

	var connMu sync.Mutex
	ptyMgr := NewPTYManager(conn, &connMu)
	defer ptyMgr.CloseAll()

	authenticated := false

	for {
		frame, err := ReadFrame(conn)
		if err != nil {
			if err != io.EOF {
				log.Printf("[rex-tcp] read error (%s): %v", remoteAddr, err)
			}
			return
		}

		if !authenticated && frame.Opcode != OpAuthHandshake {
			log.Printf("[rex-tcp] unauthenticated frame from %s", remoteAddr)
			connMu.Lock()
			_ = WriteFrame(conn, OpError, frame.StreamID, []byte("unauthenticated"))
			connMu.Unlock()
			return
		}

		switch frame.Opcode {
		case OpAuthHandshake:
			token := string(frame.Payload)
			if token != s.cfg.Token {
				log.Printf("[rex-tcp] auth failed for %s", remoteAddr)
				connMu.Lock()
				_ = WriteFrame(conn, OpError, frame.StreamID, []byte("invalid token"))
				connMu.Unlock()
				return
			}
			authenticated = true
			connMu.Lock()
			_ = WriteFrame(conn, OpAuthHandshake, frame.StreamID, []byte("RXP/2.0 OK"))
			connMu.Unlock()
			log.Printf("[rex-tcp] client authenticated (%s)", remoteAddr)

		case OpPTYSpawn:
			cols, rows := uint16(80), uint16(24)
			if len(frame.Payload) >= 4 {
				cols = binary.BigEndian.Uint16(frame.Payload[0:2])
				rows = binary.BigEndian.Uint16(frame.Payload[2:4])
			}
			if err := ptyMgr.Spawn(frame.StreamID, cols, rows); err != nil {
				connMu.Lock()
				_ = WriteFrame(conn, OpError, frame.StreamID, []byte(err.Error()))
				connMu.Unlock()
			} else {
				connMu.Lock()
				_ = WriteFrame(conn, OpPTYSpawn, frame.StreamID, []byte("PTY_SPAWNED"))
				connMu.Unlock()
			}

		case OpPTYData:
			_ = ptyMgr.WriteStdin(frame.StreamID, frame.Payload)

		case OpPTYResize:
			if len(frame.Payload) >= 4 {
				cols := binary.BigEndian.Uint16(frame.Payload[0:2])
				rows := binary.BigEndian.Uint16(frame.Payload[2:4])
				_ = ptyMgr.Resize(frame.StreamID, cols, rows)
			}

		case OpPTYClose:
			ptyMgr.Close(frame.StreamID)

		case OpNativeFileOp:
			resBytes := HandleNativeFileOp(frame.Payload)
			connMu.Lock()
			_ = WriteFrame(conn, OpNativeFileOp, frame.StreamID, resBytes)
			connMu.Unlock()

		case OpNativeSysInfo:
			info := collectSysinfo()
			infoBytes, _ := json.Marshal(info)
			connMu.Lock()
			_ = WriteFrame(conn, OpNativeSysInfo, frame.StreamID, infoBytes)
			connMu.Unlock()

		case OpPing:
			connMu.Lock()
			_ = WriteFrame(conn, OpPong, frame.StreamID, frame.Payload)
			connMu.Unlock()

		case OpAgentGuide:
			connMu.Lock()
			_ = WriteFrame(conn, OpAgentGuide, frame.StreamID, []byte(REXProtocolRule))
			connMu.Unlock()

		default:
			log.Printf("[rex-tcp] unknown opcode: 0x%02x", frame.Opcode)
		}
	}
}

func (s *Server) handleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[rex] upgrade error: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("[rex] client connected: %s", r.RemoteAddr)

	if !s.doHandshake(conn) {
		return
	}

	s.handleSession(conn, r.RemoteAddr)
	log.Printf("[rex] client disconnected: %s", r.RemoteAddr)
}

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
		log.Printf("[rex] auth failed: invalid token")
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

	log.Printf("[rex] client authenticated (RXP/1.0 persistent session active)")
	return true
}

func (s *Server) handleSession(conn *websocket.Conn, remoteAddr string) {
	var mu sync.Mutex
	streams := make(map[string]context.CancelFunc)

	// Set Pong handler to keep persistent channel alive
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Time{})
		return nil
	})

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				log.Printf("[rex] session ended gracefully by client (%s)", remoteAddr)
			} else if websocket.IsUnexpectedCloseError(err) {
				log.Printf("[rex] connection dropped: %v", err)
			}
			for _, cancel := range streams {
				cancel()
			}
			return
		}

		var msg map[string]interface{}
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Printf("[rex] invalid JSON frame: %v", err)
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
			log.Printf("[rex] unknown frame type: %q", msgType)
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

func sendJSON(conn *websocket.Conn, mu interface{ Lock(); Unlock() }, v interface{}) error {
	mu.Lock()
	defer mu.Unlock()
	return conn.WriteJSON(v)
}
