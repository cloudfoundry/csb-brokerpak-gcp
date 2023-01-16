package apps

import (
	"fmt"
	"os"
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
	JavaDBApp        AppCode = "javadbapp/javadbapp-1.0.0.jar"
)

func (a AppCode) Dir() string {
	for _, d := range []string{"apps", "../apps"} {
		p := fmt.Sprintf("%s/%s", d, string(a))
		_, err := os.Stat(p)
		if err == nil {
			return p
		}
	}

	panic(fmt.Sprintf("could not find source for app: %s", a))
}

func WithApp(app AppCode) Option {
	switch app {
	case StackdriverTrace, JavaDBApp:
		return WithDir(app.Dir())
	default:
		return WithPreBuild(app.Dir())
	}
}
