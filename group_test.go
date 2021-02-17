package admin_test

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/qor/admin"
	. "github.com/qor/admin/tests/dummy"
	"github.com/qor/admin/tests/utils"
	"github.com/qor/qor"
	qorTestUtils "github.com/qor/qor/test/utils"
	"github.com/qor/roles"
)

func genResourcePermissions(resourceList [][]string) admin.ResourcePermissions {
	results := []admin.ResourcePermission{}

	for _, r := range resourceList {
		acs := []admin.ResourceActionPermission{}
		for i, resourceAction := range r {
			// the first element of the slice is ResourceName, we only need actions here.
			if i == 0 {
				continue
			}
			acs = append(acs, admin.ResourceActionPermission{Name: resourceAction, Allowed: true})
		}

		rp := admin.ResourcePermission{Name: r[0], Allowed: true, Actions: acs}
		results = append(results, rp)
	}

	return results
}

func TestGroupMenuPermission(t *testing.T) {
	group, ctx := groupTestEnvPrep(t, genResourcePermissions([][]string{{"Company"}, {"Credit Card"}}))

	companyMenu := Admin.GetMenu("Companies")
	if !companyMenu.HasPermission(roles.Read, ctx) {
		t.Error("user should have permission to access allowed Company resource")
	}

	// check no group permission menu
	group.ResourcePermissions = admin.ResourcePermissions{}
	utils.AssertNoErr(t, db.Save(&group).Error)
	if companyMenu.HasPermission(roles.Read, ctx) {
		t.Error("user should not have permission to access company when it is not allowed")
	}

	individualMenuWithPermission := Admin.AddMenu(&admin.Menu{Name: "ExternalURL", Permission: roles.Allow(roles.CRUD, Role_developer)})
	if individualMenuWithPermission.HasPermission(roles.Read, ctx) {
		t.Error("admin user should not have permission to access menu which is visible to Developer only")
	}
}

func TestGroupNestedMenuPermission(t *testing.T) {
	qorTestUtils.ResetDBTables(db, &admin.Group{}, &User{})
	user := User{Name: LoggedInUserName, Role: Role_system_administrator}
	utils.AssertNoErr(t, db.Save(&user).Error)

	Admin.AddMenu(&admin.Menu{Name: "MenuA", Ancestors: []string{"MenuA Father"}})

	group := admin.Group{Name: "test group", Users: fmt.Sprintf("%d", user.ID), ResourcePermissions: genResourcePermissions([][]string{{"Company"}, {"MenuA"}})}
	utils.AssertNoErr(t, db.Save(&group).Error)

	ctx := &admin.Context{Context: &qor.Context{CurrentUser: user, DB: Admin.DB}, Admin: Admin, Settings: map[string]interface{}{}}

	nestedMenu := Admin.GetMenu("MenuA Father")
	if !nestedMenu.HasPermission(roles.Read, ctx) {
		t.Error("menu with sub menus should have permission when at least one of the sub menu is allowed to access")
	}
}

func TestNestedMenuRolePermission(t *testing.T) {
	qorTestUtils.ResetDBTables(db, &admin.Group{}, &User{})
	user := User{Name: LoggedInUserName, Role: Role_system_administrator}
	utils.AssertNoErr(t, db.Save(&user).Error)

	Admin.AddMenu(&admin.Menu{Name: "MenuA", Ancestors: []string{"MenuA Father"}, Permission: roles.Allow(roles.CRUD, Role_system_administrator)})

	ctx := &admin.Context{Context: &qor.Context{CurrentUser: user, DB: Admin.DB}, Admin: Admin, Settings: map[string]interface{}{}}
	ctx.Roles = []string{Role_system_administrator}

	nestedMenu := Admin.GetMenu("MenuA Father")
	if !nestedMenu.HasPermission(roles.Read, ctx) {
		t.Error("menu with sub menus should have permission when at least one of the sub menu is allowed to access either by group or role permission")
	}
}

func TestIndividualNoPermissionMenu(t *testing.T) {
	_, ctx := groupTestEnvPrep(t, genResourcePermissions([][]string{{"Company"}, {"Credit Card"}}))

	// Check no permission menu
	noPermissionMenu := Admin.AddMenu(&admin.Menu{Name: "Dashboard", Link: "/admin"})
	if noPermissionMenu.HasPermission(roles.Read, ctx) {
		t.Error("individual menu with no permission set should not be accessible without group permission")
	}

	Admin.SetGroupEnabled(false)
	if !noPermissionMenu.HasPermission(roles.Read, ctx) {
		t.Error("individual menu with no permission set should be accessible when group permission is not enabled")
	}
	Admin.SetGroupEnabled(true)
}

