package apps

import (
	"fmt"
	"os"
	"path/filepath"
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
	for _, d := range []string{"apps", "../apps"} {
		p := filepath.Join(d, string(a))
		_, err := os.Stat(p)
		if err == nil {
			return p
		}
	}

	panic(fmt.Sprintf("could not find source for app: %s", a))
}

func WithApp(app AppCode) Option {
	switch app {
	case StackdriverTrace, JDBCTestApp:
		return WithDir(app.Dir())
	default:
		return WithPreBuild(app.Dir())
	}
}
