package utils

import (
	"fmt"
	"strings"
)

func Notify(msg ...any) {
	escapeShellArg := func(s string) string {
		s = strings.ReplaceAll(s, `'`, `\'`)
		s = strings.ReplaceAll(s, `"`, `\"`)
		return s
	}
	message := escapeShellArg(fmt.Sprint(msg...))
	CmdMap := map[string][]string{
		"linux":  {"bash", fmt.Sprintf("notify-send '%s'", message)},
		"darwin": {"bash", fmt.Sprintf(`osascript -e 'display notification "%s" with title "%s"'`, message, "msg")},
	}
	OSType := GetOSType()
	if cmd, ok := CmdMap[OSType]; ok {
		RunScript(cmd[0], cmd[1])
		return
	}
	fmt.Println("unsupported OS:", OSType)
	return
}
