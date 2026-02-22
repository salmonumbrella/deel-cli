package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const openClawEnvRelativePath = ".openclaw/.env"

// LoadOpenClawEnv loads ~/.openclaw/.env when present.
// Existing process environment variables always take precedence.
func LoadOpenClawEnv() error {
	envPath, err := openClawEnvPath()
	if err != nil {
		return fmt.Errorf("resolve OpenClaw env path: %w", err)
	}

	return loadEnvFile(envPath)
}

func openClawEnvPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, openClawEnvRelativePath), nil
}

func loadEnvFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("open env file: %w", err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Warning: failed to close env file %s: %v\n", path, closeErr)
		}
	}()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		key, value, ok, err := parseEnvLine(scanner.Text())
		if err != nil {
			continue
		}
		if !ok {
			continue
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("set env %q from %s: %w", key, path, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan env file: %w", err)
	}

	return nil
}

func parseEnvLine(line string) (key string, value string, ok bool, err error) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return "", "", false, nil
	}

	if strings.HasPrefix(line, "export ") {
		line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
	}

	idx := strings.IndexRune(line, '=')
	if idx <= 0 {
		return "", "", false, fmt.Errorf("expected KEY=VALUE")
	}

	key = strings.TrimSpace(line[:idx])
	if key == "" {
		return "", "", false, fmt.Errorf("missing key")
	}

	value = strings.TrimSpace(line[idx+1:])
	if len(value) >= 2 {
		if value[0] == '"' && value[len(value)-1] == '"' {
			unquoted, uErr := strconv.Unquote(value)
			if uErr != nil {
				return "", "", false, fmt.Errorf("invalid quoted value: %w", uErr)
			}
			value = unquoted
		} else if value[0] == '\'' && value[len(value)-1] == '\'' {
			value = value[1 : len(value)-1]
		}
	}

	return key, value, true, nil
}
