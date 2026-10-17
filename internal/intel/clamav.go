package intel

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// DefaultClamAVAddress is the default localhost clamd TCP endpoint.
const DefaultClamAVAddress = "127.0.0.1:3310"

var (
	ErrClamAVUnavailable       = errors.New("clamav daemon is unreachable")
	ErrClamAVProtocol          = errors.New("clamav protocol error")
	ErrClamAVNonLoopback       = errors.New("clamav host must be a local loopback address (e.g. 127.0.0.1, [::1], localhost)")
	ErrClamAVInvalidAddress    = errors.New("invalid clamav address format: must be host:port")
	ErrClamAVSizeLimitExceeded = errors.New("file exceeds ClamAV daemon StreamMaxLength limit")
)

// ValidateClamAVAddress validates that the target address is a local loopback address (IPv4, IPv6, or localhost).
// This ensures APK payloads are only streamed to a local daemon and prevents external network transmission.
func ValidateClamAVAddress(address string) (string, error) {
	addr := strings.TrimSpace(address)
	if addr == "" {
		return DefaultClamAVAddress, nil
	}

	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrClamAVInvalidAddress, err)
	}

	if host == "" {
		return "", fmt.Errorf("%w: host cannot be empty", ErrClamAVInvalidAddress)
	}

	portNum, err := strconv.Atoi(port)
	if err != nil || portNum < 1 || portNum > 65535 {
		return "", fmt.Errorf("%w: port %q must be between 1 and 65535", ErrClamAVInvalidAddress, port)
	}

	if strings.EqualFold(host, "localhost") {
		return net.JoinHostPort(host, strconv.Itoa(portNum)), nil
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return "", fmt.Errorf("%w: hostname %q is not permitted; only loopback (localhost, 127.0.0.1, [::1]) is allowed", ErrClamAVNonLoopback, host)
	}

	if !ip.IsLoopback() {
		return "", fmt.Errorf("%w: IP %q is not a local loopback address", ErrClamAVNonLoopback, host)
	}

	return net.JoinHostPort(host, strconv.Itoa(portNum)), nil
}

// ClamAVResult represents the outcome of a ClamAV scan.
type ClamAVResult struct {
	Infected bool   `json:"infected"`
	Virus    string `json:"virus,omitempty"`
	Raw      string `json:"raw"`
}

// ClamAVClient interacts with a local ClamAV daemon (clamd) over TCP loopback.
type ClamAVClient struct {
	address string
	timeout time.Duration
}

// NewClamAVClient initializes a new clamd client.
func NewClamAVClient(address string) *ClamAVClient {
	addr := strings.TrimSpace(address)
	if addr == "" {
		addr = DefaultClamAVAddress
	}
	return &ClamAVClient{
		address: addr,
		timeout: 10 * time.Second,
	}
}

func (c *ClamAVClient) dial(ctx context.Context) (net.Conn, error) {
	validAddr, err := ValidateClamAVAddress(c.address)
	if err != nil {
		return nil, err
	}
	d := net.Dialer{Timeout: c.timeout}
	conn, err := d.DialContext(ctx, "tcp", validAddr)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrClamAVUnavailable, err)
	}
	return conn, nil
}

// Ping checks if the ClamAV daemon is responsive.
func (c *ClamAVClient) Ping(ctx context.Context) error {
	conn, err := c.dial(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(c.timeout)); err != nil {
		return err
	}

	if _, err := conn.Write([]byte("zPING\x00")); err != nil {
		return fmt.Errorf("write ping failed: %w", err)
	}

	buf := make([]byte, 16)
	n, err := conn.Read(buf)
	if err != nil {
		return fmt.Errorf("read pong failed: %w", err)
	}

	resp := string(bytes.TrimSpace(buf[:n]))
	if !strings.Contains(resp, "PONG") {
		return fmt.Errorf("%w: unexpected response %q", ErrClamAVProtocol, resp)
	}
	return nil
}

// readUntilTerminator reads from conn until a NUL byte '\x00', newline '\n', EOF, or error occurs.
func readUntilTerminator(conn net.Conn) (string, error) {
	var sb strings.Builder
	buf := make([]byte, 512)
	for {
		n, err := conn.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			if idx := bytes.IndexAny(chunk, "\x00\r\n"); idx != -1 {
				sb.Write(chunk[:idx])
				return strings.TrimSpace(sb.String()), nil
			}
			sb.Write(chunk)
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			if sb.Len() > 0 {
				return strings.TrimSpace(sb.String()), nil
			}
			return "", err
		}
	}
	return strings.TrimSpace(sb.String()), nil
}

