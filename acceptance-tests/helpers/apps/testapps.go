package apps

import (
	"csbbrokerpakgcp/acceptance-tests/helpers/testpath"
)

type AppCode string

const (
	Storage          AppCode = "storageapp"
	MySQL            AppCode = "mysqlapp"
	PostgreSQL       AppCode = "postgresqlapp"
	StackdriverTrace AppCode = "stackdrivertraceapp"
	JDBCTestApp      AppCode = "jdbctestapp"
	SpringStorageApp AppCode = "springstorageapp"
)

func (a AppCode) Dir() string {
	return testpath.BrokerpakFile("acceptance-tests", "apps", string(a))
}

func WithApp(app AppCode) Option {
	switch app {
	case StackdriverTrace, JDBCTestApp:
		return WithDir(app.Dir())
	case SpringStorageApp:
		return WithMavenPreBuild(app.Dir())
	default:
		return WithGoPreBuild(app.Dir())
	}
}
