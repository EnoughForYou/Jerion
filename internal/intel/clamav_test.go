package intel_test

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/EnoughForYou/Jerion/internal/intel"
)

// startMockClamd creates a local TCP listener simulating the ClamAV daemon.
func startMockClamd(t *testing.T, handler func(conn net.Conn)) (string, func()) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start mock clamd listener: %v", err)
	}

	done := make(chan struct{})
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				select {
				case <-done:
					return
				default:
					return
				}
			}
			go handler(conn)
		}
	}()

	cleanup := func() {
		close(done)
		_ = listener.Close()
	}

	return listener.Addr().String(), cleanup
}

func TestClamAV_PingAndVersion(t *testing.T) {
	addr, cleanup := startMockClamd(t, func(conn net.Conn) {
		defer conn.Close()
		buf := make([]byte, 64)
		n, err := conn.Read(buf)
		if err != nil {
			return
		}
		cmd := string(bytes.TrimSpace(buf[:n]))

		if strings.Contains(cmd, "PING") {
			_, _ = conn.Write([]byte("PONG\x00"))
		} else if strings.Contains(cmd, "VERSION") {
			_, _ = conn.Write([]byte("ClamAV 1.4.1/27600/Tue Oct 14 09:30:00 2026\x00"))
		}
	})
	defer cleanup()

	client := intel.NewClamAVClient(addr)
	ctx := context.Background()

	// 1. Test Ping
	if err := client.Ping(ctx); err != nil {
		t.Fatalf("Ping failed: %v", err)
	}

	// 2. Test Version
	ver, err := client.Version(ctx)
	if err != nil {
		t.Fatalf("Version failed: %v", err)
	}
	if !strings.Contains(ver, "ClamAV") {
		t.Fatalf("expected version containing 'ClamAV', got %q", ver)
	}
}

func TestClamAV_ScanCleanPayload(t *testing.T) {
	addr, cleanup := startMockClamd(t, func(conn net.Conn) {
		defer conn.Close()
		// Read zINSTREAM command
		cmdBuf := make([]byte, 10)
		_, _ = conn.Read(cmdBuf)

		// Read streamed chunks
		for {
			var length uint32
			if err := binary.Read(conn, binary.BigEndian, &length); err != nil {
				return
			}
			if length == 0 {
				break
			}
			data := make([]byte, length)
			_, _ = io.ReadFull(conn, data)
		}

		// Reply clean
		_, _ = conn.Write([]byte("stream: OK\x00"))
	})
	defer cleanup()

	client := intel.NewClamAVClient(addr)
	ctx := context.Background()

	sampleData := bytes.NewReader([]byte("PK\x03\x04benign-android-package-content"))
	res, err := client.ScanStream(ctx, sampleData)
	if err != nil {
		t.Fatalf("ScanStream failed: %v", err)
	}
	if res.Infected {
		t.Fatalf("expected clean result, got infected")
	}
	if !strings.Contains(res.Raw, "OK") {
		t.Fatalf("expected raw OK, got %s", res.Raw)
	}
}

func TestClamAV_ScanInfectedPayload(t *testing.T) {
	addr, cleanup := startMockClamd(t, func(conn net.Conn) {
		defer conn.Close()
		cmdBuf := make([]byte, 10)
		_, _ = conn.Read(cmdBuf)

		for {
			var length uint32
			if err := binary.Read(conn, binary.BigEndian, &length); err != nil {
				return
			}
			if length == 0 {
				break
			}
			data := make([]byte, length)
			_, _ = io.ReadFull(conn, data)
		}

		// Reply infected
		_, _ = conn.Write([]byte("stream: Android.Adware.HiddenAd-9876 FOUND\x00"))
	})
	defer cleanup()

	client := intel.NewClamAVClient(addr)
	ctx := context.Background()

	sampleData := bytes.NewReader([]byte("malicious-payload-signature"))
	res, err := client.ScanStream(ctx, sampleData)
	if err != nil {
		t.Fatalf("ScanStream failed: %v", err)
	}
	if !res.Infected {
		t.Fatalf("expected infected result")
	}
	if res.Virus != "Android.Adware.HiddenAd-9876" {
		t.Fatalf("unexpected virus name: %q", res.Virus)
	}
}

func TestClamAV_UnreachableDaemon(t *testing.T) {
	// Pick an address that is not listening
	client := intel.NewClamAVClient("127.0.0.1:54321")
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	err := client.Ping(ctx)
	if err == nil {
		t.Fatalf("expected error when clamd is unreachable")
	}
	if !errors.Is(err, intel.ErrClamAVUnavailable) {
		t.Fatalf("expected ErrClamAVUnavailable, got %v", err)
	}
}
