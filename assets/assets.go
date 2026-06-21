package assets

import "embed"

//go:embed sprites manifest
var FS embed.FS

func ReadSprite(name string) ([]byte, error) {
	return FS.ReadFile("sprites/" + name)
}
