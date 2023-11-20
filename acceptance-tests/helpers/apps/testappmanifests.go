package apps

import (
	"csbbrokerpakgcp/acceptance-tests/helpers/testpath"
)

type ManifestCode string

const (
	MySQLTLSTestAppManifest       ManifestCode = "jdbctestapp/manifest.yml"
	MySQLNoAutoTLSTestAppManifest ManifestCode = "jdbctestapp/manifest-no-autotls.yml"
	PgSQLTLSTestAppManifest       ManifestCode = "jdbctestapp/manifest-postgres.yml"
	PgSQLNoAutoTLSTestAppManifest ManifestCode = "jdbctestapp/manifest-postgres-no-autotls.yml"
)

func (a ManifestCode) Path() string {
	return testpath.BrokerpakFile("acceptance-tests", "apps", string(a))
}

func WithTestAppManifest(manifest ManifestCode) Option {
	return func(a *App) {
		a.manifest = manifest.Path()
	}
}
