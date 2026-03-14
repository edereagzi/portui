package update

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestChecksumForAsset(t *testing.T) {
	checksums := "abc123  portui_darwin_arm64.tar.gz\ndef456  another.tar.gz"

	got, err := checksumForAsset(checksums, "portui_darwin_arm64.tar.gz")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "abc123" {
		t.Fatalf("expected abc123, got %q", got)
	}
}

func TestChecksumForAssetNotFound(t *testing.T) {
	checksums := "abc123  portui_darwin_arm64.tar.gz"

	_, err := checksumForAsset(checksums, "missing.tar.gz")
	if err == nil {
		t.Fatal("expected error when checksum entry is missing")
	}
}

func TestSameVersion(t *testing.T) {
	if !sameVersion("v1.2.3", "1.2.3") {
		t.Fatal("expected versions to be treated as same")
	}
	if sameVersion("1.2.3", "1.2.4") {
		t.Fatal("expected different versions to not match")
	}
}

func TestExtractBinary(t *testing.T) {
	archive := buildTarGz(t, map[string]string{
		"README.md": "hello",
		"portui":    "binary-content",
	})

	got, err := extractBinary(archive)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != "binary-content" {
		t.Fatalf("unexpected binary content: %q", string(got))
	}
}

func TestExtractBinaryNotFound(t *testing.T) {
	archive := buildTarGz(t, map[string]string{
		"README.md": "hello",
	})

	_, err := extractBinary(archive)
	if err == nil {
		t.Fatal("expected error when binary is missing")
	}
}

func TestDownloadSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	originalClient := httpClient
	httpClient = server.Client()
	t.Cleanup(func() {
		httpClient = originalClient
	})

	got, err := download(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != "ok" {
		t.Fatalf("expected 'ok', got %q", string(got))
	}
}

func TestDownloadNon2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("not found"))
	}))
	defer server.Close()

	originalClient := httpClient
	httpClient = server.Client()
	t.Cleanup(func() {
		httpClient = originalClient
	})

	_, err := download(server.URL)
	if err == nil {
		t.Fatal("expected error for non-2xx response")
	}
}

func buildTarGz(t *testing.T, files map[string]string) []byte {
	t.Helper()

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)

	for name, content := range files {
		hdr := &tar.Header{Name: name, Mode: 0o644, Size: int64(len(content))}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatalf("write header: %v", err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatalf("write body: %v", err)
		}
	}

	if err := tw.Close(); err != nil {
		t.Fatalf("close tar writer: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("close gzip writer: %v", err)
	}

	return buf.Bytes()
}
