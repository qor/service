package admin

import (
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/utils"
	"github.com/qor/roles"
	"github.com/qor/validations"
)

type UserModel interface {
	GetUsersByIDs(*gorm.DB, []string) interface{}
}

// RegisterGroup enable group permission system to admin.
// IMPORTANT: call this function after all the resource registration.
// resources registered later than this, will not be managed by group permission system.
func RegisterGroup(adm *Admin, userSelectRes *Resource, userModel UserModel, resConfig *Config) *Resource {
	adm.DB.AutoMigrate(&Group{})
	adm.SetGroupEnabled(true)

	if resConfig.Name == "" {
		resConfig.Name = "Groups"
	}

	group := adm.AddResource(&Group{}, resConfig)
	resourceList := GenResourceList(adm)

	group.IndexAttrs("ID", "Name", "CreatedAt", "UpdatedAt")
	group.NewAttrs("Name",
		&Section{
			Title: "Resource Permission",
			Rows: [][]string{
				{"ResourcePermissions"},
			}},
		&Section{
			Title: "People in this group",
			Rows: [][]string{
				{"Users"},
			}})
	group.EditAttrs("Name",
		&Section{
			Title: "Resource Permission",
			Rows: [][]string{
				{"ResourcePermissions"},
			}},
		&Section{
			Title: "People in this group",
			Rows: [][]string{
				{"Users"},
			}})

	group.Meta(&Meta{
		Name: "Users", Label: "",
		Config: &SelectManyConfig{
			RemoteDataResource: userSelectRes,
		},
		Setter: func(record interface{}, metaValue *resource.MetaValue, context *qor.Context) {
			if g, ok := record.(*Group); ok {
				primaryKeys := utils.ToArray(metaValue.Value)
				g.Users = strings.Join(primaryKeys, ",")
			}
		},
		Valuer: func(record interface{}, context *qor.Context) interface{} {
			if g, ok := record.(*Group); ok {
				ids := strings.Split(g.Users, ",")

				return userModel.GetUsersByIDs(context.GetDB(), ids)
			}

			return nil
		},
	})

	group.Meta(&Meta{Name: "Name", Label: "Group Name"})
	group.Meta(&Meta{Name: "ResourcePermissions", Label: "Resource Permissions", Type: "group_permission",
		Valuer: func(record interface{}, context *qor.Context) interface{} {
			if g, ok := record.(*Group); ok {
				results := []ResourcePermission{}

				for _, r := range resourceList {
					acs := []ResourceActionPermission{}
					for i, resourceAction := range r {
						// the first element of the slice is ResourceName, we only need actions here.
						if i == 0 {
							continue
						}
						acs = append(acs, ResourceActionPermission{Name: resourceAction, Allowed: g.HasResourceActionPermission(r[0], resourceAction)})
					}

					rp := ResourcePermission{Name: r[0], Allowed: g.HasResourcePermission(r[0]), Actions: acs}
					results = append(results, rp)
				}

				return results
			}

			return nil
		},
	})

	group.AddValidator(&resource.Validator{
		Handler: func(value interface{}, metaValues *resource.MetaValues, ctx *qor.Context) error {
			if meta := metaValues.Get("Name"); meta != nil {
				if name := utils.ToString(meta.Value); strings.TrimSpace(name) == "" {
					return validations.NewError(value, "Group Name", "Group Name can't be blank")
				}
			}
			return nil
		},
	})

	return initGroupSelectorRes(adm)
}

func initGroupSelectorRes(adm *Admin) *Resource {
	res := adm.AddResource(&Group{}, &Config{Name: "GroupSelector"})
	res.SearchAttrs("ID", "Name")
	adm.GetMenu("GroupSelectors").Permission = roles.Deny(roles.CRUD, roles.Anyone)
	return res
}

// GenResourceList collects resources with actions and menus that registered in admin.
// [][]string{resourceName, action1Name, action2Name}
func GenResourceList(adm *Admin) [][]string {
	availableResourcesName := [][]string{}
	resourceNames := []string{}

	for _, r := range adm.GetResources() {
		if r.Config.SkipGroupControl || r.Config.Invisible {
			continue
		}

		actionNames := []string{}
		for _, acts := range r.GetActions() {
			if acts.SkipGroupControl {
				continue
			}
			actionNames = append(actionNames, acts.Name)
		}
		resourceDetail := []string{r.Name}
		resourceDetail = append(resourceDetail, actionNames...)
		availableResourcesName = append(availableResourcesName, resourceDetail)
		resourceNames = append(resourceNames, r.Name)
	}

	for _, m := range adm.GetMenus() {
		// when menu has sub menus, it is not to be counted as a resource, when checking permission, if one of its sub menu is granted, the parent menu has permission too.
		for _, offspringMenu := range GetSelfMenuTree(m) {
			if m.Invisible {
				continue
			}

			// Only show menus with no associated resource, If menu belongs to a resource, we check that resource permission instead of menu's.
			if offspringMenu.AssociatedResource == nil {
				availableResourcesName = append(availableResourcesName, []string{offspringMenu.Name})
			}
		}
	}

	return availableResourcesName
}
