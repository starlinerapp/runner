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
)

const version = "v1.16.0"

func Ensure() error {
	if _, err := exec.LookPath("firecracker"); err == nil {
		return nil
	}

	url := downloadURL(version, runtime.GOARCH)

	tmpDir, err := os.MkdirTemp("", "firecracker-*")
	if err != nil {
		return err
	}
	defer func(path string) {
		_ = os.RemoveAll(path)
	}(tmpDir)

	archivePath := filepath.Join(tmpDir, "firecracker.tgz")

	fmt.Println("Downloading Firecracker from ", url)
	if err := downloadFile(url, archivePath); err != nil {
		return err
	}

	fmt.Println("Extracting...")
	if err := extractBinary(archivePath, "/usr/local/bin/firecracker"); err != nil {
		return err
	}

	fmt.Println("Firecracker installed successfully")
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

func extractBinary(tgzPath, target string) error {
	file, err := os.Open(tgzPath)
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

		if filepath.Base(hdr.Name) == "firecracker" {
			out, err := os.Create(target)
			if err != nil {
				return err
			}
			defer func(out *os.File) {
				_ = out.Close()
			}(out)

			if _, err := io.Copy(out, tr); err != nil {
				return err
			}
			return os.Chmod(target, 0755)
		}
	}
	return fmt.Errorf("firecracker binary not found in %s", tgzPath)
}

func downloadURL(version, arch string) string {
	return fmt.Sprintf(
		"https://github.com/firecracker-microvm/firecracker/releases/download/%s/firecracker-%s-%s.tgz",
		version,
		version,
		arch,
	)
}
