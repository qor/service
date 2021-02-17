package admin

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
)

type Group struct {
	gorm.Model

	Name  string
	Users string

	ResourcePermissions ResourcePermissions `sql:"type:text;"`
}

type ResourcePermissions []ResourcePermission

type ResourcePermission struct {
	Name    string
	Allowed bool
	Actions []ResourceActionPermission
}

type ResourceActionPermission struct {
	Name    string
	Allowed bool
}

func (g Group) HasResourcePermission(name string) bool {
	for _, res := range g.ResourcePermissions {
		if res.Name == name && res.Allowed {
			return true
		}
	}

	return false
}

func (g Group) IncludeUserID(uid string) bool {
	return Contains(strings.Split(g.Users, ","), uid)
}

func (g Group) HasResourceActionPermission(resName string, actionName string) bool {
	for _, res := range g.ResourcePermissions {
		if res.Name == resName && res.Allowed {
			for _, resAction := range res.Actions {
				if resAction.Name == actionName && resAction.Allowed {
					return true
				}
			}
		}
	}

	return false
}

// Scan scan value from database into struct
func (rp *ResourcePermissions) Scan(value interface{}) error {
	if bytes, ok := value.([]byte); ok {
		json.Unmarshal(bytes, rp)
	} else if str, ok := value.(string); ok {
		json.Unmarshal([]byte(str), rp)
	} else if strs, ok := value.([]string); ok {
		for _, str := range strs {
			json.Unmarshal([]byte(str), rp)
		}
	}
	return nil
}

// Value get value from struct, and save into database
func (rp ResourcePermissions) Value() (driver.Value, error) {
	result, err := json.Marshal(rp)
	return string(result), err
}

func (g Group) TableName() string {
	return "qor_groups"
}

const (
	permissionTypeResource = "resource"
	permissionTypeAction   = "action"
)

// ResourceAllowedByGroup checks if current user allowed to access given resource
func ResourceAllowedByGroup(context *Context, resName string) bool {
	return checkPermission(context.Context, resName, "", permissionTypeResource)
}

// ActionAllowedByGroup checks if current user allowed to access given action of given resource
func ActionAllowedByGroup(context *qor.Context, resName string, actionName string) bool {
	return checkPermission(context, resName, actionName, permissionTypeAction)
}

func checkPermission(context *qor.Context, resName string, actionName string, permissionType string) (result bool) {
	uid := context.CurrentUser.GetID()
	db := context.DB.New()

	idStr := fmt.Sprintf("%d", uid)
	groups := []Group{}
	if err := db.Find(&groups).Error; err != nil {
		log.Print(err.Error())
		return false
	}

	for _, g := range groups {
		if len(g.ResourcePermissions) != 0 && g.IncludeUserID(idStr) {
			switch permissionType {
			case permissionTypeResource:
				result = g.HasResourcePermission(resName)
			case permissionTypeAction:
				result = g.HasResourceActionPermission(resName, actionName)
			}

			if result {
				break
			}
		}
	}

	return
}

// RegisterUserToGroups register user into groups
func RegisterUserToGroups(db *gorm.DB, groupIDs []uint, uid *uint) (err error) {
	if len(groupIDs) == 0 {
		return errors.New("group ids must be provided")
	}

	if uid == nil {
		return errors.New("user id must be provided")
	}

	idStr := fmt.Sprintf("%d", *uid)
	groups := []Group{}
	if err = db.Where("id IN (?)", groupIDs).Find(&groups).Error; err != nil {
		return err
	}

	if len(groups) == 0 {
		return fmt.Errorf("no group can be found by given ids %v, please have a check", groupIDs)
	}

	tx := db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	for _, g := range groups {
		if !Contains(strings.Split(g.Users, ","), idStr) {
			userIDs := strings.Split(g.Users, ",")
			userIDs = append(userIDs, idStr)
			g.Users = strings.Join(userIDs, ",")
			err = tx.Save(&g).Error
		}
	}

	return nil
}
