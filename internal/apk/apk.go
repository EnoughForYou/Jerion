package apk

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// APKDetail captures complete forensic inspection results for an APK.
type APKDetail struct {
	FilePath         string            `json:"filePath"`
	SHA256           string            `json:"sha256"`
	FileSize         int64             `json:"fileSize"`
	Manifest         ManifestData      `json:"manifest"`
	CertFingerprints []string          `json:"certFingerprints"`
	DetectedAdSDKs   map[string]string `json:"detectedAdSdks"`
	DEXStringCount   int               `json:"dexStringCount"`
}

// Inspector performs static inspection of Android package files.
type Inspector struct{}

// NewInspector creates an APK Inspector.
func NewInspector() *Inspector {
	return &Inspector{}
}

// Inspect opens an APK archive and extracts manifest, code strings, and certificates.
func (insp *Inspector) Inspect(apkPath string) (*APKDetail, error) {
	fi, err := os.Stat(apkPath)
	if err != nil {
		return nil, fmt.Errorf("stat apk: %w", err)
	}

	// Compute file SHA256
	fileSHA, err := computeSHA256(apkPath)
	if err != nil {
		return nil, fmt.Errorf("hash apk: %w", err)
	}

	zr, err := zip.OpenReader(apkPath)
	if err != nil {
		return nil, fmt.Errorf("open apk zip: %w", err)
	}
	defer zr.Close()

	detail := &APKDetail{
		FilePath:       apkPath,
		SHA256:         fileSHA,
		FileSize:       fi.Size(),
		DetectedAdSDKs: make(map[string]string),
	}

	// 1. Parse AndroidManifest.xml
	for _, f := range zr.File {
		if f.Name == "AndroidManifest.xml" {
			rc, err := f.Open()
			if err == nil {
				manifest, parseErr := ParseAXML(rc)
				_ = rc.Close()
				if parseErr == nil && manifest != nil {
					detail.Manifest = *manifest
				}
			}
			break
		}
	}

	// 2. Parse DEX files (classes.dex, classes2.dex...)
	for _, f := range zr.File {
		if strings.HasPrefix(f.Name, "classes") && strings.HasSuffix(f.Name, ".dex") {
			rc, err := f.Open()
			if err == nil {
				dexAnalysis, dexErr := ParseDEX(rc)
				_ = rc.Close()
				if dexErr == nil && dexAnalysis != nil {
					detail.DEXStringCount += dexAnalysis.TotalStrings
					for k, v := range dexAnalysis.DetectedAdSDKs {
						detail.DetectedAdSDKs[k] = v
					}
				}
			}
		}
	}

	// 3. Extract Certificate Fingerprints
	certs, err := ParseCertificateFingerprints(&zr.Reader)
	if err == nil {
		detail.CertFingerprints = certs
	}

	return detail, nil
}

func computeSHA256(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// CreateSyntheticAPK generates a valid .apk file on disk for unit testing.
func CreateSyntheticAPK(destPath, pkg, versionName string, versionCode int64, permissions, dexStrings []string, certOrg string) (string, string, error) {
	if err := os.MkdirAll(filepath.Dir(destPath), 0o750); err != nil {
		return "", "", err
	}

	f, err := os.Create(destPath)
	if err != nil {
		return "", "", err
	}
	defer f.Close()

	zw := zip.NewWriter(f)

	// 1. AndroidManifest.xml
	axmlBytes := GenerateSyntheticAXML(pkg, versionName, versionCode, permissions)
	w, err := zw.Create("AndroidManifest.xml")
	if err != nil {
		return "", "", err
	}
	if _, err := w.Write(axmlBytes); err != nil {
		return "", "", err
	}

	// 2. classes.dex
	dexBytes := GenerateSyntheticDEX(dexStrings)
	w, err = zw.Create("classes.dex")
	if err != nil {
		return "", "", err
	}
	if _, err := w.Write(dexBytes); err != nil {
		return "", "", err
	}

	// 3. META-INF certificate
	certBytes, certFP, err := GenerateTestCertificate(certOrg)
	if err != nil {
		return "", "", err
	}
	w, err = zw.Create("META-INF/CERT.RSA")
	if err != nil {
		return "", "", err
	}
	if _, err := w.Write(certBytes); err != nil {
		return "", "", err
	}

	if err := zw.Close(); err != nil {
		return "", "", err
	}
	_ = f.Close()

	fileSHA, err := computeSHA256(destPath)
	if err != nil {
		return "", "", err
	}

	return certFP, fileSHA, nil
}
