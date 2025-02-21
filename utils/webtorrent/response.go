package webtorrent

import "github.com/hertz-contrib/websocket"

type errorResponse struct {
	FailureReason string `json:"failure reason,omitempty"`
}

func ResponseErr(conn *websocket.Conn, err error) error {
	return conn.WriteJSON(errorResponse{FailureReason: err.Error()})
}
