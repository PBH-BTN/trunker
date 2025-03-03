package bittorrent

import (
	"testing"

	"github.com/PBH-BTN/trunker/utils/conv"
	"github.com/stretchr/testify/assert"
)

func TestParsePeerID(t *testing.T) {
	tests := []struct {
		name     string
		peerID   string
		expected string
	}{
		{
			name:     "validPeerIDWithDashPrefix",
			peerID:   "-qB452B-w(33PK*XDMdD",
			expected: "-qB452B-",
		},
		{
			name:     "validPeerIDWithDifferentDashPrefix",
			peerID:   string(conv.TransUTF8To8859_1([]byte("-BC0211-\f\u00a9\u00c4Q\u0000 \u00a7\u000f\u00d1\u00cc\u0084\u0014"))),
			expected: "-BC0211-",
		},
		{
			name:     "validPeerIDWithSlashPrefix",
			peerID:   "-SP3604ndasgasgwegaeg",
			expected: "-SP3604",
		},
		{
			name:     "validPeerIDWithSlashPrefixAndNumberSuffic",
			peerID:   "-SP36041dasgasgwegaeg",
			expected: "-SP3604",
		},
		{
			name:     "Random PeerId",
			peerID:   "qB452B-w(33PK*XDMdD",
			expected: "unknown",
		},
		{
			name:     "peerIDWithLessThanEightCharacters",
			peerID:   "qB452B",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParsePeerID(tt.peerID)
			assert.Equal(t, tt.expected, result)
		})
	}
}
