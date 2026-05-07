package envmanager

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func ParseEnvFile(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	out := make(map[string]string)
	s := bufio.NewScanner(f)
	lineNo := 0
	for s.Scan() {
		lineNo++
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid env line %d", lineNo)
		}
		k := strings.TrimSpace(parts[0])
		v := strings.TrimSpace(parts[1])
		if strings.HasPrefix(v, `"`) && strings.HasSuffix(v, `"`) && len(v) >= 2 {
			v = v[1 : len(v)-1]
		}
		if _, exists := out[k]; exists {
			return nil, fmt.Errorf("duplicate variable %s in input", k)
		}
		out[k] = v
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
