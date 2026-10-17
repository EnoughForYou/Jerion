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

func TestValidateClamAVAddress(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectError error
	}{
		// Valid loopback addresses
		{
			name:     "standard IPv4 loopback default",
			input:    "127.0.0.1:3310",
			expected: "127.0.0.1:3310",
		},
		{
			name:     "alternative IPv4 loopback IP",
			input:    "127.0.0.2:3310",
			expected: "127.0.0.2:3310",
		},
		{
			name:     "localhost hostname",
			input:    "localhost:3310",
			expected: "localhost:3310",
		},
		{
			name:     "uppercase LOCALHOST hostname",
			input:    "LOCALHOST:3310",
			expected: "LOCALHOST:3310",
		},
		{
			name:     "IPv6 loopback",
			input:    "[::1]:3310",
			expected: "[::1]:3310",
		},
		{
			name:     "empty address defaults to 127.0.0.1:3310",
			input:    "",
			expected: "127.0.0.1:3310",
		},
		{
			name:     "whitespace address defaults to 127.0.0.1:3310",
			input:    "   \t\n  ",
			expected: "127.0.0.1:3310",
		},
		{
			name:     "high valid port 65535",
			input:    "127.0.0.1:65535",
			expected: "127.0.0.1:65535",
		},
		{
			name:     "low valid port 1",
			input:    "127.0.0.1:1",
			expected: "127.0.0.1:1",
		},

		// Non-loopback IPs (must be rejected)
		{
			name:        "private LAN IP 10.0.0.1",
			input:       "10.0.0.1:3310",
			expectError: intel.ErrClamAVNonLoopback,
		},
		{
			name:        "private LAN IP 192.168.1.1",
			input:       "192.168.1.1:3310",
			expectError: intel.ErrClamAVNonLoopback,
		},
		{
			name:        "private LAN IP 172.16.0.5",
			input:       "172.16.0.5:3310",
			expectError: intel.ErrClamAVNonLoopback,
		},
		{
			name:        "public DNS IP 8.8.8.8",
			input:       "8.8.8.8:3310",
			expectError: intel.ErrClamAVNonLoopback,
		},
		{
			name:        "wildcard IP 0.0.0.0",
			input:       "0.0.0.0:3310",
			expectError: intel.ErrClamAVNonLoopback,
		},
		{
			name:        "external domain example.com",
			input:       "example.com:3310",
			expectError: intel.ErrClamAVNonLoopback,
		},
		{
			name:        "subdomain attempting to spoof localhost",
			input:       "localhost.attacker.com:3310",
			expectError: intel.ErrClamAVNonLoopback,
		},

		// Invalid format / ports
		{
			name:        "missing port in IPv4",
			input:       "127.0.0.1",
			expectError: intel.ErrClamAVInvalidAddress,
		},
		{
			name:        "missing port in localhost",
			input:       "localhost",
			expectError: intel.ErrClamAVInvalidAddress,
		},
		{
			name:        "missing host with colon and port",
			input:       ":3310",
			expectError: intel.ErrClamAVInvalidAddress,
		},
		{
			name:        "port zero",
			input:       "127.0.0.1:0",
			expectError: intel.ErrClamAVInvalidAddress,
		},
		{
			name:        "port above 65535",
			input:       "127.0.0.1:65536",
			expectError: intel.ErrClamAVInvalidAddress,
		},
		{
			name:        "non-numeric port",
			input:       "127.0.0.1:port",
			expectError: intel.ErrClamAVInvalidAddress,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res, err := intel.ValidateClamAVAddress(tc.input)
			if tc.expectError != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil (result: %q)", tc.expectError, res)
				}
				if !errors.Is(err, tc.expectError) {
					t.Fatalf("expected error %v, got %v", tc.expectError, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if res != tc.expected {
					t.Fatalf("expected %q, got %q", tc.expected, res)
				}
			}
		})
	}
}

func TestClamAV_SizeLimitExceeded_BrokenPipe(t *testing.T) {
	addr, cleanup := startMockClamd(t, func(conn net.Conn) {
		defer conn.Close()
		// Read zINSTREAM command
		cmdBuf := make([]byte, 10)
		_, _ = conn.Read(cmdBuf)

		// Read one chunk length and chunk data
		var length uint32
		if err := binary.Read(conn, binary.BigEndian, &length); err == nil && length > 0 {
			data := make([]byte, length)
			_, _ = io.ReadFull(conn, data)
		}

		// Simulate clamd hitting StreamMaxLength: send error and close write end
		_, _ = conn.Write([]byte("INSTREAM size limit exceeded. ERROR\x00"))
		if tcpConn, ok := conn.(*net.TCPConn); ok {
			_ = tcpConn.CloseWrite()
		}
		drainBuf := make([]byte, 4096)
		for {
			if _, err := conn.Read(drainBuf); err != nil {
				break
			}
		}
	})
	defer cleanup()

	client := intel.NewClamAVClient(addr)
	ctx := context.Background()

	// Large payload (128 KB) to trigger chunk writes after closure
	largeData := bytes.Repeat([]byte("A"), 128*1024)
	_, err := client.ScanStream(ctx, bytes.NewReader(largeData))
	if err == nil {
		t.Fatalf("expected error for exceeded size limit, got nil")
	}
	if !errors.Is(err, intel.ErrClamAVSizeLimitExceeded) {
		t.Fatalf("expected ErrClamAVSizeLimitExceeded, got %v", err)
	}
}

