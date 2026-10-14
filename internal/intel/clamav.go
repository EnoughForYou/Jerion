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
	"strings"
	"time"
)

// DefaultClamAVAddress is the default localhost clamd TCP endpoint.
const DefaultClamAVAddress = "127.0.0.1:3310"

var (
	ErrClamAVUnavailable = errors.New("clamav daemon is unreachable")
	ErrClamAVProtocol    = errors.New("clamav protocol error")
)

// ClamAVResult represents the outcome of a ClamAV scan.
type ClamAVResult struct {
	Infected bool   `json:"infected"`
	Virus    string `json:"virus,omitempty"`
	Raw      string `json:"raw"`
}

// ClamAVClient interacts with a local or remote ClamAV daemon (clamd) over TCP.
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
	d := net.Dialer{Timeout: c.timeout}
	conn, err := d.DialContext(ctx, "tcp", c.address)
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

	var sb strings.Builder
	buf := make([]byte, 512)
	for {
		n, err := conn.Read(buf)
		if n > 0 {
			sb.Write(bytes.TrimRight(buf[:n], "\x00\r\n"))
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return sb.String(), nil
		}
		if sb.Len() > 0 {
			break
		}
	}

	ver := strings.TrimSpace(sb.String())
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
				return nil, fmt.Errorf("failed to write chunk length: %w", err)
			}
			if _, err := conn.Write(buf[:n]); err != nil {
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
		return nil, fmt.Errorf("failed to write EOF chunk: %w", err)
	}

	// Read response
	var sb strings.Builder
	resBuf := make([]byte, 1024)
	for {
		n, err := conn.Read(resBuf)
		if n > 0 {
			sb.Write(bytes.TrimRight(resBuf[:n], "\x00\r\n"))
		}
		if err != nil {
			break
		}
		if sb.Len() > 0 {
			break
		}
	}

	raw := strings.TrimSpace(sb.String())
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

	return &ClamAVResult{
		Infected: false,
		Raw:      raw,
	}, nil
}
