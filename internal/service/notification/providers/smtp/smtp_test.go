package smtp

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/squall-chua/go-grpc-auth/internal/service/notification"
)

func TestSMTPProviderSendsMessage(t *testing.T) {
	server, _, port := startTestSMTPServer(t)
	defer server.close()

	p, err := New(Config{
		Host:        "127.0.0.1",
		Port:        port,
		FromAddress: "noreply@example.com",
		FromName:    "Test",
	})
	if err != nil {
		t.Fatal(err)
	}

	err = p.Send(context.Background(), notification.EmailMessage{
		To:       "alice@example.com",
		Subject:  "Hello",
		HTMLBody: "<p>Hi</p>",
		TextBody: "Hi",
	})
	if err != nil {
		t.Fatalf("send: %v", err)
	}

	got := server.received(2 * time.Second)
	if !strings.Contains(got, "RCPT TO:<alice@example.com>") {
		t.Errorf("missing RCPT: %s", got)
	}
	if !strings.Contains(got, "Subject: Hello") {
		t.Errorf("missing subject: %s", got)
	}
}

// --- minimal test SMTP server ---

type testSMTPServer struct {
	ln       net.Listener
	mu       sync.Mutex
	captured string
	done     chan struct{}
}

func startTestSMTPServer(t *testing.T) (*testSMTPServer, string, int) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	s := &testSMTPServer{ln: ln, done: make(chan struct{})}
	go s.accept()
	addr := ln.Addr().String()
	_, portStr, _ := net.SplitHostPort(addr)
	var port int
	fmt.Sscanf(portStr, "%d", &port)
	return s, addr, port
}

func (s *testSMTPServer) accept() {
	conn, err := s.ln.Accept()
	if err != nil {
		return
	}
	defer conn.Close()
	defer close(s.done)
	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)
	writeLine := func(line string) { w.WriteString(line + "\r\n"); w.Flush() }
	writeLine("220 test.local ESMTP")
	var captured strings.Builder
	inData := false
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return
		}
		captured.WriteString(line)
		trimmed := strings.TrimRight(line, "\r\n")
		switch {
		case inData:
			if trimmed == "." {
				inData = false
				writeLine("250 OK")
			}
		case strings.HasPrefix(strings.ToUpper(trimmed), "EHLO"), strings.HasPrefix(strings.ToUpper(trimmed), "HELO"):
			writeLine("250-test.local")
			writeLine("250 OK")
		case strings.HasPrefix(strings.ToUpper(trimmed), "MAIL FROM"), strings.HasPrefix(strings.ToUpper(trimmed), "RCPT TO"):
			writeLine("250 OK")
		case strings.EqualFold(trimmed, "DATA"):
			writeLine("354 End data with <CR><LF>.<CR><LF>")
			inData = true
		case strings.EqualFold(trimmed, "QUIT"):
			writeLine("221 Bye")
			s.mu.Lock()
			s.captured = captured.String()
			s.mu.Unlock()
			return
		default:
			writeLine("250 OK")
		}
	}
}

func (s *testSMTPServer) received(timeout time.Duration) string {
	select {
	case <-s.done:
	case <-time.After(timeout):
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.captured
}

func (s *testSMTPServer) close() { s.ln.Close() }
