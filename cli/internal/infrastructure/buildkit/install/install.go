package install

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"starliner.app/runner/internal/infrastructure/privileged"
)

const version = "v0.30.0"

var binaries = []string{"buildkitd", "buildctl"}

func Ensure() error {
	if _, err := exec.LookPath("buildkitd"); err == nil {
		return nil
	}

	url, err := downloadURL()
	if err != nil {
		return err
	}

	tmpDir, err := os.MkdirTemp("", "buildkit-*")
	if err != nil {
		return err
	}
	defer func(path string) {
		_ = os.RemoveAll(path)
	}(tmpDir)

	archivePath := filepath.Join(tmpDir, "buildkit.tar.gz")

	fmt.Println("Downloading BuildKit from", url)
	if err := downloadFile(url, archivePath); err != nil {
		return err
	}

	fmt.Println("Extracting...")
	for _, name := range binaries {
		target := filepath.Join("/usr/local/bin", name)
		if err := extractBinary(archivePath, name, target); err != nil {
			return err
		}
	}

	fmt.Println("BuildKit installed successfully")
	return nil
}

func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s: HTTP %s", url, resp.Status)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer func(out *os.File) {
		_ = out.Close()
	}(out)

	_, err = io.Copy(out, resp.Body)
	return err
}

func extractBinary(archivePath, binaryName, target string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer func() {
		_ = gzr.Close()
	}()

	tr := tar.NewReader(gzr)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if filepath.Base(hdr.Name) != binaryName {
			continue
		}

		tmp, err := os.CreateTemp("", "buildkit-*")
		if err != nil {
			return err
		}
		tmpPath := tmp.Name()

		if _, err := io.Copy(tmp, tr); err != nil {
			_ = tmp.Close()
			_ = os.Remove(tmpPath)
			return err
		}

		if err := tmp.Close(); err != nil {
			_ = os.Remove(tmpPath)
			return err
		}

		if err := privileged.Run("install", "-m", "0755", tmpPath, target); err != nil {
			_ = os.Remove(tmpPath)
			return err
		}

		_ = os.Remove(tmpPath)
		return nil
	}

	return fmt.Errorf("buildkit binary %q not found in %s", binaryName, archivePath)
}

func releasePlatform() (string, error) {
	switch runtime.GOARCH {
	case "amd64":
		return "linux-amd64", nil
	case "arm64":
		return "linux-arm64", nil
	default:
		return "", fmt.Errorf("unsupported architecture %q", runtime.GOARCH)
	}
}

func downloadURL() (string, error) {
	platform, err := releasePlatform()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(
		"https://github.com/moby/buildkit/releases/download/%s/buildkit-%s.%s.tar.gz",
		version,
		version,
		platform,
	), nil
}
