package apps

import (
	"csbbrokerpakgcp/acceptance-tests/helpers/testpath"
)

type ManifestCode string

const (
	MySQLTLSTestAppManifest       ManifestCode = "jdbctestapp/manifest.yml"
	MySQLNoAutoTLSTestAppManifest ManifestCode = "jdbctestapp/manifest-no-autotls.yml"
	PostgresTestAppManifest       ManifestCode = "jdbctestapp/manifest-postgres.yml"
	StorageTestAppManifest        ManifestCode = "springstorageapp/manifest-google-storage.yml"
	PubSubTestAppManifest         ManifestCode = "springpubsubapp/manifest-google-pubsub.yml"
)

func (a ManifestCode) Path() string {
	return testpath.BrokerpakFile("acceptance-tests", "apps", string(a))
}

func WithTestAppManifest(manifest ManifestCode) Option {
	return func(a *App) {
		a.manifest = manifest.Path()
	}
}
