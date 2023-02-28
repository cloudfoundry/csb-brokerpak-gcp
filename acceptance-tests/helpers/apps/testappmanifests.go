package apps

import (
	"fmt"
	"os"
	"path/filepath"
)

type ManifestCode string

const (
	MySQLTLSTestAppManifest       ManifestCode = "jdbctestapp/manifest.yml"
	MySQLNoAutoTLSTestAppManifest ManifestCode = "jdbctestapp/manifest-no-autotls.yml"
)

func (a ManifestCode) Path() string {
	for _, d := range []string{"apps", "../apps"} {
		p, err := filepath.Abs(filepath.Join(d, string(a)))
		if err != nil {
			panic(fmt.Sprintf("error resolving absolute path: %s", err))
		}

		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	panic(fmt.Sprintf("could not find source for app manifest: %s", a))
}

func WithTestAppManifest(manifest ManifestCode) Option {
	return func(a *App) {
		a.manifest = manifest.Path()
	}
}
