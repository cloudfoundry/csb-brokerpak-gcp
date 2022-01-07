package services

import (
	"acceptancetests/helpers/apps"
	"acceptancetests/helpers/bindings"
)

func (s *ServiceInstance) BindWithParams(app *apps.App, params string) *bindings.Binding {
	return bindings.Bind(s.Name, app.Name, params)
}

func (s *ServiceInstance) Bind(app *apps.App) *bindings.Binding {
	return bindings.Bind(s.Name, app.Name, "")
}
