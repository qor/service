## QOR Admin

Instantly create a beautiful, cross platform, configurable Admin Interface and API for managing your data in minutes.

[![GoDoc](https://godoc.org/github.com/qor/admin?status.svg)](https://godoc.org/github.com/qor/admin)
[![Build Status](https://travis-ci.com/qor/admin.svg?branch=master)](https://travis-ci.com/qor/admin)

**For security issues, please send us an email to security@getqor.com and give us time to respond BEFORE posting as an issue or reporting on public forums.**

## Features

- Generate Admin Interface for managing data
- RESTFul JSON API
- Association handling
- Search and filtering
- Actions/Batch Actions
- Authentication and Authorization
- Extendability

## Quick Start

```go
package main

import (
  "fmt"
  "net/http"
  "github.com/jinzhu/gorm"
  _ "github.com/mattn/go-sqlite3"
  "github.com/qor/admin"
)

// Create a GORM-backend model
type User struct {
  gorm.Model
  Name string
}

// Create another GORM-backend model
type Product struct {
  gorm.Model
  Name        string
  Description string
}

func main() {
  DB, _ := gorm.Open("sqlite3", "demo.db")
  DB.AutoMigrate(&User{}, &Product{})

  // Initialize
  Admin := admin.New(&admin.AdminConfig{DB: DB})

  // Allow to use Admin to manage User, Product
  Admin.AddResource(&User{})
  Admin.AddResource(&Product{})

  // initialize an HTTP request multiplexer
  mux := http.NewServeMux()

  // Mount admin interface to mux
  Admin.MountTo("/admin", mux)

  fmt.Println("Listening on: 9000")
  http.ListenAndServe(":9000", mux)
}
```

`go run main.go` and visit `localhost:9000/admin` to see the result!

## Live DEMO

* Live Demo [http://demo.getqor.com/admin](http://demo.getqor.com/admin)
* Source Code of Live Demo [https://github.com/qor/qor-example](https://github.com/qor/qor-example)

## Documentation

<https://doc.getqor.com/admin>

### Group permission system.

QOR Admin already has a "role" system to control permission. However, it can only be managed by the developer with hardcoded configuration. The group permission system aim to let admin users can manage permissions at runtime. Once group system is enabled, user with no group permission cannot see or operate the resource unless it has proper role to access it.

The principle is GroupPermission should not affect any previous behavior when it is NOT enabled. But if you want to use it, be aware that once GroupPermission is enabled, resources with blank `Permission` will NOT be accessible by default. You have to assign it to a group of users before the user can access it.

#### Usage

To use group permission, First, you should enable Authentication system of QOR Admin and the "user" model should implements these two interfaces.

```go
// GetUsersByIDs is used for group's user selector, it should return user list by given user ids.
func (u User) GetUsersByIDs(db *gorm.DB, ids []string) interface{}
// GetID returns user's id for permission check
func (u User) GetID() uint
```

Then enable group permission in admin.
IMPORTANT: resources registered later than this will not be managed by group permission system. So call this function after all the resources that you wanted are registered.

```go
// adm is a qor admin instance
// InitUserSelectorRes(adm) returns an *admin.Resource of user for selector, an example attached below
// User{} is the "user" struct
// Last one is config. We recommend to set permission to this group like this, so that the initial user with role "Developer" could access group at the beginning. the permission check logic between role and group permission will be explained later.
admin.RegisterGroup(adm, InitUserSelectorRes(adm), User{}, &admin.Config{Name: "Groups", Permission: roles.Allow(roles.CRUD, "Developer")})

func InitUserSelectorRes(adm *admin.Admin) *admin.Resource {
    // SkipGroupControl makes this resource invisible in resource list selector of group
    res := adm.AddResource(&User{}, &admin.Config{Name: "UserSelector", SkipGroupControl: true})
    res.SearchAttrs("ID", "Name")
    searchHander := res.SearchHandler
    res.SearchHandler = func(keyword string, ctx *qor.Context) *gorm.DB {
      ctx.SetDB(ctx.DB.Where("role <> ? AND role <> ?", Role_developer))
      return searchHander(keyword, ctx)
    }
    // This hide menu from the sidebar and group resource list selector
    adm.GetMenu("UserSelectors").Invisible = true

    return res
}
```

And that's it, login to QOR Admin, you should see Group is there.

#### How to integrate group into user form
If you want to set user's group in user edit form, The `admin.RegisterGroup` returns a group selector resource and a function `func RegisterUserToGroups(db *gorm.DB, groupIDs []uint, uid *uint) (err error)` can register user into groups.

#### The permission check logic about role and group permission.
In short, The role always has higher priority than group permission.
If ResourceA allow role "Developer" to access and UserA is a "Developer" but does not belong to any group. UserA still can access ResourceA.
But if UserA is a "Editor", UserA cannot access ResourceA.

So, A resource `roles.Allow(roles.CRUD, roles.Anyone)` skips the group permission system.

#### Set exceptional resource or actions.
Resources and actions both have a config called `SkipGroupControl`. If this is true, the current object will not appear in the resource list of group page and group permission check always return true.

## License

Released under the [MIT License](http://opensource.org/licenses/MIT).
