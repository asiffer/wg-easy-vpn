package utils

import (
	"bufio"
	"errors"
	"os"
	"strings"
)

var ErrNoDefaultRoute = errors.New("no default route found")

// GetDefaultInterface returns the network interface used for the default route.
// It parses /proc/net/route to find the interface associated with destination 0.0.0.0.
func GetDefaultInterface() (string, error) {
	f, err := os.Open("/proc/net/route")
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Scan() // skip header line
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		// fields[0] = interface, fields[1] = destination (hex)
		if len(fields) >= 2 && fields[1] == "00000000" {
			return fields[0], nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", ErrNoDefaultRoute
}