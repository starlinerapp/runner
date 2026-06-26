package prompt

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func String(r io.Reader, w io.Writer, label, defaultValue string) (string, error) {
	if defaultValue != "" {
		if _, err := fmt.Fprintf(w, "%s [%s]: ", label, defaultValue); err != nil {
			return "", err
		}
	} else {
		if _, err := fmt.Fprintf(w, "%s: ", label); err != nil {
			return "", err
		}
	}

	scanner := bufio.NewScanner(r)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return "", err
		}
		return defaultValue, nil
	}

	value := strings.TrimSpace(scanner.Text())
	if value == "" {
		return defaultValue, nil
	}

	return value, nil
}

func Labels(r io.Reader, w io.Writer, defaultLabels []string) ([]string, error) {
	defaultValue := strings.Join(defaultLabels, ",")
	input, err := String(r, w, "Labels (comma-separated)", defaultValue)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(input, ",")
	labels := make([]string, 0, len(parts))
	for _, part := range parts {
		if label := strings.TrimSpace(part); label != "" {
			labels = append(labels, label)
		}
	}

	return labels, nil
}

func PositiveInt(r io.Reader, w io.Writer, label string, defaultValue int) (int, error) {
	scanner := bufio.NewScanner(r)
	for {
		if _, err := fmt.Fprintf(w, "%s [%d]: ", label, defaultValue); err != nil {
			return 0, err
		}
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return 0, err
			}
			return defaultValue, nil
		}

		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			return defaultValue, nil
		}

		value, err := strconv.Atoi(text)
		if err != nil || value < 1 {
			if _, writeErr := fmt.Fprintln(w, "Enter a positive integer."); writeErr != nil {
				return 0, writeErr
			}
			continue
		}

		return value, nil
	}
}