// Version retrieves the ClamAV engine and virus signature database version.
func (c *ClamAVClient) Version(ctx context.Context) (string, error) {
	conn, err := c.dial(ctx)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(c.timeout)); err != nil {
		return "", err
	}

	if _, err := conn.Write([]byte("zVERSION\x00")); err != nil {
		return "", fmt.Errorf("write version command failed: %w", err)
	}

	ver, err := readUntilTerminator(conn)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("read version failed: %w", err)
	}
	if ver == "" {
		return "", fmt.Errorf("%w: empty version response", ErrClamAVProtocol)
	}
	return ver, nil
}

// ScanStream streams an io.Reader (such as an APK) to clamd using the zINSTREAM protocol.
func (c *ClamAVClient) ScanStream(ctx context.Context, r io.Reader) (*ClamAVResult, error) {
	conn, err := c.dial(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// Use extended timeout for payload scanning (up to 2 minutes)
	scanTimeout := 2 * time.Minute
	if err := conn.SetDeadline(time.Now().Add(scanTimeout)); err != nil {
		return nil, err
	}

	if _, err := conn.Write([]byte("zINSTREAM\x00")); err != nil {
		return nil, fmt.Errorf("failed to initiate instream: %w", err)
	}

	// drainPendingResponse attempts to read an error response that clamd may have sent
	// prior to resetting or aborting the socket connection (e.g. StreamMaxLength exceeded).
	drainPendingResponse := func() string {
		_ = conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		resp, _ := readUntilTerminator(conn)
		return resp
	}

	// Stream chunks: [uint32 big-endian size][chunk bytes]
	buf := make([]byte, 32*1024) // 32KB chunks
	lenBuf := make([]byte, 4)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		n, readErr := r.Read(buf)
		if n > 0 {
			binary.BigEndian.PutUint32(lenBuf, uint32(n))
			if _, err := conn.Write(lenBuf); err != nil {
				if pending := drainPendingResponse(); pending != "" {
					return parseClamAVResponse(pending)
				}
				return nil, fmt.Errorf("failed to write chunk length: %w", err)
			}
			if _, err := conn.Write(buf[:n]); err != nil {
				if pending := drainPendingResponse(); pending != "" {
					return parseClamAVResponse(pending)
				}
				return nil, fmt.Errorf("failed to write chunk data: %w", err)
			}
		}
		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				break
			}
			return nil, fmt.Errorf("failed reading stream: %w", readErr)
		}
	}

	// Send terminal chunk: [0, 0, 0, 0]
	binary.BigEndian.PutUint32(lenBuf, 0)
	if _, err := conn.Write(lenBuf); err != nil {
		if pending := drainPendingResponse(); pending != "" {
			return parseClamAVResponse(pending)
		}
		return nil, fmt.Errorf("failed to write EOF chunk: %w", err)
	}

	// Reset read deadline for normal scan completion
	_ = conn.SetReadDeadline(time.Now().Add(scanTimeout))

	// Read response until terminator NUL/newline/EOF
	raw, err := readUntilTerminator(conn)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("failed reading clamd response: %w", err)
	}
	return parseClamAVResponse(raw)
}

// ScanFile opens and streams a local file to clamd.
func (c *ClamAVClient) ScanFile(ctx context.Context, filePath string) (*ClamAVResult, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot open file for clamav scan: %w", err)
	}
	defer f.Close()

	return c.ScanStream(ctx, f)
}

func parseClamAVResponse(raw string) (*ClamAVResult, error) {
	if raw == "" {
		return nil, fmt.Errorf("%w: empty response from daemon", ErrClamAVProtocol)
	}

	lower := strings.ToLower(raw)
	if strings.Contains(lower, "limit exceeded") {
		return nil, fmt.Errorf("%w: %s", ErrClamAVSizeLimitExceeded, raw)
	}

	// Clean result: "stream: OK"
	if strings.HasSuffix(raw, "OK") {
		return &ClamAVResult{
			Infected: false,
			Raw:      raw,
		}, nil
	}

	// Infected result: "stream: <VirusName> FOUND"
	if strings.HasSuffix(raw, "FOUND") {
		virusName := strings.TrimPrefix(raw, "stream:")
		virusName = strings.TrimSuffix(virusName, "FOUND")
		virusName = strings.TrimSpace(virusName)

		return &ClamAVResult{
			Infected: true,
			Virus:    virusName,
			Raw:      raw,
		}, nil
	}

	// Error response: "stream: ... ERROR"
	if strings.Contains(raw, "ERROR") {
		return nil, fmt.Errorf("%w: %s", ErrClamAVProtocol, raw)
	}

	// Fail-closed: any unrecognized response is a protocol violation, NEVER treated as clean.
	return nil, fmt.Errorf("%w: unrecognized clamd response %q", ErrClamAVProtocol, raw)
}
