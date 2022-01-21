package apps

import "fmt"

type AppCode string

const (
	Dataproc         AppCode = "dataprocapp"
	Spanner          AppCode = "spannerapp"
	Storage          AppCode = "storageapp"
	MySQL            AppCode = "mysqlapp"
	PostgreSQL       AppCode = "postgresqlapp"
	Redis            AppCode = "redisapp"
	StackdriverTrace AppCode = "stackdrivertraceapp"
)

func (a AppCode) Dir() string {
	return fmt.Sprintf("../apps/%s", string(a))
}

func WithApp(app AppCode) Option {
	switch app {
	case StackdriverTrace:
		return WithDir(app.Dir())
	default:
		return WithPreBuild(app.Dir())
	}
}
