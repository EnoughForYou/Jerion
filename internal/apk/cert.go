package apk

import (
	"archive/zip"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"io"
	"math/big"
	"path/filepath"
	"strings"
	"time"
)

// ParseCertificateFingerprints scans META-INF/ within an APK zip and calculates SHA-256 fingerprints.
func ParseCertificateFingerprints(zr *zip.Reader) ([]string, error) {
	var fingerprints []string

	for _, f := range zr.File {
		ext := strings.ToUpper(filepath.Ext(f.Name))
		if strings.HasPrefix(f.Name, "META-INF/") && (ext == ".RSA" || ext == ".DSA" || ext == ".EC" || ext == ".CERT") {
			rc, err := f.Open()
			if err != nil {
				continue
			}
			data, err := io.ReadAll(rc)
			_ = rc.Close()
			if err != nil {
				continue
			}

			// Attempt parsing as raw certificates
			certs, err := x509.ParseCertificates(data)
			if err == nil && len(certs) > 0 {
				for _, c := range certs {
					h := sha256.Sum256(c.Raw)
					fp := strings.ToUpper(hex.EncodeToString(h[:]))
					if !containsString(fingerprints, fp) {
						fingerprints = append(fingerprints, fp)
					}
				}
				continue
			}

			// If wrapped in PKCS#7 / CMS, scan for embedded certificate sequences
			// Fallback: extract certificate by searching for SEQUENCE tags with matching signatures
			if fp := extractCertFallback(data); fp != "" {
				if !containsString(fingerprints, fp) {
					fingerprints = append(fingerprints, fp)
				}
			}
		}
	}

	return fingerprints, nil
}

func extractCertFallback(data []byte) string {
	// Look for DER encoded X509 certificates inside PKCS#7 envelope
	for i := 0; i < len(data)-4; i++ {
		if data[i] == 0x30 && data[i+1] == 0x82 {
			length := int(data[i+2])<<8 | int(data[i+3])
			if i+4+length <= len(data) {
				candidate := data[i : i+4+length]
				cert, err := x509.ParseCertificate(candidate)
				if err == nil && cert != nil {
					h := sha256.Sum256(cert.Raw)
					return strings.ToUpper(hex.EncodeToString(h[:]))
				}
			}
		}
	}
	return ""
}

// GenerateTestCertificate creates an in-memory DER certificate and its SHA-256 fingerprint for testing.
func GenerateTestCertificate(org string) ([]byte, string, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, "", err
	}

	serialNumber, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 62))
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{org},
			CommonName:   "Jerion Synthetic Test Signer",
		},
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, "", err
	}

	h := sha256.Sum256(derBytes)
	fp := strings.ToUpper(hex.EncodeToString(h[:]))

	return derBytes, fp, nil
}

func containsString(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}
