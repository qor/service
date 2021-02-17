package dummy

import (
	"net/http"

	"github.com/qor/roles"
)

const (
	Role_editor               = "Editor"
	Role_supervisor           = "Supervisor"
	Role_system_administrator = "admin"
	Role_developer            = "Developer"
)

// InitRoles initialize roles of the system.
func InitRoles() {
	roles.Register(Role_editor, func(req *http.Request, currentUser interface{}) bool {
		return currentUser != nil && currentUser.(User).Role == Role_editor
	})
	roles.Register(Role_supervisor, func(req *http.Request, currentUser interface{}) bool {
		return currentUser != nil && currentUser.(User).Role == Role_supervisor
	})
	roles.Register(Role_system_administrator, func(req *http.Request, currentUser interface{}) bool {
		return currentUser != nil && currentUser.(User).Role == Role_system_administrator
	})
	roles.Register(Role_developer, func(req *http.Request, currentUser interface{}) bool {
		return currentUser != nil && currentUser.(User).Role == Role_developer
	})
}