func TestGroupMenuPermissionShouldHasLowerPriorityThanRole(t *testing.T) {
	qorTestUtils.ResetDBTables(db, &admin.Group{}, &User{})
	user := User{Name: LoggedInUserName, Role: Role_system_administrator}
	utils.AssertNoErr(t, db.Save(&user).Error)

	// setup Admin, group enabled but this user has no group registered
	ctx := &admin.Context{Context: &qor.Context{CurrentUser: user, DB: Admin.DB},
		Admin: Admin, Settings: map[string]interface{}{}}
	ctx.Context.Roles = []string{Role_system_administrator}

	Admin.AddResource(&Profile{}, &admin.Config{Permission: roles.Allow(roles.CRUD, Role_system_administrator)})
	profileMenu := Admin.GetMenu("Profiles")
	if !profileMenu.HasPermission(roles.Read, ctx) {
		t.Error("user should have permission to access roles allowed resource")
	}
}

func TestRouterGroupPermission(t *testing.T) {
	// Allow to access company test
	RouterGroupPermissionTest(t, [][]string{{"Company"}, {"Credit Card"}}, 200)
	// Not allowed to access company test
	RouterGroupPermissionTest(t, [][]string{{"Credit Card"}}, 404)
}

func RouterGroupPermissionTest(t *testing.T, resourcePermissions [][]string, responseCode int) {
	qorTestUtils.ResetDBTables(db, &admin.Group{}, &User{}, &Company{})
	user := User{Name: LoggedInUserName, Role: Role_system_administrator}
	utils.AssertNoErr(t, db.Save(&user).Error)

	group := admin.Group{Name: "test group", Users: fmt.Sprintf("%d", user.ID), ResourcePermissions: genResourcePermissions(resourcePermissions)}
	utils.AssertNoErr(t, db.Save(&group).Error)

	newCompanyName := "a test company"

	updatedCompanyName := "a new company"
	company := Company{Name: "old company"}
	utils.AssertNoErr(t, db.Save(&company).Error)

	toBeDeletedCompanyName := "a legacy company"
	toBeDeletedCompany := Company{Name: toBeDeletedCompanyName}
	utils.AssertNoErr(t, db.Save(&toBeDeletedCompany).Error)

	cases := []struct {
		desc         string
		url          string
		responseCode int
		formValues   url.Values
	}{
		{"read", server.URL + "/admin/companies", responseCode, nil},
		{"create", server.URL + "/admin/companies", responseCode, url.Values{"QorResource.Name": {newCompanyName}}},
		{"update", server.URL + fmt.Sprintf("/admin/companies/%d", company.ID), responseCode, url.Values{"QorResource.Name": {updatedCompanyName}}},
		{"delete", server.URL + fmt.Sprintf("/admin/companies/%d", toBeDeletedCompany.ID), responseCode, url.Values{"_method": {"delete"}}},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			var (
				resp *http.Response
				err  error
			)

			switch c.desc {
			case "read":
				resp, err = http.Get(c.url)
			case "create", "update", "delete":
				resp, err = http.PostForm(c.url, c.formValues)
			}

			utils.AssertNoErr(t, err)

			if got, want := resp.StatusCode, c.responseCode; want != got {
				t.Errorf("expect user with group permission to have %v when %s companies but got %v", want, c.desc, got)
			}
		})
	}
}

func TestRegisterUserToGroups(t *testing.T) {
	qorTestUtils.ResetDBTables(db, &admin.Group{}, &User{})
	user := User{Name: LoggedInUserName, Role: Role_system_administrator}
	utils.AssertNoErr(t, db.Save(&user).Error)

	groupA := createTestGroup("A")
	groupB := createTestGroup("B")
	groupC := createTestGroup("C")

	err := admin.RegisterUserToGroups(db, []uint{groupA.ID, groupB.ID}, &user.ID)
	if err != nil {
		t.Fatal(err)
	}

	utils.AssertNoErr(t, db.First(groupA, groupA.ID).Error)
	utils.AssertNoErr(t, db.First(groupB, groupB.ID).Error)
	utils.AssertNoErr(t, db.First(groupC, groupC.ID).Error)

	if groupA.Users != fmt.Sprintf(",%d", user.ID) {
		t.Error("user didn't registered in group A", groupA.Users, user.ID)
	}
	if groupB.Users != fmt.Sprintf(",%d", user.ID) {
		t.Error("user didn't registered in group B", groupB.Users, user.ID)
	}
	if groupC.Users != "" {
		t.Error("user incorrectly registered in group C")
	}
}