func TestClamAV_SizeLimitExceeded_NormalResponse(t *testing.T) {
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

		// Reply with limit exceeded in response phase
		_, _ = conn.Write([]byte("stream: Size limit exceeded. ERROR\x00"))
	})
	defer cleanup()

	client := intel.NewClamAVClient(addr)
	ctx := context.Background()

	sampleData := bytes.NewReader([]byte("small-payload"))
	_, err := client.ScanStream(ctx, sampleData)
	if err == nil {
		t.Fatalf("expected error for size limit exceeded, got nil")
	}
	if !errors.Is(err, intel.ErrClamAVSizeLimitExceeded) {
		t.Fatalf("expected ErrClamAVSizeLimitExceeded, got %v", err)
	}
}

func TestClamAV_NonLoopbackDialBlocked(t *testing.T) {
	client := intel.NewClamAVClient("10.0.0.1:3310")
	ctx := context.Background()

	err := client.Ping(ctx)
	if !errors.Is(err, intel.ErrClamAVNonLoopback) {
		t.Fatalf("expected ErrClamAVNonLoopback for Ping, got %v", err)
	}

	_, err = client.Version(ctx)
	if !errors.Is(err, intel.ErrClamAVNonLoopback) {
		t.Fatalf("expected ErrClamAVNonLoopback for Version, got %v", err)
	}

	_, err = client.ScanStream(ctx, bytes.NewReader([]byte("test")))
	if !errors.Is(err, intel.ErrClamAVNonLoopback) {
		t.Fatalf("expected ErrClamAVNonLoopback for ScanStream, got %v", err)
	}
}

func TestClamAV_FailClosedOnTruncatedOrUnrecognizedResponse(t *testing.T) {
	// 1. Truncated response without FOUND or OK
	addr, cleanup := startMockClamd(t, func(conn net.Conn) {
		defer conn.Close()
		buf := make([]byte, 10)
		_, _ = conn.Read(buf)
		// Drain stream
		for {
			var length uint32
			if err := binary.Read(conn, binary.BigEndian, &length); err != nil || length == 0 {
				break
			}
			data := make([]byte, length)
			_, _ = io.ReadFull(conn, data)
		}
		// Send truncated response and close
		_, _ = conn.Write([]byte("stream: Eicar-Signature-Partial\x00"))
	})
	defer cleanup()

	client := intel.NewClamAVClient(addr)
	_, err := client.ScanStream(context.Background(), bytes.NewReader([]byte("sample")))
	if err == nil {
		t.Fatalf("expected error for truncated/unrecognized response, got clean result")
	}
	if !errors.Is(err, intel.ErrClamAVProtocol) {
		t.Fatalf("expected ErrClamAVProtocol, got %v", err)
	}

	// 2. Garbage response
	addr2, cleanup2 := startMockClamd(t, func(conn net.Conn) {
		defer conn.Close()
		buf := make([]byte, 10)
		_, _ = conn.Read(buf)
		for {
			var length uint32
			if err := binary.Read(conn, binary.BigEndian, &length); err != nil || length == 0 {
				break
			}
			data := make([]byte, length)
			_, _ = io.ReadFull(conn, data)
		}
		_, _ = conn.Write([]byte("UNKNOWN_DAEMON_OUTPUT_12345\n"))
	})
	defer cleanup2()

	client2 := intel.NewClamAVClient(addr2)
	_, err2 := client2.ScanStream(context.Background(), bytes.NewReader([]byte("sample")))
	if err2 == nil {
		t.Fatalf("expected error for garbage response, got clean result")
	}
	if !errors.Is(err2, intel.ErrClamAVProtocol) {
		t.Fatalf("expected ErrClamAVProtocol, got %v", err2)
	}
}

func TestClamAV_SegmentedResponseReassembly(t *testing.T) {
	addr, cleanup := startMockClamd(t, func(conn net.Conn) {
		defer conn.Close()
		buf := make([]byte, 10)
		_, _ = conn.Read(buf)
		for {
			var length uint32
			if err := binary.Read(conn, binary.BigEndian, &length); err != nil || length == 0 {
				break
			}
			data := make([]byte, length)
			_, _ = io.ReadFull(conn, data)
		}
		// Send response across two separate writes without trailing newline on first write
		_, _ = conn.Write([]byte("stream: Android.Trojan.Banker-4321"))
		time.Sleep(20 * time.Millisecond)
		_, _ = conn.Write([]byte(" FOUND\x00"))
	})
	defer cleanup()

	client := intel.NewClamAVClient(addr)
	res, err := client.ScanStream(context.Background(), bytes.NewReader([]byte("sample")))
	if err != nil {
		t.Fatalf("ScanStream failed on segmented response: %v", err)
	}
	if !res.Infected {
		t.Fatalf("expected infected for reassembled response")
	}
	if res.Virus != "Android.Trojan.Banker-4321" {
		t.Fatalf("expected virus Android.Trojan.Banker-4321, got %q", res.Virus)
	}
}
