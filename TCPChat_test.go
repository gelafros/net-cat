package main

import (
	"net"
	"strings"
	"testing"
	"time"

	"net-cat/handlers"
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

func TestHandleClient_RequiresNonEmptyNameAndTrims(t *testing.T) {
	server := handlers.NewServer()
	serverConn, clientConn := net.Pipe()
	defer clientConn.Close()

	done := make(chan struct{})
	go func() {
		handlers.HandleClient(serverConn, server)
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
		t.Fatalf("expected retry prompt, got %q", retryPrompt)
	}

	if _, err := clientConn.Write([]byte("   Bob   \n")); err != nil {
		t.Fatalf("failed to send valid name: %v", err)
	}

	welcome := readUntilContains(t, clientConn, "Welcome to chat room, Bob")
	if !strings.Contains(welcome, "Welcome to chat room, Bob") {
		t.Fatalf("expected welcome message for Bob, got %q", welcome)
	}

	_ = clientConn.Close()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("HandleClient did not return after client disconnect")
	}
}

func TestHandleClient_RejectsDuplicateName(t *testing.T) {
	server := handlers.NewServer()

	// First client claims "Alice".
	serverConn1, clientConn1 := net.Pipe()
	done1 := make(chan struct{})
	go func() {
		handlers.HandleClient(serverConn1, server)
		close(done1)
	}()
	readUntilContains(t, clientConn1, "[ENTER YOUR NAME]: ")
	if _, err := clientConn1.Write([]byte("Alice\n")); err != nil {
		t.Fatalf("first client failed to send name: %v", err)
	}
	readUntilContains(t, clientConn1, "Welcome to chat room, Alice")

	// Second client tries same name, then a new one.
	serverConn2, clientConn2 := net.Pipe()
	defer clientConn2.Close()
	done2 := make(chan struct{})
	go func() {
		handlers.HandleClient(serverConn2, server)
		close(done2)
	}()
	readUntilContains(t, clientConn2, "[ENTER YOUR NAME]: ")

	if _, err := clientConn2.Write([]byte("Alice\n")); err != nil {
		t.Fatalf("second client failed to send duplicate name: %v", err)
	}
	dupMsg := readUntilContains(t, clientConn2, "Name in use or server full.")
	if !strings.Contains(dupMsg, "[ENTER YOUR NAME]: ") {
		t.Fatalf("expected re-prompt after duplicate name, got %q", dupMsg)
	}

	if _, err := clientConn2.Write([]byte("Alice2\n")); err != nil {
		t.Fatalf("second client failed to send unique name: %v", err)
	}
	readUntilContains(t, clientConn2, "Welcome to chat room, Alice2")

	_ = clientConn1.Close()
	_ = clientConn2.Close()

	select {
	case <-done1:
	case <-time.After(2 * time.Second):
		t.Fatal("first HandleClient goroutine did not return")
	}
	select {
	case <-done2:
	case <-time.After(2 * time.Second):
		t.Fatal("second HandleClient goroutine did not return")
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
