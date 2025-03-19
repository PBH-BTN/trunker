package model

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"net"
	"testing"

	"github.com/cilium/fake"
	"github.com/cristalhq/bencode"
	"github.com/stretchr/testify/assert"
	"github.com/zhangyunhao116/fastrand"
)

func TestFastBencodeErrorResponse(t *testing.T) {
	for i := 0; i < 100; i++ {
		e := ErrorResponse{
			FailureReason: rand.Text(),
			Retry:         rand.Text()[:8],
			TrackerId:     rand.Text()[:10],
		}
		buf := bytes.NewBuffer(make([]byte, 0, 1024))
		assert.Nil(t, e.Bencode(buf))
		ref, _ := bencode.Marshal(e)
		assert.Equal(t, string(ref), buf.String())
	}
}

func ipv4Generator(num int) []byte {
	buf := make([]byte, 0, 6*num)
	port := make([]byte, 2)

	for i := 0; i < num; i++ {
		p := fastrand.Intn(65533)
		binary.BigEndian.PutUint16(port, uint16(p))
		buf = append(buf, net.ParseIP(fake.IP(fake.WithIPv4())).To4()...)
		buf = append(buf, port...)
	}
	return buf
}

func ipv6Generator(num int) []byte {
	buf := make([]byte, 0, 18*num)
	port := make([]byte, 2)
	for i := 0; i < num; i++ {
		p := fastrand.Intn(65533)
		binary.BigEndian.PutUint16(port, uint16(p))
		buf = append(buf, net.ParseIP(fake.IP(fake.WithIPv6())).To16()...)
		buf = append(buf, port...)
	}
	return buf
}

func TestFastBencodeAnnounceCompactResponse(t *testing.T) {
	for i := 0; i < 100; i++ {
		a := AnnounceCompactResponse{
			Interval:   int64(fastrand.Intn(2600)),
			Peers:      ipv4Generator(fastrand.Intn(40)),
			Peers6:     ipv6Generator(fastrand.Intn(40)),
			ExternalIp: string(net.ParseIP(fake.IP())),
			TrackerId:  "default",
			Complete:   fastrand.Intn(300),
			Incomplete: fastrand.Intn(300),
		}
		buf := bytes.NewBuffer(make([]byte, 0, 1024))
		assert.Nil(t, a.Bencode(buf))
		ref, _ := bencode.Marshal(a)
		assert.Equal(t, string(ref), buf.String())
	}
}

// Benchmark for bencode.Marshal
func BenchmarkBencodeErrorMarshal(b *testing.B) {
	e := ErrorResponse{
		FailureReason: "bad request",
		Retry:         "never",
		TrackerId:     "default",
	}
	for i := 0; i < b.N; i++ {
		_, _ = bencode.Marshal(e)
	}
}

// Benchmark for cust marshal
func BenchmarkBencodeCustomMarshal(b *testing.B) {
	e := ErrorResponse{
		FailureReason: "bad request",
		Retry:         "never",
		TrackerId:     "default",
	}
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	for i := 0; i < b.N; i++ {
		_ = e.Bencode(buf)
		buf.Reset()
	}
}

func BenchmarkBencodeAnnounceCompactResponse(b *testing.B) {
	a := AnnounceCompactResponse{
		Interval:   int64(fastrand.Intn(2600)),
		Peers:      ipv4Generator(fastrand.Intn(40)),
		Peers6:     ipv6Generator(fastrand.Intn(40)),
		ExternalIp: string(net.ParseIP(fake.IP())),
		TrackerId:  "default",
		Complete:   fastrand.Intn(300),
		Incomplete: fastrand.Intn(300),
	}
	for i := 0; i < b.N; i++ {
		_, _ = bencode.Marshal(a)
	}
}

func BenchmarkCustomAnnounceCompactBencode(b *testing.B) {
	a := AnnounceCompactResponse{
		Interval:   int64(fastrand.Intn(2600)),
		Peers:      ipv4Generator(fastrand.Intn(40)),
		Peers6:     ipv6Generator(fastrand.Intn(40)),
		ExternalIp: string(net.ParseIP(fake.IP())),
		TrackerId:  "default",
		Complete:   fastrand.Intn(300),
		Incomplete: fastrand.Intn(300),
	}
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	for i := 0; i < b.N; i++ {
		_ = a.Bencode(buf)
		buf.Reset()
	}
}
