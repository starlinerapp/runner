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
		fmt.Fprintf(w, "%s [%s]: ", label, defaultValue)
	} else {
		fmt.Fprintf(w, "%s: ", label)
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

func PositiveInt(r io.Reader, w io.Writer, label string) (int, error) {
	scanner := bufio.NewScanner(r)
	for {
		fmt.Fprintf(w, "%s: ", label)
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return 0, err
			}
			return 0, fmt.Errorf("max concurrent jobs is required")
		}

		value, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
		if err != nil || value < 1 {
			fmt.Fprintln(w, "Enter a positive integer.")
			continue
		}

		return value, nil
	}
}
