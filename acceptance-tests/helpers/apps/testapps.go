package apps

import "fmt"

type AppCode string

const (
	Storage    AppCode = "storageapp"
	MySQL      AppCode = "mysqlapp"
	PostgreSQL AppCode = "postgresqlapp"
	Redis	   AppCode = "redisapp"
)

func (a AppCode) Dir() string {
	return fmt.Sprintf("../apps/%s", string(a))
}

func WithApp(app AppCode) Option {
	return WithPreBuild(app.Dir())
}
