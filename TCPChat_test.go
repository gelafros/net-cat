package main

import (
	"net"
	"strings"
	"testing"
	"time"
)

func TestGetPort(t *testing.T) {
	t.Run("default port when no argument", func(t *testing.T) {
		got, err := getPort([]string{"./TCPChat"})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if got != "8989" {
			t.Fatalf("expected 8989, got %q", got)
		}
	})

	t.Run("custom port when provided", func(t *testing.T) {
		got, err := getPort([]string{"./TCPChat", "9000"})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if got != "9000" {
			t.Fatalf("expected 9000, got %q", got)
		}
	})

	t.Run("usage error for too many arguments", func(t *testing.T) {
		_, err := getPort([]string{"./TCPChat", "9000", "extra"})
		if err == nil {
			t.Fatal("expected usage error, got nil")
		}
	})
}

func TestHandleClient_RequiresNonEmptyName(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer clientConn.Close()

	done := make(chan struct{})
	go func() {
		handleClient(serverConn, "[ENTER YOUR NAME]: ", false)
		close(done)
	}()

	initialPrompt := readUntilContains(t, clientConn, "[ENTER YOUR NAME]: ")
	if !strings.Contains(initialPrompt, "[ENTER YOUR NAME]: ") {
		t.Fatalf("expected initial prompt, got %q", initialPrompt)
	}

	if _, err := clientConn.Write([]byte("\n")); err != nil {
		t.Fatalf("failed to send empty name: %v", err)
	}

	retryPrompt := readUntilContains(t, clientConn, "You must enter a name")
	if !strings.Contains(retryPrompt, "[ENTER YOUR NAME]: ") {
		t.Fatalf("expected retry name prompt, got %q", retryPrompt)
	}

	if _, err := clientConn.Write([]byte("Alice\n")); err != nil {
		t.Fatalf("failed to send valid name: %v", err)
	}

	welcome := readUntilContains(t, clientConn, "Welcome to chat room, Alice\n")
	if !strings.Contains(welcome, "Welcome to chat room, Alice") {
		t.Fatalf("expected welcome message, got %q", welcome)
	}

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("handleClient did not return after valid username")
	}
}

func TestHandshake_RejectsEmptyAndTrimsName(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer clientConn.Close()

	result := make(chan string, 1)
	go func() {
		result <- handshake(serverConn)
	}()

	if _, err := clientConn.Write([]byte("\n")); err != nil {
		t.Fatalf("failed to send empty name: %v", err)
	}

	retryPrompt := readUntilContains(t, clientConn, "You must enter a name")
	if !strings.Contains(retryPrompt, "[ENTER YOUR NAME]: ") {
		t.Fatalf("expected retry prompt after empty name, got %q", retryPrompt)
	}

	if _, err := clientConn.Write([]byte("   Bob   \n")); err != nil {
		t.Fatalf("failed to send valid name: %v", err)
	}

	select {
	case got := <-result:
		if got != "Bob" {
			t.Fatalf("expected trimmed name Bob, got %q", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("handshake did not return after valid username")
	}
}

func TestHandleClient_ServerFullSkipsHandshake(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer clientConn.Close()

	done := make(chan struct{})
	go func() {
		handleClient(serverConn, "Server Busy", true)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("handleClient did not return when server is full")
	}

	if err := clientConn.SetReadDeadline(time.Now().Add(200 * time.Millisecond)); err != nil {
		t.Fatalf("failed to set read deadline: %v", err)
	}

	var buf [1]byte
	_, err := clientConn.Read(buf[:])
	if err == nil {
		t.Fatal("expected no data from handleClient when server is full")
	}
	if ne, ok := err.(net.Error); !ok || !ne.Timeout() {
		t.Fatalf("expected read timeout with no data, got %v", err)
	}
}

func readUntilContains(t *testing.T, conn net.Conn, want string) string {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	buf := make([]byte, 256)
	var out strings.Builder

	for time.Now().Before(deadline) {
		if err := conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond)); err != nil {
			t.Fatalf("failed to set read deadline: %v", err)
		}

		n, err := conn.Read(buf)
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				continue
			}
			t.Fatalf("read failed while waiting for %q: %v", want, err)
		}

		out.Write(buf[:n])
		if strings.Contains(out.String(), want) {
			return out.String()
		}
	}

	t.Fatalf("timed out waiting for %q; received %q", want, out.String())
	return ""
}
