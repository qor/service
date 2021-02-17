package admin

import (
	"path"

	"github.com/qor/qor"
	"github.com/qor/roles"
)

// GetMenus get all sidebar menus for admin
func (admin Admin) GetMenus() []*Menu {
	return admin.menus
}

// GetSelfMenuTree returns all the offspring menus of given menu. if given menu has no submenu, return itself.
func GetSelfMenuTree(m *Menu) (result []*Menu) {
	if len(m.GetSubMenus()) == 0 {
		result = append(result, m)
		return
	}

	for _, subM := range m.GetSubMenus() {
		result = append(result, GetSelfMenuTree(subM)...)
	}

	return
}

// AddMenu add a menu to admin sidebar
func (admin *Admin) AddMenu(menu *Menu) *Menu {
	// TODO: Consider warn user for menus larger than 2 levels. since we only support 2 levels menu atm.
	// if len(menu.Ancestors) > 1 {
	// 	panic(fmt.Sprintf("QOR admin now only support 2 levels of menu, but got %q", menu.Ancestors))
	// }

	menu.router = admin.router

	names := append(menu.Ancestors, menu.Name)

	if old := admin.GetMenu(names...); old != nil {
		if len(names) > 1 || len(old.Ancestors) == 0 {
			old.Link = menu.Link
			old.RelativePath = menu.RelativePath
			old.Priority = menu.Priority
			old.Permissioner = menu.Permissioner
			old.Permission = menu.Permission
			old.RelativePath = menu.RelativePath
			*menu = *old
			return old
		}
	}

	admin.menus = appendMenu(admin.menus, menu.Ancestors, menu)

	return menu
}

// GetMenu get sidebar menu with name
func (admin Admin) GetMenu(name ...string) *Menu {
	return getMenu(admin.menus, name...)
}

////////////////////////////////////////////////////////////////////////////////
// Sidebar Menu
////////////////////////////////////////////////////////////////////////////////

// Menu admin sidebar menu definiation
type Menu struct {
	Name               string
	IconName           string
	Link               string
	RelativePath       string
	Priority           int
	Ancestors          []string
	Permissioner       HasPermissioner
	Permission         *roles.Permission
	Invisible          bool
	AssociatedResource *Resource

	subMenus []*Menu
	router   *Router
}

// URL return menu's URL
func (menu Menu) URL() string {
	if menu.Link != "" {
		return menu.Link
	}

	if (menu.router != nil) && (menu.RelativePath != "") {
		return path.Join(menu.router.Prefix, menu.RelativePath)
	}

	return menu.RelativePath
}

// HasPermission check menu has permission or not
func (menu Menu) HasPermission(mode roles.PermissionMode, context *Context) (result bool) {
	// When menu has no Permission and Permissioner set, this implicitly means it has no resource associated.
	// But it also can be controlled by group permission
	if menu.Permission == nil && menu.Permissioner == nil {
		result = true
	}

	checkMenuRolePermission := func(menu Menu, previousResult bool) bool {
		if menu.Permission != nil {
			var roles = []interface{}{}
			for _, role := range context.Roles {
				roles = append(roles, role)
			}
			return menu.Permission.HasPermission(mode, roles...)
		} else if menu.Permissioner != nil {
			// When group is enabled, resource with no Permission set will no longer return true. But return group permission result instead.
			if context.Context.Config == nil {
				context.Context.Config = &qor.Config{GroupPermissionEnabled: context.Admin.IsGroupEnabled(), GroupPermissionResult: previousResult}
			} else {
				context.Context.Config.GroupPermissionEnabled = context.Admin.IsGroupEnabled()
				context.Context.Config.GroupPermissionResult = previousResult
			}

			return menu.Permissioner.HasPermission(mode, context.Context)
		}

		return previousResult
	}

	// Check group permission first, since it has lower priority than roles.
	if context.Admin.IsGroupEnabled() {
		// If menu has sub menus, we check sub menus permission instead.
		// As long as one of the sub menus has permission, then the parent menus has permission too.
		for _, m := range GetSelfMenuTree(&menu) {
			menuName := m.Name
			// If menu belongs to a resource, we check that resource permission instead of menu's.
			if m.AssociatedResource != nil {
				menuName = m.AssociatedResource.Name
			}

			result = ResourceAllowedByGroup(context, menuName)
			if result {
				break
			}

			if checkMenuRolePermission(*m, result) {
				result = true
				break
			}
		}
	}

	result = checkMenuRolePermission(menu, result)

	return
}

// GetSubMenus get submenus for a menu
func (menu *Menu) GetSubMenus() []*Menu {
	return menu.subMenus
}

func getMenu(menus []*Menu, names ...string) *Menu {
	if len(names) > 0 {
		name := names[0]
		for _, menu := range menus {
			if len(names) > 1 {
				if menu.Name == name {
					return getMenu(menu.subMenus, names[1:]...)
				}
			} else {
				if menu.Name == name {
					return menu
				}
				if len(menu.subMenus) > 0 {
					if m := getMenu(menu.subMenus, name); m != nil {
						return m
					}
				}
			}
		}
	}
	return nil
}

// generateMenu generates menu from the ancestors, only keep the first and last one.
// E.g. ancestors is []string{"Management", "Product Management"}, menu is "Product".
// The result would be "Management > Product". QOR only support 2 levels of menu.
func generateMenu(menus []string, menu *Menu) *Menu {
	menuCount := len(menus)
	for index := range menus {
		menu = &Menu{Name: menus[menuCount-index-1], subMenus: []*Menu{menu}}
	}

	return menu
}

func appendMenu(menus []*Menu, ancestors []string, menu *Menu) []*Menu {
	if len(ancestors) > 0 {
		for _, m := range menus {
			if m.Name != ancestors[0] {
				continue
			}

			if len(ancestors) > 1 {
				m.subMenus = appendMenu(m.subMenus, ancestors[1:], menu)
			} else {
				m.subMenus = appendMenu(m.subMenus, []string{}, menu)
			}

			return menus
		}
	}

	var newMenu = generateMenu(ancestors, menu)
	var added bool
	if len(menus) == 0 {
		menus = append(menus, newMenu)
	} else if newMenu.Priority > 0 {
		for idx, menu := range menus {
			if menu.Priority > newMenu.Priority || menu.Priority <= 0 {
				menus = append(menus[0:idx], append([]*Menu{newMenu}, menus[idx:]...)...)
				added = true
				break
			}
		}
		if !added {
			menus = append(menus, menu)
		}
	} else {
		if newMenu.Priority < 0 {
			for idx := len(menus) - 1; idx >= 0; idx-- {
				menu := menus[idx]
				if menu.Priority < newMenu.Priority || menu.Priority == 0 {
					menus = append(menus[0:idx+1], append([]*Menu{newMenu}, menus[idx+1:]...)...)
					added = true
					break
				}
			}

			if !added {
				menus = append(menus, menu)
			}
		} else {
			for idx := len(menus) - 1; idx >= 0; idx-- {
				menu := menus[idx]
				if menu.Priority >= 0 {
					menus = append(menus[0:idx+1], append([]*Menu{newMenu}, menus[idx+1:]...)...)
					added = true
					break
				}
			}

			if !added {
				menus = append([]*Menu{menu}, menus...)
			}
		}
	}

	return menus
}
