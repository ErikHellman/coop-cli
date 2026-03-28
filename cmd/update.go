package cmd

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

type githubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update coop-cli to the latest release",
	RunE: func(cmd *cobra.Command, args []string) error {
		release, err := fetchLatestRelease()
		if err != nil {
			return fmt.Errorf("failed to check for updates: %w", err)
		}

		latestVersion := strings.TrimPrefix(release.TagName, "v")
		currentVersion := strings.TrimPrefix(Version, "v")

		if currentVersion == latestVersion {
			fmt.Printf("Already up to date (%s).\n", Version)
			return nil
		}

		fmt.Printf("Updating from %s to %s...\n", Version, release.TagName)

		assetName := buildAssetName(latestVersion)
		var downloadURL string
		for _, a := range release.Assets {
			if a.Name == assetName {
				downloadURL = a.BrowserDownloadURL
				break
			}
		}
		if downloadURL == "" {
			return fmt.Errorf("no release asset found for %s/%s (%s)", runtime.GOOS, runtime.GOARCH, assetName)
		}

		binary, err := downloadAndExtract(downloadURL, assetName)
		if err != nil {
			return fmt.Errorf("failed to download update: %w", err)
		}
		defer os.Remove(binary)

		execPath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("failed to find current executable: %w", err)
		}
		execPath, err = filepath.EvalSymlinks(execPath)
		if err != nil {
			return fmt.Errorf("failed to resolve executable path: %w", err)
		}

		if err := replaceExecutable(binary, execPath); err != nil {
			return fmt.Errorf("failed to replace executable: %w", err)
		}

		fmt.Printf("Updated to %s.\n", release.TagName)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

func fetchLatestRelease() (*githubRelease, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}
	return &release, nil
}

func buildAssetName(version string) string {
	ext := "tar.gz"
	if runtime.GOOS == "windows" {
		ext = "zip"
	}
	return fmt.Sprintf("coop-cli_%s_%s_%s.%s", version, runtime.GOOS, runtime.GOARCH, ext)
}

func downloadAndExtract(url, assetName string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download returned status %d", resp.StatusCode)
	}

	tmpDir, err := os.MkdirTemp("", "coop-cli-update-*")
	if err != nil {
		return "", err
	}

	archivePath := filepath.Join(tmpDir, assetName)
	f, err := os.Create(archivePath)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		return "", err
	}
	f.Close()

	binaryName := "coop-cli"
	if runtime.GOOS == "windows" {
		binaryName = "coop-cli.exe"
	}

	if strings.HasSuffix(assetName, ".tar.gz") {
		return extractTarGz(archivePath, tmpDir, binaryName)
	}
	return extractZip(archivePath, tmpDir, binaryName)
}

func extractTarGz(archivePath, destDir, binaryName string) (string, error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return "", err
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		if filepath.Base(hdr.Name) == binaryName {
			outPath := filepath.Join(destDir, binaryName)
			out, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY, 0755)
			if err != nil {
				return "", err
			}
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				return "", err
			}
			out.Close()
			return outPath, nil
		}
	}
	return "", fmt.Errorf("binary %s not found in archive", binaryName)
}

func extractZip(archivePath, destDir, binaryName string) (string, error) {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", err
	}
	defer r.Close()

	for _, zf := range r.File {
		if filepath.Base(zf.Name) == binaryName {
			rc, err := zf.Open()
			if err != nil {
				return "", err
			}
			outPath := filepath.Join(destDir, binaryName)
			out, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY, 0755)
			if err != nil {
				rc.Close()
				return "", err
			}
			if _, err := io.Copy(out, rc); err != nil {
				out.Close()
				rc.Close()
				return "", err
			}
			out.Close()
			rc.Close()
			return outPath, nil
		}
	}
	return "", fmt.Errorf("binary %s not found in archive", binaryName)
}

func replaceExecutable(newBinary, target string) error {
	// Rename current executable out of the way, then move new one in.
	// This avoids "text file busy" on Linux.
	backup := target + ".old"
	if err := os.Rename(target, backup); err != nil {
		return fmt.Errorf("backing up current binary: %w", err)
	}

	src, err := os.Open(newBinary)
	if err != nil {
		os.Rename(backup, target) // restore
		return err
	}
	defer src.Close()

	dst, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		os.Rename(backup, target) // restore
		return err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		os.Rename(backup, target) // restore
		return err
	}

	os.Remove(backup)
	return nil
}
