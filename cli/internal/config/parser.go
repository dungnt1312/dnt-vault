package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type SSHConfig struct {
	Hosts []Host
	Raw   string
}

type Host struct {
	Name         string
	HostName     string
	User         string
	Port         string
	IdentityFile string
	Other        map[string]string
}

func ParseSSHConfig(path string) (*SSHConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := &SSHConfig{
		Raw:   string(data),
		Hosts: []Host{},
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var currentHost *Host

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		key := parts[0]
		value := strings.Join(parts[1:], " ")

		if strings.EqualFold(key, "Host") {
			if currentHost != nil {
				config.Hosts = append(config.Hosts, *currentHost)
			}
			currentHost = &Host{
				Name:  value,
				Other: make(map[string]string),
			}
		} else if currentHost != nil {
			switch strings.ToLower(key) {
			case "hostname":
				currentHost.HostName = value
			case "user":
				currentHost.User = value
			case "port":
				currentHost.Port = value
			case "identityfile":
				currentHost.IdentityFile = expandPath(value)
			default:
				currentHost.Other[key] = value
			}
		}
	}

	if currentHost != nil {
		config.Hosts = append(config.Hosts, *currentHost)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return config, nil
}

func (c *SSHConfig) GetIdentityFiles() []string {
	files := make(map[string]bool)
	for _, host := range c.Hosts {
		if host.IdentityFile != "" {
			files[host.IdentityFile] = true
		}
	}

	result := make([]string, 0, len(files))
	for file := range files {
		result = append(result, file)
	}
	return result
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

func WriteSSHConfig(path, content string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	return os.WriteFile(path, []byte(content), 0600)
}

func ValidateSSHConfig(content string) error {
	if content == "" {
		return fmt.Errorf("empty config")
	}
	return nil
}
