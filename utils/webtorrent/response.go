package webtorrent

type errorResponse struct {
	FailureReason string `json:"failure reason,omitempty"`
}

type websocketConn interface {
	WriteJSON(v any) error
}

func ResponseErr(conn websocketConn, err error) error {
	return conn.WriteJSON(errorResponse{FailureReason: err.Error()})
}
