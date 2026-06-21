package assets

import (
	"bytes"
	"encoding/json"
	"image/png"
	"os"
	"strconv"
	"testing"
)

func TestEmbeddedSprites(t *testing.T) {
	names := []string{
		"degu_idle_agouti.png",
		"degu_walk_agouti.png",
		"degu_walk2_agouti.png",
		"degu_walk2_gray.png",
		"degu_walk2_dark.png",
		"degu_walk2_cream.png",
		"degu_walk2_white.png",
		"degu_walk2_pied.png",
		"degu_run_agouti.png",
		"degu_run_gray.png",
		"degu_run_dark.png",
		"degu_run_cream.png",
		"degu_run_white.png",
		"degu_run_pied.png",
		"degu_return_agouti.png",
		"degu_guard_agouti.png",
		"degu_front_agouti.png",
		"degu_sleepy_agouti.png",
		"degu_sniff_agouti.png",
		"degu_sniff_gray.png",
		"degu_sniff_dark.png",
		"degu_sniff_cream.png",
		"degu_sniff_white.png",
		"degu_sniff_pied.png",
		"degu_groom_agouti.png",
		"degu_dig_agouti.png",
		"degu_roll_agouti.png",
		"degu_nibble_agouti.png",
		"degu_patpat_agouti.png",
		"degu_alert_agouti.png",
		"degu_idle_gray.png",
		"degu_idle_dark.png",
		"degu_idle_cream.png",
		"degu_idle_white.png",
		"degu_idle_pied.png",
		"mofumouse_cursor_companion_icon.png",
	}

	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			data, err := ReadSprite(name)
			if err != nil {
				t.Fatalf("read sprite: %v", err)
			}
			img, err := png.Decode(bytes.NewReader(data))
			if err != nil {
				t.Fatalf("decode sprite: %v", err)
			}
			if got := img.Bounds().Dx(); got != 32 {
				t.Fatalf("unexpected width: %d", got)
			}
			if got := img.Bounds().Dy(); got != 32 {
				t.Fatalf("unexpected height: %d", got)
			}
		})
	}
}

func TestManifestReferencesExist(t *testing.T) {
	data, err := FS.ReadFile("manifest/assets.json")
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	var manifest struct {
		RuntimeSizes []int `json:"runtime_sizes"`
		Sprites      []struct {
			ID     string            `json:"id"`
			Path   string            `json:"path"`
			Source string            `json:"source"`
			Status string            `json:"status"`
			Tiers  map[string]string `json:"tiers"`
		} `json:"sprites"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("decode manifest: %v", err)
	}
	if len(manifest.Sprites) == 0 {
		t.Fatal("manifest has no sprites")
	}
	runtimeCount := 0
	for _, sprite := range manifest.Sprites {
		if sprite.Status == "runtime" {
			runtimeCount++
		}
		data, err := FS.ReadFile(sprite.Path)
		if err != nil {
			t.Fatalf("%s embedded path %q: %v", sprite.ID, sprite.Path, err)
		}
		if _, err := png.Decode(bytes.NewReader(data)); err != nil {
			t.Fatalf("%s decode embedded path %q: %v", sprite.ID, sprite.Path, err)
		}
		if _, err := os.Stat(sprite.Source); err != nil {
			t.Fatalf("%s source path %q: %v", sprite.ID, sprite.Source, err)
		}
		if len(manifest.RuntimeSizes) > 0 {
			for _, size := range manifest.RuntimeSizes {
				path := sprite.Tiers[jsonSizeKey(size)]
				if path == "" {
					t.Fatalf("%s missing runtime tier %d", sprite.ID, size)
				}
				data, err := FS.ReadFile(path)
				if err != nil {
					t.Fatalf("%s embedded tier %q: %v", sprite.ID, path, err)
				}
				img, err := png.Decode(bytes.NewReader(data))
				if err != nil {
					t.Fatalf("%s decode embedded tier %q: %v", sprite.ID, path, err)
				}
				if got := img.Bounds().Dx(); got != size {
					t.Fatalf("%s tier %q width = %d, want %d", sprite.ID, path, got, size)
				}
				if got := img.Bounds().Dy(); got != size {
					t.Fatalf("%s tier %q height = %d, want %d", sprite.ID, path, got, size)
				}
			}
		}
	}
	if runtimeCount < 14 {
		t.Fatalf("runtime sprite count = %d, want at least 14", runtimeCount)
	}
}

func jsonSizeKey(size int) string {
	return strconv.Itoa(size)
}
