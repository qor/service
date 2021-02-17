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
	Name             string
	Label            string
	Method           string
	URL              func(record interface{}, context *Context) string
	URLOpenType      string
	Visible          func(record interface{}, context *Context) bool
	Handler          func(argument *ActionArgument) error
	Modes            []string
	Resource         *Resource
	Permission       *roles.Permission
	SkipGroupControl bool

	// when registering an action to a resource, it belongs to this resource. this is used for group permission
	belongedResource *Resource
}

// SetBelongedResource sets belonged resource
func (a *Action) SetBelongedResource(res *Resource) {
	a.belongedResource = res
}

// GetBelongedResource gets belonged resource
func (a *Action) GetBelongedResource() *Resource {
	return a.belongedResource
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

			a.SetBelongedResource(res)
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

	action.SetBelongedResource(res)
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
func (action Action) HasPermission(mode roles.PermissionMode, context *qor.Context) (result bool) {
	// Call this for action route handler permission check.
	// This will be executed twice if this is called from IsAllowed, but always return same result
	result = action.HasGroupPermission(context)

	if action.Permission != nil {
		var roles = []interface{}{}
		for _, role := range context.Roles {
			roles = append(roles, role)
		}
		result = action.Permission.HasPermission(mode, roles...)
	}

	return
}

func (action Action) HasGroupPermission(context *qor.Context) bool {
	if action.GetBelongedResource().GetAdmin().IsGroupEnabled() {
		if action.SkipGroupControl {
			return true
		} else {
			return ActionAllowedByGroup(context, action.GetBelongedResource().Name, action.Name)
		}
	}

	return true
}

// FindSelectedRecords find selected records when run bulk actions
func (actionArgument *ActionArgument) FindSelectedRecords() []interface{} {
	var (
		context   = actionArgument.Context
		resource  = context.Resource
		records   = []interface{}{}
		sqls      []string
		sqlParams []interface{}
	)

	if len(actionArgument.PrimaryValues) == 0 {
		return records
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
	results, _ := clone.FindMany()

	resultValues := reflect.Indirect(reflect.ValueOf(results))
	for i := 0; i < resultValues.Len(); i++ {
		records = append(records, resultValues.Index(i).Interface())
	}
	return records
}

// IsAllowed check if current user has permission to view the action
func (action Action) IsAllowed(mode roles.PermissionMode, context *Context, records ...interface{}) (result bool) {
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

	result = action.HasGroupPermission(context.Context)

	if action.Permission != nil {
		result = action.HasPermission(mode, context.Context)
		return
	}

	// When group is enabled, do not fallback to check resource permission
	// since when user can access this action, the resource must be accessible.
	if context.Resource != nil && !context.Admin.IsGroupEnabled() {
		result = context.Resource.HasPermission(mode, context.Context)
		return
	}

	return
}
