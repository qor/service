package dummy

import (
	"fmt"

	"github.com/qor/admin"
	"github.com/qor/media"
	"github.com/qor/qor"
	"github.com/qor/qor/test/utils"
	"github.com/qor/roles"
)

// NewDummyAdmin generate admin for dummy app
func NewDummyAdmin(keepData ...bool) *admin.Admin {
	var (
		db     = utils.TestDB()
		models = []interface{}{&User{}, &CreditCard{}, &Address{}, &Language{}, &Profile{}, &Phone{}, &Company{}}
		Admin  = admin.New(&admin.AdminConfig{Auth: DummyAuth{}, DB: db})
	)

	media.RegisterCallbacks(db)

	InitRoles()

	for _, value := range models {
		if len(keepData) == 0 {
			db.DropTableIfExists(value)
		}
		db.AutoMigrate(value)
	}

	c := Admin.AddResource(&Company{})
	c.Action(&admin.Action{
		Name: "Publish",
		Handler: func(argument *admin.ActionArgument) (err error) {
			fmt.Println("Publish company")
			return
		},
		Method:   "GET",
		Resource: c,
		Modes:    []string{"edit"},
	})
	c.Action(&admin.Action{
		Name:       "Preview",
		Permission: roles.Deny(roles.CRUD, Role_system_administrator),
		Handler: func(argument *admin.ActionArgument) (err error) {
			fmt.Println("Preview company")
			return
		},
		Method:   "GET",
		Resource: c,
		Modes:    []string{"edit"},
	})
	c.Action(&admin.Action{
		Name:       "Approve",
		Permission: roles.Allow(roles.Read, Role_system_administrator),
		Handler: func(argument *admin.ActionArgument) (err error) {
			fmt.Println("Approve company")
			return
		},
		Method: "GET",
		Modes:  []string{"edit"},
	})
	Admin.AddResource(&CreditCard{})

	Admin.AddResource(&Language{}, &admin.Config{Name: "语种 & 语言", Priority: -1})
	user := Admin.AddResource(&User{}, &admin.Config{Permission: roles.Allow(roles.CRUD, Role_system_administrator)})
	user.Meta(&admin.Meta{
		Name: "CreditCard",
		Type: "single_edit",
	})
	user.Meta(&admin.Meta{
		Name: "Languages",
		Type: "select_many",
		Collection: func(resource interface{}, context *qor.Context) (results [][]string) {
			if languages := []Language{}; !context.GetDB().Find(&languages).RecordNotFound() {
				for _, language := range languages {
					results = append(results, []string{fmt.Sprint(language.ID), language.Name})
				}
			}
			return
		},
	})

	admin.RegisterGroup(Admin, user, User{}, &admin.Config{Name: "Groups"})

	return Admin
}

const LoggedInUserName = "QOR"

type DummyAuth struct {
}

func (DummyAuth) LoginURL(ctx *admin.Context) string {
	return "/auth/login"
}

func (DummyAuth) LogoutURL(ctx *admin.Context) string {
	return "/auth/logout"
}

func (DummyAuth) GetCurrentUser(ctx *admin.Context) qor.CurrentUser {
	u := User{}

	if err := ctx.Admin.DB.Where("name = ?", LoggedInUserName).First(&u).Error; err != nil {
		fmt.Println("Cannot load logged in user", err.Error())
		return nil
	}

	return u
}
