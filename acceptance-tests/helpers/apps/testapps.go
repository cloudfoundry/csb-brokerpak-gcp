package apps

import (
	"csbbrokerpakgcp/acceptance-tests/helpers/testpath"
)

type AppCode string

const (
	Dataproc         AppCode = "dataprocapp"
	Spanner          AppCode = "spannerapp"
	Storage          AppCode = "storageapp"
	MySQL            AppCode = "mysqlapp"
	PostgreSQL       AppCode = "postgresqlapp"
	Redis            AppCode = "redisapp"
	StackdriverTrace AppCode = "stackdrivertraceapp"
	JDBCTestApp      AppCode = "jdbctestapp"
)

func (a AppCode) Dir() string {
	return testpath.BrokerpakFile("acceptance-tests", "apps", string(a))
}

func WithApp(app AppCode) Option {
	switch app {
	case StackdriverTrace, JDBCTestApp:
		return WithDir(app.Dir())
	default:
		return WithPreBuild(app.Dir())
	}
}