func TestRegisterUserToGroupsEdgeCases(t *testing.T) {
	qorTestUtils.ResetDBTables(db, &admin.Group{}, &User{})
	user := User{Name: LoggedInUserName, Role: Role_system_administrator}
	utils.AssertNoErr(t, db.Save(&user).Error)

	groupA := createTestGroup("A")

	// Empty group ids
	err := admin.RegisterUserToGroups(db, []uint{}, &user.ID)
	if err == nil {
		t.Error("empty group ids doesn't return error")
	}

	// no user id
	err1 := admin.RegisterUserToGroups(db, []uint{groupA.ID}, nil)
	if err1 == nil {
		t.Error("blank user id doesn't return error")
	}

	// non-exist groups
	err2 := admin.RegisterUserToGroups(db, []uint{1000, 1010}, &user.ID)
	if err2 == nil {
		t.Error("non-exists group ids doesn't return error")
	}
}

type Campaign struct {
	Name string
}

func TestDefaultBehaviorWithNoGroupPermissionEnabled(t *testing.T) {
	qorTestUtils.ResetDBTables(db, &admin.Group{}, &User{}, &Campaign{})
	Admin.SetGroupEnabled(false)
	Admin.AddResource(&Campaign{})
	user := User{Name: LoggedInUserName, Role: Role_system_administrator}
	utils.AssertNoErr(t, db.Save(&user).Error)

	resp, err := http.Get(server.URL + "/admin/campaigns")

	utils.AssertNoErr(t, err)

	if got, want := resp.StatusCode, 200; want != got {
		t.Errorf("expect visit blank permission resource without group permission enabled to get %v but got %v", want, got)
	}
	Admin.SetGroupEnabled(true)
}

func TestSkipGroupPermissionResourceRouter(t *testing.T) {
	qorTestUtils.ResetDBTables(db, &admin.Group{}, &User{}, &Campaign{})
	Admin.AddResource(&Campaign{}, &admin.Config{SkipGroupControl: true})
	user := User{Name: LoggedInUserName, Role: Role_system_administrator}
	utils.AssertNoErr(t, db.Save(&user).Error)

	group := admin.Group{Name: "test group", Users: fmt.Sprintf("%d", user.ID)}
	utils.AssertNoErr(t, db.Save(&group).Error)

	resp, err := http.Get(server.URL + "/admin/campaigns")

	utils.AssertNoErr(t, err)

	if got, want := resp.StatusCode, 200; want != got {
		t.Errorf("expect visit skip group control resource to have %v but got %v", want, got)
	}
}

func TestActionIsAllowed(t *testing.T) {
	group, ctx := groupTestEnvPrep(t, genResourcePermissions([][]string{{"Company", "Publish"}, {"Credit Card"}}))
	actionPublish := Admin.GetResource("Company").GetAction("Publish")

	if !actionPublish.IsAllowed(roles.Read, ctx) {
		t.Error("action should have permission")
	}

	// check no group permission menu
	group.ResourcePermissions = genResourcePermissions([][]string{{"Company"}, {"Credit Card"}})
	utils.AssertNoErr(t, db.Save(&group).Error)
	if actionPublish.IsAllowed(roles.Read, ctx) {
		t.Error("user should not have permission to access publish action when it is not allowed")
	}
}

func TestActionIsAllowedWorkWithRolePermission(t *testing.T) {
	group, ctx := groupTestEnvPrep(t, genResourcePermissions([][]string{{"Company", "Preview"}, {"Credit Card"}}))
	actionPreview := Admin.GetResource("Company").GetAction("Preview")

	if actionPreview.IsAllowed(roles.Read, ctx) {
		t.Error("action should not have permission when group is allowed but role denied. role has higher power")
	}

	// group permission NOT allowed but role allowed
	group.ResourcePermissions = genResourcePermissions([][]string{{"Company"}, {"Credit Card"}})
	utils.AssertNoErr(t, db.Save(&group).Error)

	actionApprove := Admin.GetResource("Company").GetAction("Approve")
	if !actionApprove.IsAllowed(roles.Read, ctx) {
		t.Error("user should have permission on action when group is not allowed but role is allowed")
	}
}

