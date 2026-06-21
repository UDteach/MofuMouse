package update

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type Release struct {
	TagName string  `json:"tag_name"`
	Name    string  `json:"name"`
	HTMLURL string  `json:"html_url"`
	Assets  []Asset `json:"assets"`
}

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type Result struct {
	CurrentVersion  string
	LatestVersion   string
	ReleaseURL      string
	AssetName       string
	AssetURL        string
	UpdateAvailable bool
}

type DownloadResult struct {
	AssetName string
	Path      string
	Bytes     int64
}

func CheckLatest(repo, currentVersion string) (Result, error) {
	repo = strings.TrimSpace(repo)
	if repo == "" {
		return Result{}, fmt.Errorf("update repo is empty")
	}
	url := "https://api.github.com/repos/" + repo + "/releases/latest"
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return Result{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "MofuMouse")
	resp, err := client.Do(req)
	if err != nil {
		return Result{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Result{}, fmt.Errorf("github releases request failed: %s", resp.Status)
	}
	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return Result{}, err
	}
	assetName, assetURL := MatchingAssetForGOARCH(release.Assets, runtime.GOARCH)
	return Result{
		CurrentVersion:  currentVersion,
		LatestVersion:   release.TagName,
		ReleaseURL:      release.HTMLURL,
		AssetName:       assetName,
		AssetURL:        assetURL,
		UpdateAvailable: CompareVersions(release.TagName, currentVersion) > 0,
	}, nil
}

func matchingAsset(assets []Asset) (string, string) {
	return MatchingAssetForGOARCH(assets, runtime.GOARCH)
}

func MatchingAssetForGOARCH(assets []Asset, goarch string) (string, string) {
	bestIndex := -1
	bestScore := -1
	for i, asset := range assets {
		score := assetScore(asset.Name, goarch)
		if score > bestScore {
			bestScore = score
			bestIndex = i
		}
	}
	if bestIndex >= 0 && bestScore > 0 {
		return assets[bestIndex].Name, assets[bestIndex].BrowserDownloadURL
	}
	return "", ""
}

func assetScore(name, goarch string) int {
	name = strings.ToLower(strings.TrimSpace(name))
	arch := "x64"
	fallback := "amd64"
	if goarch == "386" {
		arch = "x86"
		fallback = "386"
	}
	if !strings.Contains(name, arch) && !strings.Contains(name, fallback) {
		return 0
	}
	if strings.Contains(name, "control") {
		return 0
	}
	score := 10
	if strings.Contains(name, "mofumouse-"+arch) || strings.Contains(name, "mofumouse-"+fallback) {
		score += 8
	}
	if strings.HasSuffix(name, ".exe") {
		score += 4
	}
	if strings.HasSuffix(name, ".zip") {
		score += 3
	}
	return score
}

func DefaultDownloadDir(version string) (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	version = safePathSegment(version)
	if version == "" {
		version = "latest"
	}
	return filepath.Join(dir, "MofuMouse", "updates", version), nil
}

func DefaultHelperDir() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "MofuMouse", "update-helper"), nil
}

func DownloadAsset(ctx context.Context, assetURL, assetName, version string) (DownloadResult, error) {
	dir, err := DefaultDownloadDir(version)
	if err != nil {
		return DownloadResult{}, err
	}
	client := &http.Client{Timeout: 5 * time.Minute}
	return DownloadAssetToDir(ctx, client, assetURL, assetName, dir)
}

func DownloadAssetToDir(ctx context.Context, client *http.Client, assetURL, assetName, dir string) (DownloadResult, error) {
	assetURL = strings.TrimSpace(assetURL)
	if assetURL == "" {
		return DownloadResult{}, fmt.Errorf("asset download URL is empty")
	}
	if client == nil {
		client = http.DefaultClient
	}
	name, err := safeAssetName(assetName)
	if err != nil {
		return DownloadResult{}, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return DownloadResult{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, assetURL, nil)
	if err != nil {
		return DownloadResult{}, err
	}
	req.Header.Set("User-Agent", "MofuMouse")
	resp, err := client.Do(req)
	if err != nil {
		return DownloadResult{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return DownloadResult{}, fmt.Errorf("asset download failed: %s", resp.Status)
	}

	finalPath := filepath.Join(dir, name)
	tempPath := finalPath + ".download"
	file, err := os.Create(tempPath)
	if err != nil {
		return DownloadResult{}, err
	}
	bytes, copyErr := io.Copy(file, resp.Body)
	closeErr := file.Close()
	if copyErr != nil {
		_ = os.Remove(tempPath)
		return DownloadResult{}, copyErr
	}
	if closeErr != nil {
		_ = os.Remove(tempPath)
		return DownloadResult{}, closeErr
	}
	if bytes == 0 {
		_ = os.Remove(tempPath)
		return DownloadResult{}, fmt.Errorf("downloaded asset is empty")
	}
	_ = os.Remove(finalPath)
	if err := os.Rename(tempPath, finalPath); err != nil {
		_ = os.Remove(tempPath)
		return DownloadResult{}, err
	}

	return DownloadResult{AssetName: name, Path: finalPath, Bytes: bytes}, nil
}

func safeAssetName(name string) (string, error) {
	name = strings.TrimSpace(name)
	separator := string(filepath.Separator)
	name = filepath.Base(strings.NewReplacer("/", separator, "\\", separator).Replace(name))
	if name == "" || name == "." || name == string(filepath.Separator) {
		return "", fmt.Errorf("asset name is empty")
	}
	var b strings.Builder
	for _, r := range name {
		switch {
		case r < 32:
			b.WriteRune('_')
		case strings.ContainsRune(`<>:"/\|?*`, r):
			b.WriteRune('_')
		default:
			b.WriteRune(r)
		}
	}
	name = strings.TrimSpace(b.String())
	ext := strings.ToLower(filepath.Ext(name))
	if ext != ".exe" && ext != ".zip" {
		return "", fmt.Errorf("unsupported update asset extension: %s", ext)
	}
	return name, nil
}

func safePathSegment(value string) string {
	value = strings.TrimSpace(value)
	var b strings.Builder
	for _, r := range value {
		switch {
		case r < 32:
			b.WriteRune('_')
		case strings.ContainsRune(`<>:"/\|?*`, r):
			b.WriteRune('_')
		default:
			b.WriteRune(r)
		}
	}
	return strings.Trim(b.String(), " .")
}

func CompareVersions(a, b string) int {
	ap := versionParts(a)
	bp := versionParts(b)
	max := len(ap)
	if len(bp) > max {
		max = len(bp)
	}
	for i := 0; i < max; i++ {
		av, bv := 0, 0
		if i < len(ap) {
			av = ap[i]
		}
		if i < len(bp) {
			bv = bp[i]
		}
		if av > bv {
			return 1
		}
		if av < bv {
			return -1
		}
	}
	return 0
}

func versionParts(v string) []int {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	var out []int
	for _, part := range strings.FieldsFunc(v, func(r rune) bool { return r == '.' || r == '-' || r == '_' }) {
		n := 0
		for _, r := range part {
			if r < '0' || r > '9' {
				break
			}
			n = n*10 + int(r-'0')
		}
		out = append(out, n)
	}
	return out
}
