package svc

import (
	"assistant/internal/bootstrap/psl"
	"assistant/internal/platform"
)

// VimEditInput toggles the vim compose popup for the currently focused input
// box. The OS-specific plumbing (window focus/geometry/typing) lives in the
// platform layer: X11/xdotool on Linux, AppleScript on macOS.
func (s *Service) VimEditInput() error {
	term := psl.GetConfig().Settings.DefaultTerminal
	return platform.Get().VimEditInput(term)
}
