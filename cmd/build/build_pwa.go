// build_pwa.go implements PWA and legacy GUI build functionality.
//
// Supports building desktop applications from:
//   - Local static web application directories
//   - Live PWA URLs (downloads and packages)

package build

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/host-uk/core/pkg/i18n"
	"github.com/leaanthony/debme"
	"github.com/leaanthony/gosod"
	"golang.org/x/net/html"
)

// Error sentinels for build commands
var (
	errPathRequired = errors.New("the --path flag is required")
	errURLRequired  = errors.New("the --url flag is required")
)

// runPwaBuild downloads a PWA from URL and builds it.
func runPwaBuild(pwaURL string) error {
	fmt.Printf("%s %s\n", i18n.T("cmd.build.pwa.starting"), pwaURL)

	tempDir, err := os.MkdirTemp("", "core-pwa-build-*")
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.T("cmd.build.pwa.error.create_temp_dir"), err)
	}
	// defer os.RemoveAll(tempDir) // Keep temp dir for debugging
	fmt.Printf("%s %s\n", i18n.T("cmd.build.pwa.downloading_to"), tempDir)

	if err := downloadPWA(pwaURL, tempDir); err != nil {
		return fmt.Errorf("%s: %w", i18n.T("cmd.build.pwa.error.download_failed"), err)
	}

	return runBuild(tempDir)
}

// downloadPWA fetches a PWA from a URL and saves assets locally.
func downloadPWA(baseURL, destDir string) error {
	// Fetch the main HTML page
	resp, err := http.Get(baseURL)
	if err != nil {
		return fmt.Errorf("%s %s: %w", i18n.T("cmd.build.pwa.error.fetch_url"), baseURL, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.T("cmd.build.pwa.error.read_response"), err)
	}

	// Find the manifest URL from the HTML
	manifestURL, err := findManifestURL(string(body), baseURL)
	if err != nil {
		// If no manifest, it's not a PWA, but we can still try to package it as a simple site.
		fmt.Printf("%s %s\n", i18n.T("cmd.build.pwa.warning"), i18n.T("cmd.build.pwa.no_manifest"))
		if err := os.WriteFile(filepath.Join(destDir, "index.html"), body, 0644); err != nil {
			return fmt.Errorf("%s: %w", i18n.T("cmd.build.pwa.error.write_index"), err)
		}
		return nil
	}

	fmt.Printf("%s %s\n", i18n.T("cmd.build.pwa.found_manifest"), manifestURL)

	// Fetch and parse the manifest
	manifest, err := fetchManifest(manifestURL)
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.T("cmd.build.pwa.error.fetch_manifest"), err)
	}

	// Download all assets listed in the manifest
	assets := collectAssets(manifest, manifestURL)
	for _, assetURL := range assets {
		if err := downloadAsset(assetURL, destDir); err != nil {
			fmt.Printf("%s %s %s: %v\n", i18n.T("cmd.build.pwa.warning"), i18n.T("cmd.build.pwa.asset_download_failed"), assetURL, err)
		}
	}

	// Also save the root index.html
	if err := os.WriteFile(filepath.Join(destDir, "index.html"), body, 0644); err != nil {
		return fmt.Errorf("%s: %w", i18n.T("cmd.build.pwa.error.write_index"), err)
	}

	fmt.Println(i18n.T("cmd.build.pwa.download_complete"))
	return nil
}

