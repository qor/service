package admin

import (
	"html/template"
	"path/filepath"
	"reflect"

	"github.com/qor/assetfs"
	"github.com/qor/qor"
	"github.com/qor/qor/utils"
	"github.com/qor/roles"
)

var (
	globalViewPaths []string
	globalAssetFSes []assetfs.Interface
)

// HasPermissioner has permission interface
type HasPermissioner interface {
	HasPermission(roles.PermissionMode, *qor.Context) bool
}

// ResourceNamer is an interface for models that defined method `ResourceName`
type ResourceNamer interface {
	ResourceName() string
}

// I18n define admin's i18n interface
type I18n interface {
	Scope(scope string) I18n
	Default(value string) I18n
	T(locale string, key string, args ...interface{}) template.HTML
}

// RegisterViewPath register view path for all assetfs
func RegisterViewPath(pth string) {
	globalViewPaths = append(globalViewPaths, pth)

	for _, assetFS := range globalAssetFSes {
		if assetFS.RegisterPath(filepath.Join(utils.AppRoot, "vendor", pth)) != nil {
			for _, gopath := range utils.GOPATH() {
				if assetFS.RegisterPath(filepath.Join(gopath, "src", pth)) == nil {
					break
				}
			}
		}
	}
}

func equal(a, b interface{}) bool {
	return reflect.DeepEqual(a, b)
}

// Contains determine whether the second parameter(single) is contained by
// the first parameter(collection)
func Contains(a interface{}, b interface{}) bool {
	if _, ok := b.(int); ok {
		if tempB, ok := a.([]int); ok {
			for _, v := range tempB {
				if b == v {
					return true
				}
			}
		}
	}

	if _, ok := b.(uint); ok {
		if tempB, ok := a.([]uint); ok {
			for _, v := range tempB {
				if b == v {
					return true
				}
			}
		}
	}

	if _, ok := b.(string); ok {
		if tempB, ok := a.([]string); ok {
			for _, v := range tempB {
				if b == v {
					return true
				}
			}
		}

	}

	return false
}
