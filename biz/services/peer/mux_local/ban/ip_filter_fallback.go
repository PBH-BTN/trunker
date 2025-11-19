//go:build !linux

package ban

import (
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

// NewIPFilter creates a new IP filter
// On non-Linux platforms (macOS, Windows, etc.), this always returns
// a fallback implementation that filters at application layer
func NewIPFilter(ifaceName string) (IPFilter, error) {
	if ifaceName != "" {
		hlog.Warnf("XDP is not supported on this platform, interface '%s' will be ignored", ifaceName)
	}
	hlog.Info("Using fallback IP filter (application-level filtering)")
	return newFallbackIPFilter(), nil
}
