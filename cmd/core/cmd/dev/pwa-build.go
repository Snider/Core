package dev

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/leaanthony/clir"
	"github.com/leaanthony/debme"
	"github.com/leaanthony/gosod"
	"golang.org/x/net/html"
)

//go:embed all:tmpl/gui
var guiTemplate embed.FS

// AddBuildCommand adds the new build command and its subcommands to the clir app.
func AddBuildCommand(parent *clir.Command) {
	buildCmd := parent.NewSubCommand("build", "Builds a web application into a standalone desktop app.")

	// --- `build from-path` command ---
	fromPathCmd := buildCmd.NewSubCommand("from-path", "Build from a local directory.")
	var fromPath string
	fromPathCmd.StringFlag("path", "The path to the static web application files.", &fromPath)
	fromPathCmd.Action(func() error {
		if fromPath == "" {
			return fmt.Errorf("the --path flag is required")
		}
		return runBuild(fromPath)
	})

	// --- `build pwa` command ---
	pwaCmd := buildCmd.NewSubCommand("pwa", "Build from a live PWA URL.")
	var pwaURL string
	pwaCmd.StringFlag("url", "The URL of the PWA to build.", &pwaURL)
	pwaCmd.Action(func() error {
		if pwaURL == "" {
			return fmt.Errorf("a URL argument is required")
		}
		return runPwaBuild(pwaURL)
	})
}

// --- PWA Build Logic ---

func runPwaBuild(pwaURL string) error {
	fmt.Printf("Starting PWA build from URL: %s\n", pwaURL)

	tempDir, err := os.MkdirTemp("", "core-pwa-build-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}
	// defer os.RemoveAll(tempDir) // Keep temp dir for debugging
	fmt.Printf("Downloading PWA to temporary directory: %s\n", tempDir)

	if err := downloadPWA(pwaURL, tempDir); err != nil {
		return fmt.Errorf("failed to download PWA: %w", err)
	}

	return runBuild(tempDir)
}

func downloadPWA(baseURL, destDir string) error {
	// Fetch the main HTML page
	resp, err := http.Get(baseURL)
	if err != nil {
		return fmt.Errorf("failed to fetch URL %s: %w", baseURL, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Find the manifest URL from the HTML
	//manifestURL, err := findManifestURL(string(body), baseURL)
	//if err != nil {
	//	// If no manifest, it's not a PWA, but we can still try to package it as a simple site.
	//	fmt.Println("Warning: no manifest file found. Proceeding with basic site download.")
	//	if err := os.WriteFile(filepath.Join(destDir, "index.html"), body, 0644); err != nil {
	//		return fmt.Errorf("failed to write index.html: %w", err)
	//	}
	//	return nil
	//}
	//
	//fmt.Printf("Found manifest: %s\n", manifestURL)
	//
	//// Fetch and parse the manifest
	//manifest, err := fetchManifest(baseURL, manifestURL)
	//if err != nil {
	//	return fmt.Errorf("failed to fetch or parse manifest: %w", err)
	//}

	// Download all assets listed in the manifest
	//assets := collectAssets(manifest, manifestURL)
	//for _, assetURL := range assets {
	//	if err := downloadAsset(assetURL, destDir); err != nil {
	//		fmt.Printf("Warning: failed to download asset %s: %v\n", assetURL, err)
	//	}
	//}

	// Also save the root index.html
	if err := os.WriteFile(filepath.Join(destDir, "index.html"), body, 0644); err != nil {
		return fmt.Errorf("failed to write index.html: %w", err)
	}

	fmt.Println("PWA download complete.")
	return nil
}

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
		return "", fmt.Errorf("no <link rel=\"manifest\"> tag found")
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

func fetchManifest(manifest map[string]interface{}, manifestURL string) (map[string]interface{}, error) {
	resp, err := http.Get(manifestURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var manifestData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&manifestData); err != nil {
		return nil, err
	}
	return manifestData, nil
}

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

// --- Standard Build Logic ---

func runBuild(fromPath string) error {
	fmt.Printf("Starting build from path: %s\n", fromPath)

	info, err := os.Stat(fromPath)
	if err != nil {
		return fmt.Errorf("invalid path specified: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("path specified must be a directory")
	}

	buildDir := ".core/build/app"
	htmlDir := filepath.Join(buildDir, "html")
	appName := filepath.Base(fromPath)
	if strings.HasPrefix(appName, "core-pwa-build-") {
		appName = "pwa-app"
	}
	outputExe := appName

	if err := os.RemoveAll(buildDir); err != nil {
		return fmt.Errorf("failed to clean build directory: %w", err)
	}

	// 1. Generate the project from the embedded template
	fmt.Println("Generating application from template...")
	templateFS, err := debme.FS(guiTemplate, "tmpl/gui")
	if err != nil {
		return fmt.Errorf("failed to anchor template filesystem: %w", err)
	}
	sod := gosod.New(templateFS)
	if sod != nil {
		return fmt.Errorf("failed to create new sod instance: %w", sod)
	}

	templateData := map[string]string{"AppName": appName}
	if err := sod.Extract(buildDir, templateData); err != nil {
		return fmt.Errorf("failed to extract template: %w", err)
	}

	// 2. Copy the user's web app files
	fmt.Println("Copying application files...")
	if err := copyDir(fromPath, htmlDir); err != nil {
		return fmt.Errorf("failed to copy application files: %w", err)
	}

	// 3. Compile the application
	fmt.Println("Compiling application...")

	// Run go mod tidy
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = buildDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go mod tidy failed: %w", err)
	}

	// Run go build
	cmd = exec.Command("go", "build", "-o", outputExe)
	cmd.Dir = buildDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go build failed: %w", err)
	}

	fmt.Printf("\nBuild successful! Executable created at: %s/%s\n", buildDir, outputExe)
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
