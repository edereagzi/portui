package update

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const repo = "edereagzi/portui"

var httpClient = &http.Client{Timeout: 30 * time.Second}

type githubRelease struct {
	TagName string `json:"tag_name"`
}

func Run(currentVersion string) (string, bool, error) {
	osName, arch, err := platform()
	if err != nil {
		return "", false, err
	}

	latestTag, err := latestTag()
	if err != nil {
		return "", false, err
	}

	if currentVersion != "dev" && sameVersion(currentVersion, latestTag) {
		return latestTag, false, nil
	}

	assetName := releaseAssetName(osName, arch)
	base := fmt.Sprintf("https://github.com/%s/releases/download/%s", repo, latestTag)
	assetURL := fmt.Sprintf("%s/%s", base, assetName)
	checksumsURL := fmt.Sprintf("%s/checksums.txt", base)

	archiveBytes, err := download(assetURL)
	if err != nil {
		return "", false, fmt.Errorf("download release asset: %w", err)
	}
	checksumsBytes, err := download(checksumsURL)
	if err != nil {
		return "", false, fmt.Errorf("download checksums: %w", err)
	}

	expected, err := checksumForAsset(string(checksumsBytes), assetName)
	if err != nil {
		return "", false, err
	}
	actual := sha256Sum(archiveBytes)
	if expected != actual {
		return "", false, fmt.Errorf("checksum mismatch for %s", assetName)
	}

	binaryBytes, err := extractBinary(osName, archiveBytes)
	if err != nil {
		return "", false, err
	}

	if err := replaceCurrentExecutable(binaryBytes); err != nil {
		return "", false, err
	}

	return latestTag, true, nil
}

func latestTag() (string, error) {
	respBytes, err := download("https://api.github.com/repos/" + repo + "/releases/latest")
	if err != nil {
		return "", fmt.Errorf("fetch latest release: %w", err)
	}
	var release githubRelease
	if err := json.Unmarshal(respBytes, &release); err != nil {
		return "", fmt.Errorf("parse release metadata: %w", err)
	}
	if release.TagName == "" {
		return "", errors.New("latest release tag is empty")
	}
	return release.TagName, nil
}

func platform() (string, string, error) {
	var osName string
	switch runtime.GOOS {
	case "darwin", "linux", "windows":
		osName = runtime.GOOS
	default:
		return "", "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	var arch string
	switch runtime.GOARCH {
	case "amd64", "arm64":
		arch = runtime.GOARCH
	default:
		return "", "", fmt.Errorf("unsupported architecture: %s", runtime.GOARCH)
	}

	return osName, arch, nil
}

func releaseAssetName(osName, arch string) string {
	if osName == "windows" {
		return fmt.Sprintf("portui_%s_%s.zip", osName, arch)
	}
	return fmt.Sprintf("portui_%s_%s.tar.gz", osName, arch)
}

func sameVersion(current, latest string) bool {
	return strings.TrimPrefix(current, "v") == strings.TrimPrefix(latest, "v")
}

func download(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "portui-self-update")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("download failed: %s", resp.Status)
	}

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func checksumForAsset(checksums, asset string) (string, error) {
	for _, line := range strings.Split(checksums, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		if fields[1] == asset {
			return fields[0], nil
		}
	}
	return "", fmt.Errorf("checksum not found for %s", asset)
}

func sha256Sum(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func extractBinary(osName string, archive []byte) ([]byte, error) {
	if osName == "windows" {
		return extractBinaryZip(archive)
	}
	return extractBinaryTarGz(archive)
}

func extractBinaryTarGz(archive []byte) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		if filepath.Base(hdr.Name) == "portui" {
			return io.ReadAll(tr)
		}
	}

	return nil, errors.New("portui binary not found in archive")
}

func extractBinaryZip(archive []byte) ([]byte, error) {
	r, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		return nil, err
	}

	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		if filepath.Base(f.Name) != "portui.exe" {
			continue
		}
		rc, openErr := f.Open()
		if openErr != nil {
			return nil, openErr
		}
		defer rc.Close()
		return io.ReadAll(rc)
	}

	return nil, errors.New("portui binary not found in archive")
}

func replaceCurrentExecutable(binary []byte) error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	tempPath := exePath + ".new"

	if err := os.WriteFile(tempPath, binary, 0o755); err != nil {
		return fmt.Errorf("write temp binary: %w", err)
	}

	if runtime.GOOS == "windows" {
		if err := scheduleWindowsReplacement(exePath, tempPath); err != nil {
			_ = os.Remove(tempPath)
			return err
		}
		return nil
	}

	if err := os.Rename(tempPath, exePath); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("replace executable: %w", err)
	}
	return nil
}

func scheduleWindowsReplacement(exePath, tempPath string) error {
	backupPath := exePath + ".old"
	scriptPath := exePath + ".update.cmd"
	script := windowsUpdateScript(exePath, tempPath, backupPath)

	if err := os.WriteFile(scriptPath, []byte(script), 0o600); err != nil {
		return fmt.Errorf("write update script: %w", err)
	}

	cmd := exec.Command("cmd", "/C", "start", "", "/B", scriptPath)
	if err := cmd.Start(); err != nil {
		_ = os.Remove(scriptPath)
		return fmt.Errorf("launch update script: %w", err)
	}

	return nil
}

func windowsUpdateScript(exePath, tempPath, backupPath string) string {
	return fmt.Sprintf("@echo off\r\nsetlocal\r\nset \"DST=%s\"\r\nset \"NEW=%s\"\r\nset \"BAK=%s\"\r\nfor /l %%%%i in (1,1,30) do (\r\n  move /Y \"%%DST%%\" \"%%BAK%%\" >nul 2>nul && goto replaced\r\n  timeout /t 1 /nobreak >nul\r\n)\r\nexit /b 1\r\n:replaced\r\nmove /Y \"%%NEW%%\" \"%%DST%%\" >nul 2>nul && goto done\r\nmove /Y \"%%BAK%%\" \"%%DST%%\" >nul 2>nul\r\nexit /b 1\r\n:done\r\ndel /f /q \"%%BAK%%\" >nul 2>nul\r\ndel /f /q \"%%~f0\" >nul 2>nul\r\n", exePath, tempPath, backupPath)
}
