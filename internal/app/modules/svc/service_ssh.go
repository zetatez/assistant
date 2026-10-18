package svc

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"assistant/internal/bootstrap/psl"
	"assistant/internal/platform"
)

func (s *Service) SysSSHConnect() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	sshConfigPath := psl.GetConfig().Settings.PathSSHSecret
	sshConfigPath = strings.ReplaceAll(sshConfigPath, "~", homeDir)

	saved := map[string]struct {
		Host     string `json:"host"`
		User     string `json:"user"`
		Password string `json:"password"`
		Desc     string `json:"desc"`
	}{}

	var lines []string
	if data, err := os.ReadFile(sshConfigPath); err == nil {
		if err := json.Unmarshal(data, &saved); err == nil {
			for _, entry := range saved {
				line := fmt.Sprintf("%s@%s  # %s", entry.User, entry.Host, entry.Desc)
				lines = append(lines, line)
			}
		}
	}

	knownHostsPath := path.Join(homeDir, ".ssh/known_hosts")
	if data, err := os.ReadFile(knownHostsPath); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			fields := strings.Fields(line)
			if len(fields) >= 1 && fields[0] != "" {
				host := strings.TrimSpace(fields[0])
				if !strings.Contains(host, ",") {
					lines = append(lines, host)
				}
			}
		}
	}

	if len(lines) == 0 {
		return fmt.Errorf("no SSH entries found")
	}

	choice, err := platform.Get().Select("ssh to", lines)
	if err != nil || strings.TrimSpace(choice) == "" {
		return err
	}
	choice = strings.TrimSpace(choice)

	host, user, password := "", "", ""
	parts := strings.Fields(choice)
	for _, p := range parts {
		if strings.Contains(p, "@") {
			parts2 := strings.Split(p, "@")
			user = parts2[0]
			host = parts2[1]
		} else if strings.Contains(p, "#") {
			// description, skip
		} else if host == "" {
			host = p
		}
	}

	if host == "" {
		return fmt.Errorf("no host selected")
	}
	if user == "" {
		user = "root"
	}

	if entry, ok := saved[user+"@"+host]; ok {
		password = entry.Password
	}

	if password == "" {
		out2, err := platform.Get().Input("password", true)
		if err != nil || strings.TrimSpace(out2) == "" {
			return err
		}
		password = strings.TrimSpace(out2)
	}

	term := psl.GetConfig().Settings.DefaultTerminal
	return platform.Get().SSHConnect(term, user, host, password, 22)
}