func TestSkipGroupControlAction(t *testing.T) {
	_, ctx := groupTestEnvPrep(t, genResourcePermissions([][]string{{"Company", "Publish"}, {"Credit Card"}}))
	res := Admin.GetResource("Credit Card")
	actionPublish := res.Action(&admin.Action{
		Name: "Publish",
		Handler: func(argument *admin.ActionArgument) (err error) {
			fmt.Println("Publish company")
			return
		},
		Method:           "GET",
		SkipGroupControl: true,
		Modes:            []string{"edit"},
	})

	if !actionPublish.IsAllowed(roles.Read, ctx) {
		t.Error("action should have permission when skipped group control")
	}
}

func TestAllowedActions(t *testing.T) {
	_, ctx := groupTestEnvPrep(t, genResourcePermissions([][]string{{"Company", "Preview", "Publish"}, {"Credit Card"}}))
	var fakeRecord interface{}
	actions := Admin.GetResource("Company").GetActions()

	allowedActions := ctx.AllowedActions(actions, "edit", fakeRecord)
	results := []string{}
	for _, a := range allowedActions {
		results = append(results, a.Name)
	}

	if len(results) != 2 || !admin.Contains(results, "Publish") || !admin.Contains(results, "Approve") {
		t.Error("allowed action is not correct")
	}
}

func TestActionHasPermission(t *testing.T) {
	_, ctx := groupTestEnvPrep(t, genResourcePermissions([][]string{{"Company", "Preview", "Publish"}, {"Credit Card"}}))

	actionPreview := Admin.GetResource("Company").GetAction("Preview")
	actionPublish := Admin.GetResource("Company").GetAction("Publish")

	if actionPreview.HasPermission(roles.Read, ctx.Context) {
		t.Error("should not has permission when role is denied")
	}

	if !actionPublish.HasPermission(roles.Read, ctx.Context) {
		t.Error("should has permission when permission is not set but group is allowed")
	}

	Admin.SetGroupEnabled(false)
	res := Admin.GetResource("Credit Card")
	actionApprove := res.Action(&admin.Action{
		Name: "Approve",
		Handler: func(argument *admin.ActionArgument) (err error) {
			fmt.Println("Approve company")
			return
		},
		Method: "GET",
		Modes:  []string{"edit"},
	})

	if !actionApprove.IsAllowed(roles.Read, ctx) {
		t.Error("action should have permission when group is disabled")
	}
	Admin.SetGroupEnabled(true)
}

func TestActionRoutePermission(t *testing.T) {
	groupTestEnvPrep(t, genResourcePermissions([][]string{{"Company", "Preview", "Publish"}, {"Credit Card"}}))
	company := Company{Name: "old company"}
	utils.AssertNoErr(t, db.Save(&company).Error)

	var (
		resp *http.Response
		err  error
	)

	resp, err = http.Get(server.URL + fmt.Sprintf("/admin/companies/%d/publish", company.ID))
	if err != nil {
		t.Error("request failed")
	}
	if resp.StatusCode != 200 {
		t.Error("user cannot operate allowed action")
	}

	resp, err = http.Get(server.URL + fmt.Sprintf("/admin/companies/%d/preview", company.ID))
	if err != nil {
		t.Error("request failed")
	}
	if resp.StatusCode != 404 {
		t.Error("user should not be able to operate forbidden action")
	}
}

func groupTestEnvPrep(t *testing.T, resourcePermission admin.ResourcePermissions) (admin.Group, *admin.Context) {
	qorTestUtils.ResetDBTables(db, &admin.Group{}, &User{})
	user := User{Name: LoggedInUserName, Role: Role_system_administrator}
	utils.AssertNoErr(t, db.Save(&user).Error)

	group := admin.Group{Name: "test group", Users: fmt.Sprintf("%d", user.ID), ResourcePermissions: resourcePermission}
	utils.AssertNoErr(t, db.Save(&group).Error)

	// setup Admin
	ctx := &admin.Context{Context: &qor.Context{CurrentUser: user, DB: Admin.DB}, Admin: Admin, Settings: map[string]interface{}{}}
	ctx.Roles = []string{Role_system_administrator}

	return group, ctx
}

func createTestGroup(name string) *admin.Group {
	group := admin.Group{Name: name}
	if err := db.Save(&group).Error; err != nil {
		panic(err)
	}

	return &group
}
