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
			expected: "-qB452B",
		},
		{
			name:     "validPeerIDWithDifferentDashPrefix",
			peerID:   string(conv.TransUTF8To8859_1([]byte("-BC0211-\f\u00a9\u00c4Q\u0000 \u00a7\u000f\u00d1\u00cc\u0084\u0014"))),
			expected: "-BC0211",
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
		{
			name:     "aria21",
			peerID:   "A2-1-35-0-��)\u0014\u0015��\u0006�",
			expected: "A2-1-35-0",
		},
		{
			name:     "aria22",
			peerID:   "A2-1-34-0-��W��f�F",
			expected: "A2-1-34-0",
		},
		{
			name:     "middleDash1",
			peerID:   "TIX0332-77wfm2rcxovo",
			expected: "TIX0332",
		},
		{
			name:     "middleDash2",
			peerID:   "TIX0262-e1f1a6e7g7f3",
			expected: "TIX0262",
		},
		{
			name:     "FD6",
			peerID:   "FD68Ki0o~Jd0mWb(GCY5",
			expected: "FD6",
		},
		{
			name:     "FD5",
			peerID:   "-FD51]�-FdrWCsIvJAk4",
			expected: "-FD51",
		},
		{
			name:     "BS",
			peerID:   "12BS�\u007F]���A%��o\u001F�U\u0006�",
			expected: "12BS",
		},
		{
			name:     "MG",
			peerID:   "MG-3.01.43194l*c0n8.",
			expected: "MG-3.01.4319",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParsePeerID(tt.peerID)
			assert.Equal(t, tt.expected, result)
		})
	}
}
