package model

import (
	"io"
	"strconv"

	"github.com/PBH-BTN/trunker/utils/conv"
)

// this file implements the fast bencode encoding method

// d14:failure reason11:bad request8:retry in5:never10:tracker id7:defaulte

func (e *ErrorResponse) Bencode(w io.Writer) error {
	_, _ = w.Write(conv.UnsafeStringToBytes("d14:failure reason"))
	writeString(w, e.FailureReason)
	_, _ = w.Write(conv.UnsafeStringToBytes("8:retry in"))
	writeString(w, e.Retry)
	if len(e.TrackerId) > 0 {
		_, _ = w.Write(conv.UnsafeStringToBytes("10:tracker id"))
		writeString(w, e.TrackerId)
	}
	_, err := w.Write(conv.UnsafeStringToBytes("e"))
	return err
}

func writeString(w io.Writer, s string) {
	_, _ = w.Write(conv.UnsafeStringToBytes(strconv.Itoa(len(s))))
	_, _ = w.Write(conv.UnsafeStringToBytes(":"))
	_, _ = w.Write(conv.UnsafeStringToBytes(s))
}

func writeBytes(w io.Writer, s []byte) {
	_, _ = w.Write(conv.UnsafeStringToBytes(strconv.Itoa(len(s))))
	_, _ = w.Write(conv.UnsafeStringToBytes(":"))
	_, _ = w.Write(s)
}

func writeInt(w io.Writer, i int) {
	_, _ = w.Write(conv.UnsafeStringToBytes("i"))
	_, _ = w.Write(conv.UnsafeStringToBytes(strconv.Itoa(i)))
	_, _ = w.Write(conv.UnsafeStringToBytes("e"))
}

// d8:completei0e11:external ip4:�_@�10:incompletei2e8:intervali3596e5:peers6:tuG��10:tracker id7:Tribbiee
func (a *AnnounceCompactResponse) Bencode(w io.Writer) error {
	_, _ = w.Write(conv.UnsafeStringToBytes("d8:complete"))
	writeInt(w, a.Complete)
	_, _ = w.Write(conv.UnsafeStringToBytes("11:external ip"))
	writeString(w, a.ExternalIp)
	_, _ = w.Write(conv.UnsafeStringToBytes("10:incomplete"))
	writeInt(w, a.Incomplete)
	_, _ = w.Write(conv.UnsafeStringToBytes("8:interval"))
	writeInt(w, int(a.Interval))
	if len(a.Peers) > 0 {
		_, _ = w.Write(conv.UnsafeStringToBytes("5:peers"))
		writeBytes(w, a.Peers)
	}
	if len(a.Peers6) > 0 {
		_, _ = w.Write(conv.UnsafeStringToBytes("6:peers6"))
		writeBytes(w, a.Peers6)
	}
	if len(a.TrackerId) > 0 {
		_, _ = w.Write(conv.UnsafeStringToBytes("10:tracker id"))
		writeString(w, a.TrackerId)
	}
	_, err := w.Write(conv.UnsafeStringToBytes("e"))
	return err
}
