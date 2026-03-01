package ws

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"io"
	"net"
	"net/http"
	"strings"
)

const (
	wsGUID       = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	sendBufSize  = 256
)

// client represents a connected WebSocket client.
type client struct {
	hub  *Hub
	conn net.Conn
	send chan []byte
}

// newClient upgrades the HTTP connection to a WebSocket connection.
func newClient(h *Hub, w http.ResponseWriter, r *http.Request) (*client, error) {
	key := r.Header.Get("Sec-Websocket-Key")
	if key == "" {
		return nil, io.ErrUnexpectedEOF
	}

	// Compute accept key.
	sum := sha1.Sum([]byte(key + wsGUID))
	accept := base64.StdEncoding.EncodeToString(sum[:])

	hj, ok := w.(http.Hijacker)
	if !ok {
		return nil, io.ErrUnexpectedEOF
	}
	conn, buf, err := hj.Hijack()
	if err != nil {
		return nil, err
	}

	// Write handshake response.
	resp := "HTTP/1.1 101 Switching Protocols\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Accept: " + accept + "\r\n\r\n"
	if _, err := io.WriteString(conn, resp); err != nil {
		conn.Close()
		return nil, err
	}
	_ = buf

	return &client{hub: h, conn: conn, send: make(chan []byte, sendBufSize)}, nil
}

// writePump pumps messages from the send channel to the WebSocket connection.
func (c *client) writePump() {
	defer c.conn.Close()
	for msg := range c.send {
		if err := writeTextFrame(c.conn, msg); err != nil {
			return
		}
	}
}

// readPump reads and discards incoming frames until the connection closes.
func (c *client) readPump() {
	defer c.conn.Close()
	r := bufio.NewReader(c.conn)
	for {
		// Read two header bytes.
		header := make([]byte, 2)
		if _, err := io.ReadFull(r, header); err != nil {
			return
		}
		// fin := header[0]&0x80 != 0
		opcode := header[0] & 0x0f
		masked := header[1]&0x80 != 0
		payloadLen := int64(header[1] & 0x7f)

		switch payloadLen {
		case 126:
			var ext [2]byte
			if _, err := io.ReadFull(r, ext[:]); err != nil {
				return
			}
			payloadLen = int64(binary.BigEndian.Uint16(ext[:]))
		case 127:
			var ext [8]byte
			if _, err := io.ReadFull(r, ext[:]); err != nil {
				return
			}
			payloadLen = int64(binary.BigEndian.Uint64(ext[:]))
		}

		var maskKey [4]byte
		if masked {
			if _, err := io.ReadFull(r, maskKey[:]); err != nil {
				return
			}
		}

		payload := make([]byte, payloadLen)
		if _, err := io.ReadFull(r, payload); err != nil {
			return
		}
		if masked {
			for i := range payload {
				payload[i] ^= maskKey[i%4]
			}
		}

		// Handle close frame (opcode 8).
		if opcode == 8 {
			return
		}
	}
}

// writeTextFrame writes a WebSocket text frame with the given payload.
func writeTextFrame(w io.Writer, payload []byte) error {
	n := len(payload)
	var header []byte
	if n <= 125 {
		header = []byte{0x81, byte(n)}
	} else if n <= 65535 {
		header = []byte{0x81, 126, byte(n >> 8), byte(n)}
	} else {
		header = make([]byte, 10)
		header[0] = 0x81
		header[1] = 127
		binary.BigEndian.PutUint64(header[2:], uint64(n))
	}
	if _, err := w.Write(header); err != nil {
		return err
	}
	_, err := w.Write(payload)
	return err
}

// acceptKey computes the Sec-WebSocket-Accept header value.
func acceptKey(key string) string {
	_ = strings.TrimSpace // used for import
	sum := sha1.Sum([]byte(key + wsGUID))
	return base64.StdEncoding.EncodeToString(sum[:])
}
