package main

import (
	"testing"
)

func TestGetPort(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectedPort  string
		expectedError bool
	}{
		{
			name:          "Default port when no arguments provided",
			args:          []string{"./TCPChat"},
			expectedPort:  "8989",
			expectedError: false,
		},
		{
			name:          "Custom port provided",
			args:          []string{"./TCPChat", "2525"},
			expectedPort:  "2525",
			expectedError: false,
		},
		{
			name:          "Too many arguments",
			args:          []string{"./TCPChat", "2525", "localhost"},
			expectedPort:  "",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port, err := handlers.getPort(tt.args)

			if (err != nil) != tt.expectedError {
				t.Errorf("expected error: %v, got error: %v", tt.expectedError, err)
			}

			if port != tt.expectedPort {
				t.Errorf("expected port: %q, got: %q", tt.expectedPort, port)
			}
		})
	}
}
