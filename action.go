package admin

import (
	"path"
	"reflect"
	"strings"

	"github.com/qor/qor"
	"github.com/qor/qor/utils"
	"github.com/qor/roles"
)

// ActionArgument action argument that used in handle
type ActionArgument struct {
	PrimaryValues       []string
	Context             *Context
	Argument            interface{}
	SkipDefaultResponse bool
}

// Action action definiation
type Action struct {
	Name        string
	Label       string
	Method      string
	URL         func(record interface{}, context *Context) string
	URLOpenType string
	Visible     func(record interface{}, context *Context) bool
	Handler     func(argument *ActionArgument) error
	Modes       []string
	Resource    *Resource
	Permission  *roles.Permission
}

// Action register action for qor resource
func (res *Resource) Action(action *Action) *Action {
	for _, a := range res.actions {
		if a.Name == action.Name {
			if action.Label != "" {
				a.Label = action.Label
			}

			if action.Method != "" {
				a.Method = action.Method
			}

			if action.URL != nil {
				a.URL = action.URL
			}

			if action.URLOpenType != "" {
				a.URLOpenType = action.URLOpenType
			}

			if action.Visible != nil {
				a.Visible = action.Visible
			}

			if action.Handler != nil {
				a.Handler = action.Handler
			}

			if len(action.Modes) != 0 {
				a.Modes = action.Modes
			}

			if action.Resource != nil {
				a.Resource = action.Resource
			}

			if action.Permission != nil {
				a.Permission = action.Permission
			}

			*action = *a
			return a
		}
	}

	if action.Label == "" {
		action.Label = utils.HumanizeString(action.Name)
	}

	if action.Method == "" {
		if action.URL != nil {
			action.Method = "GET"
		} else {
			action.Method = "PUT"
		}
	}

	if action.URLOpenType == "" {
		if action.Resource != nil {
			action.URLOpenType = "bottomsheet"
		} else if action.Method == "GET" {
			action.URLOpenType = "_blank"
		}
	}

	res.actions = append(res.actions, action)

	// Register Actions into Router
	{
		actionController := &Controller{Admin: res.GetAdmin(), action: action}
		primaryKeyParams := res.ParamIDName()

		if action.Resource != nil {
			// Bulk Action
			res.RegisterRoute("GET", path.Join("!action", action.ToParam()), actionController.Action, &RouteConfig{Permissioner: action, PermissionMode: roles.Update})
			// Single Resource Action
			res.RegisterRoute("GET", path.Join(primaryKeyParams, action.ToParam()), actionController.Action, &RouteConfig{Permissioner: action, PermissionMode: roles.Update})
		}

		if action.Handler != nil {
			// Bulk Action
			res.RegisterRoute("PUT", path.Join("!action", action.ToParam()), actionController.Action, &RouteConfig{Permissioner: action, PermissionMode: roles.Update})
			// Single Resource action
			res.RegisterRoute("PUT", path.Join(primaryKeyParams, action.ToParam()), actionController.Action, &RouteConfig{Permissioner: action, PermissionMode: roles.Update})
		}
	}

	return action
}

// GetActions get registered filters
func (res *Resource) GetActions() []*Action {
	return res.actions
}

// GetAction get defined action
func (res *Resource) GetAction(name string) *Action {
	for _, action := range res.actions {
		if action.Name == name {
			return action
		}
	}
	return nil
}

// ToParam used to register routes for actions
func (action Action) ToParam() string {
	return utils.ToParamString(action.Name)
}

// HasPermission check if current user has permission for the action
func (action Action) HasPermission(mode roles.PermissionMode, context *qor.Context) bool {
	if action.Permission != nil {
		var roles = []interface{}{}
		for _, role := range context.Roles {
			roles = append(roles, role)
		}
		return action.Permission.HasPermission(mode, roles...)
	}

	return true
}

// FindSelectedRecords find selected records when run bulk actions
func (actionArgument *ActionArgument) FindSelectedRecords() (records []interface{}) {

	results := actionArgument.Context.Resource.NewSlice()

	actionArgument.LoadSelectedRecords(results)

	resultValues := reflect.Indirect(reflect.ValueOf(results))
	for i := 0; i < resultValues.Len(); i++ {
		records = append(records, resultValues.Index(i).Interface())
	}
	
	return
}

// LoadSelectedRecords load selected records when run bulk actions
func (actionArgument *ActionArgument) LoadSelectedRecords(out interface{}) {
	var (
		context   = actionArgument.Context
		resource  = context.Resource
		sqls      []string
		sqlParams []interface{}
	)

	if len(actionArgument.PrimaryValues) == 0 {
		return
	}

	clone := context.clone()
	for _, primaryValue := range actionArgument.PrimaryValues {
		primaryQuerySQL, primaryParams := resource.ToPrimaryQueryParams(primaryValue, context.Context)
		sqls = append(sqls, primaryQuerySQL)
		sqlParams = append(sqlParams, primaryParams...)
	}

	if len(sqls) > 0 {
		clone.SetDB(clone.GetDB().Where(strings.Join(sqls, " OR "), sqlParams...))
		clone.Searcher.Pagination.CurrentPage = -1
	}

	context = clone.Searcher.parseContext(true)

	if context.HasError() {
		return
	}

	clone.Searcher.Resource.CallFindMany(out, context)

	return
}

// IsAllowed check if current user has permission to view the action
func (action Action) isAllowed(mode roles.PermissionMode, context *Context, records ...interface{}) bool {
	if action.Visible != nil {
		if len(records) == 0 {
			if !action.Visible(nil, context) {
				return false
			}
		}
		for _, record := range records {
			if !action.Visible(record, context) {
				return false
			}
		}
	}

	if action.Permission != nil {
		return action.HasPermission(mode, context.Context)
	}

	if context.Resource != nil {
		return context.Resource.HasPermission(mode, context.Context)
	}
	return true
}
