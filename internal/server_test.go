package internal

import (
	"testing"
)

// Tests the getPort function with different command-line arguments
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
			port, err := getPort(tt.args)

			if (err != nil) != tt.expectedError {
				t.Errorf("expected error: %v, got error: %v", tt.expectedError, err)
			}

			if port != tt.expectedPort {
				t.Errorf("expected port: %q, got: %q", tt.expectedPort, port)
			}
		})
	}
}

func TestStartServer(t *testing.T) {
	// Test 1: Successful start (using port 0 to avoid conflicts)
	t.Run("Valid arguments", func(t *testing.T) {
		listener, err := StartServer([]string{"./TCPChat", "0"})
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if listener == nil {
			t.Fatalf("Expected a valid listener, got nil")
		}
		listener.Close() // Clean up the listener so it doesn't leak memory
	})

	// Test 2: Invalid arguments (should fail getPort)
	t.Run("Invalid arguments", func(t *testing.T) {
		listener, err := StartServer([]string{"./TCPChat", "2525", "localhost"})
		if err == nil {
			t.Errorf("Expected an error for invalid arguments, got nil")
		}
		if listener != nil {
			listener.Close()
		}
	})

	// Test 3: Invalid port number (should fail net.Listen)
	t.Run("Invalid port", func(t *testing.T) {
		_, err := StartServer([]string{"./TCPChat", "999999"})
		if err == nil {
			t.Errorf("Expected an error for invalid port, got nil")
		}
	})
}