// findManifestURL extracts the manifest URL from HTML content.
func findManifestURL(htmlContent, baseURL string) (string, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return "", err
	}

	var manifestPath string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "link" {
			var rel, href string
			for _, a := range n.Attr {
				if a.Key == "rel" {
					rel = a.Val
				}
				if a.Key == "href" {
					href = a.Val
				}
			}
			if rel == "manifest" && href != "" {
				manifestPath = href
				return
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	if manifestPath == "" {
		return "", fmt.Errorf("%s", i18n.T("cmd.build.pwa.error.no_manifest_tag"))
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	manifestURL, err := base.Parse(manifestPath)
	if err != nil {
		return "", err
	}

	return manifestURL.String(), nil
}

// fetchManifest downloads and parses a PWA manifest.
func fetchManifest(manifestURL string) (map[string]interface{}, error) {
	resp, err := http.Get(manifestURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var manifest map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return nil, err
	}
	return manifest, nil
}

// collectAssets extracts asset URLs from a PWA manifest.
func collectAssets(manifest map[string]interface{}, manifestURL string) []string {
	var assets []string
	base, _ := url.Parse(manifestURL)

	// Add start_url
	if startURL, ok := manifest["start_url"].(string); ok {
		if resolved, err := base.Parse(startURL); err == nil {
			assets = append(assets, resolved.String())
		}
	}

	// Add icons
	if icons, ok := manifest["icons"].([]interface{}); ok {
		for _, icon := range icons {
			if iconMap, ok := icon.(map[string]interface{}); ok {
				if src, ok := iconMap["src"].(string); ok {
					if resolved, err := base.Parse(src); err == nil {
						assets = append(assets, resolved.String())
					}
				}
			}
		}
	}

	return assets
}

// downloadAsset fetches a single asset and saves it locally.
func downloadAsset(assetURL, destDir string) error {
	resp, err := http.Get(assetURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	u, err := url.Parse(assetURL)
	if err != nil {
		return err
	}

	path := filepath.Join(destDir, filepath.FromSlash(u.Path))
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return err
	}

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// runBuild builds a desktop application from a local directory.
func runBuild(fromPath string) error {
	fmt.Printf("%s %s\n", i18n.T("cmd.build.from_path.starting"), fromPath)

	info, err := os.Stat(fromPath)
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.T("cmd.build.from_path.error.invalid_path"), err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s", i18n.T("cmd.build.from_path.error.must_be_directory"))
	}

	buildDir := ".core/build/app"
	htmlDir := filepath.Join(buildDir, "html")
	appName := filepath.Base(fromPath)
	if strings.HasPrefix(appName, "core-pwa-build-") {
		appName = "pwa-app"
	}
	outputExe := appName

	if err := os.RemoveAll(buildDir); err != nil {
		return fmt.Errorf("%s: %w", i18n.T("cmd.build.from_path.error.clean_build_dir"), err)
	}

	// 1. Generate the project from the embedded template
	fmt.Println(i18n.T("cmd.build.from_path.generating_template"))
	templateFS, err := debme.FS(guiTemplate, "tmpl/gui")
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.T("cmd.build.from_path.error.anchor_template"), err)
	}
	sod := gosod.New(templateFS)
	if sod == nil {
		return fmt.Errorf("%s", i18n.T("cmd.build.from_path.error.create_sod"))
	}

	templateData := map[string]string{"AppName": appName}
	if err := sod.Extract(buildDir, templateData); err != nil {
		return fmt.Errorf("%s: %w", i18n.T("cmd.build.from_path.error.extract_template"), err)
	}

	// 2. Copy the user's web app files
	fmt.Println(i18n.T("cmd.build.from_path.copying_files"))
	if err := copyDir(fromPath, htmlDir); err != nil {
		return fmt.Errorf("%s: %w", i18n.T("cmd.build.from_path.error.copy_files"), err)
	}

	// 3. Compile the application
	fmt.Println(i18n.T("cmd.build.from_path.compiling"))

	// Run go mod tidy
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = buildDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %w", i18n.T("cmd.build.from_path.error.go_mod_tidy"), err)
	}

	// Run go build
	cmd = exec.Command("go", "build", "-o", outputExe)
	cmd.Dir = buildDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %w", i18n.T("cmd.build.from_path.error.go_build"), err)
	}

	fmt.Printf("\n%s %s/%s\n", i18n.T("cmd.build.from_path.success"), buildDir, outputExe)
	return nil
}

// copyDir recursively copies a directory from src to dst.
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		dstFile, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		_, err = io.Copy(dstFile, srcFile)
		return err
	})
}
