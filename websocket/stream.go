// Package websocket provides WebSocket transport support for JSON-RPC
// 2.0.
package websocket

import (
	"io"
	"bytes"
	"bufio"

	"github.com/gorilla/websocket"
	"github.com/sourcegraph/jsonrpc2"
)

// A ObjectStream is a jsonrpc2.ObjectStream that uses a WebSocket to
// send and receive JSON-RPC 2.0 objects.
type ObjectStream struct {
	conn *websocket.Conn
	codec jsonrpc2.ObjectCodec
}

// NewObjectStream creates a new jsonrpc2.ObjectStream for sending and
// receiving JSON-RPC 2.0 objects over a WebSocket.
func NewObjectStream(conn *websocket.Conn, codec jsonrpc2.ObjectCodec) ObjectStream {
	return ObjectStream{conn: conn, codec: codec}
}

// WriteObject implements jsonrpc2.ObjectStream.
func (t ObjectStream) WriteObject(obj interface{}) (err error) {
	stream, err := t.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return
	}

	err = t.codec.WriteObject(stream, obj)
	if err != nil {
		return
	}

	err = stream.Close()
	if err != nil {
		return
	}

	return
}

// ReadObject implements jsonrpc2.ObjectStream.
func (t ObjectStream) ReadObject(v interface{}) (err error) {
	// messageType, p , err := t.conn.ReadMessage()
	_, p , err := t.conn.ReadMessage()
	if err != nil {
		return;
	}

	rd := bytes.NewReader(p)
	r := bufio.NewReader(rd)
	err = t.codec.ReadObject(r, v)
	if err != nil {
		return
	}

	if e, ok := err.(*websocket.CloseError); ok {
		if e.Code == websocket.CloseAbnormalClosure && e.Text == io.ErrUnexpectedEOF.Error() {
			// Suppress a noisy (but harmless) log message by
			// unwrapping this error.
			err = io.ErrUnexpectedEOF
			return
		}
	}

	return
}

// Close implements jsonrpc2.ObjectStream.
func (t ObjectStream) Close() error {
	return t.conn.Close()
}