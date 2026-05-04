package i18n

import (
	"embed"
	"encoding/json"
	"io/fs"

	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"sigs.k8s.io/yaml"
)

//go:embed locale.*
var mFile embed.FS

// Bundle is the i18n bundle
var Bundle *i18n.Bundle

func init() {
	Bundle = i18n.NewBundle(language.English)
	Bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
	Bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	Bundle.RegisterUnmarshalFunc("yaml", func(data []byte, v any) error {
		return yaml.Unmarshal(data, v)
	})
	if err := fs.WalkDir(mFile, ".", func(path string, d fs.DirEntry, err error) error {
		if path != "." {
			_, err := Bundle.LoadMessageFileFS(mFile, path)
			if err != nil {
				return err
			}
		}
		return err
	}); err != nil {
		panic(err)
	}
}
