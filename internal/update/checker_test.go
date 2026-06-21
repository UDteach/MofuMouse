package update

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"v1.2.0", "v1.1.9", 1},
		{"v1.0.0", "1.0.0", 0},
		{"v1.0.0", "v1.0.1", -1},
		{"v2.0", "v10.0", -1},
	}
	for _, tt := range tests {
		if got := CompareVersions(tt.a, tt.b); got != tt.want {
			t.Fatalf("CompareVersions(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestMatchingAssetForGOARCHPrefersMainExecutable(t *testing.T) {
	assets := []Asset{
		{Name: "mofumouse-control-x64.exe", BrowserDownloadURL: "control"},
		{Name: "mofumouse-x64.exe", BrowserDownloadURL: "main-x64"},
		{Name: "mofumouse-x86.exe", BrowserDownloadURL: "main-x86"},
	}

	name, url := MatchingAssetForGOARCH(assets, "amd64")
	if name != "mofumouse-x64.exe" || url != "main-x64" {
		t.Fatalf("amd64 asset = %q %q, want mofumouse-x64.exe main-x64", name, url)
	}

	name, url = MatchingAssetForGOARCH(assets, "386")
	if name != "mofumouse-x86.exe" || url != "main-x86" {
		t.Fatalf("386 asset = %q %q, want mofumouse-x86.exe main-x86", name, url)
	}
}

func TestMatchingAssetForGOARCHDoesNotSelectControlCenterOnly(t *testing.T) {
	assets := []Asset{
		{Name: "mofumouse-control-x64.exe", BrowserDownloadURL: "control"},
	}
	name, url := MatchingAssetForGOARCH(assets, "amd64")
	if name != "" || url != "" {
		t.Fatalf("control-only asset = %q %q, want empty", name, url)
	}
}

func TestSafeAssetNameRejectsUnexpectedExtensions(t *testing.T) {
	if _, err := safeAssetName("../mofumouse-x64.ps1"); err == nil {
		t.Fatal("safeAssetName should reject script assets")
	}
	if got, err := safeAssetName(`..\mofumouse-x64.exe`); err != nil || got != "mofumouse-x64.exe" {
		t.Fatalf("safeAssetName exe = %q, %v", got, err)
	}
}

func TestDownloadAssetToDir(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != "MofuMouse" {
			t.Fatalf("User-Agent = %q, want MofuMouse", r.Header.Get("User-Agent"))
		}
		_, _ = w.Write([]byte("fake exe bytes"))
	}))
	defer server.Close()

	dir := t.TempDir()
	result, err := DownloadAssetToDir(context.Background(), server.Client(), server.URL, "mofumouse-x64.exe", dir)
	if err != nil {
		t.Fatalf("DownloadAssetToDir: %v", err)
	}
	if result.AssetName != "mofumouse-x64.exe" {
		t.Fatalf("asset name = %q", result.AssetName)
	}
	if result.Bytes != int64(len("fake exe bytes")) {
		t.Fatalf("bytes = %d", result.Bytes)
	}
	data, err := os.ReadFile(filepath.Join(dir, "mofumouse-x64.exe"))
	if err != nil {
		t.Fatalf("read downloaded file: %v", err)
	}
	if string(data) != "fake exe bytes" {
		t.Fatalf("downloaded data = %q", string(data))
	}
	if _, err := os.Stat(filepath.Join(dir, "mofumouse-x64.exe.download")); !os.IsNotExist(err) {
		t.Fatalf("temporary download file should not remain: %v", err)
	}
}
