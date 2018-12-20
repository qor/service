package admin

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/utils"
	"github.com/qor/roles"
)

// Meta meta struct definition
type Meta struct {
	Name            string
	Type            string
	Label           string
	FieldName       string
	Setter          func(resource interface{}, metaValue *resource.MetaValue, context *qor.Context)
	Valuer          func(interface{}, *qor.Context) interface{}
	FormattedValuer func(interface{}, *qor.Context) interface{}
	Resource        *Resource
	Permission      *roles.Permission
	Config          MetaConfigInterface

	Metas      []resource.Metaor
	Collection interface{}
	resource.Meta
	baseResource *Resource
}

// metaConfig meta config
type metaConfig struct {
}

// GetTemplate get customized template for meta
func (metaConfig) GetTemplate(context *Context, metaType string) ([]byte, error) {
	return nil, errors.New("not implemented")
}

// MetaConfigInterface meta config interface
type MetaConfigInterface interface {
	resource.MetaConfigInterface
}

// GetMetas get sub metas
func (meta *Meta) GetMetas() []resource.Metaor {
	if len(meta.Metas) > 0 {
		return meta.Metas
	} else if meta.Resource == nil {
		return []resource.Metaor{}
	} else {
		return meta.Resource.GetMetas([]string{})
	}
}

// GetResource get resource from meta
func (meta *Meta) GetResource() resource.Resourcer {
	return meta.Resource
}

// DBName get meta's db name
func (meta *Meta) DBName() string {
	if meta.FieldStruct != nil {
		return meta.FieldStruct.DBName
	}
	return ""
}

func getField(fields []*gorm.StructField, name string) (*gorm.StructField, bool) {
	for _, field := range fields {
		if field.Name == name || field.DBName == name {
			return field, true
		}
	}
	return nil, false
}

func (meta *Meta) setBaseResource(base *Resource) {
	res := meta.Resource
	res.base = base

	findOneHandle := res.FindOneHandler
	res.FindOneHandler = func(value interface{}, metaValues *resource.MetaValues, context *qor.Context) (err error) {
		if metaValues != nil {
			return findOneHandle(value, metaValues, context)
		}

		if primaryKey := res.GetPrimaryValue(context.Request); primaryKey != "" {
			clone := context.Clone()
			baseValue := base.NewStruct()
			if err = base.FindOneHandler(baseValue, nil, clone); err == nil {
				scope := clone.GetDB().NewScope(nil)
				sql := fmt.Sprintf("%v = ?", scope.Quote(res.PrimaryDBName()))
				err = context.GetDB().Model(baseValue).Where(sql, primaryKey).Related(value).Error
			}
		}
		return
	}

	res.FindManyHandler = func(value interface{}, context *qor.Context) error {
		var (
			err       error
			clone     = context.Clone()
			baseValue = base.NewStruct()
		)

		if err = base.FindOneHandler(baseValue, nil, clone); err == nil {
			base.FindOneHandler(baseValue, nil, clone)
			return context.GetDB().Model(baseValue).Related(value).Error
		}
		return err
	}

	res.SaveHandler = func(value interface{}, context *qor.Context) error {
		var (
			err       error
			clone     = context.Clone()
			baseValue = base.NewStruct()
		)

		if err = base.FindOneHandler(baseValue, nil, clone); err == nil {
			base.FindOneHandler(baseValue, nil, clone)
			return context.GetDB().Model(baseValue).Association(meta.FieldName).Append(value).Error
		}
		return err
	}

	res.DeleteHandler = func(value interface{}, context *qor.Context) (err error) {
		var clone = context.Clone()
		var baseValue = base.NewStruct()
		if primryKey := res.GetPrimaryValue(context.Request); primryKey != "" {
			var scope = clone.GetDB().NewScope(nil)
			var sql = fmt.Sprintf("%v = ?", scope.Quote(res.PrimaryDBName()))
			if err = context.GetDB().First(value, sql, primryKey).Error; err == nil {
				if err = base.FindOneHandler(baseValue, nil, clone); err == nil {
					base.FindOneHandler(baseValue, nil, clone)
					return context.GetDB().Model(baseValue).Association(meta.FieldName).Delete(value).Error
				}
			}
		}
		return
	}
}

// SetPermission set meta's permission
func (meta *Meta) SetPermission(permission *roles.Permission) {
	meta.Permission = permission
	meta.Meta.Permission = permission
	if meta.Resource != nil {
		meta.Resource.Permission = permission
	}
}

//  HasPermission check has permission or not
func (meta Meta) HasPermission(mode roles.PermissionMode, context *qor.Context) bool {
	if meta.Permission != nil && !meta.Permission.HasPermission(mode, context.Roles...) {
		return false
	}

	if meta.baseResource != nil {
		return meta.baseResource.HasPermission(mode, context)
	}

	return true
}

func (meta *Meta) updateMeta() {
	meta.Meta = resource.Meta{
		Name:            meta.Name,
		FieldName:       meta.FieldName,
		Setter:          meta.Setter,
		Valuer:          meta.Valuer,
		FormattedValuer: meta.FormattedValuer,
		Resource:        meta.baseResource,
		Permission:      meta.Permission,
		Config:          meta.Config,
	}

	meta.PreInitialize()
	if meta.FieldStruct != nil {
		if injector, ok := reflect.New(meta.FieldStruct.Struct.Type).Interface().(resource.ConfigureMetaBeforeInitializeInterface); ok {
			injector.ConfigureQorMetaBeforeInitialize(meta)
		}
	}

	meta.Initialize()

	if meta.Label == "" {
		meta.Label = utils.HumanizeString(meta.Name)
	}

	var fieldType reflect.Type
	var hasColumn = meta.FieldStruct != nil

	if hasColumn {
		fieldType = meta.FieldStruct.Struct.Type
		for fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}
	}

	// Set Meta Type
	if hasColumn {
		if meta.Type == "" {
			if _, ok := reflect.New(fieldType).Interface().(sql.Scanner); ok {
				if fieldType.Kind() == reflect.Struct {
					fieldType = reflect.Indirect(reflect.New(fieldType)).Field(0).Type()
				}
			}

			if relationship := meta.FieldStruct.Relationship; relationship != nil {
				if relationship.Kind == "has_one" {
					meta.Type = "single_edit"
				} else if relationship.Kind == "has_many" {
					meta.Type = "collection_edit"
				} else if relationship.Kind == "belongs_to" {
					meta.Type = "select_one"
				} else if relationship.Kind == "many_to_many" {
					meta.Type = "select_many"
				}
			} else {
				switch fieldType.Kind() {
				case reflect.String:
					var tag = meta.FieldStruct.Tag
					if size, ok := utils.ParseTagOption(tag.Get("sql"))["SIZE"]; ok {
						if i, _ := strconv.Atoi(size); i > 255 {
							meta.Type = "text"
						} else {
							meta.Type = "string"
						}
					} else if text, ok := utils.ParseTagOption(tag.Get("sql"))["TYPE"]; ok && (text == "text" || text == "longtext" || text == "longblob") {
						meta.Type = "text"
					} else {
						meta.Type = "string"
					}
				case reflect.Bool:
					meta.Type = "checkbox"
				default:
					if regexp.MustCompile(`^(.*)?(u)?(int)(\d+)?`).MatchString(fieldType.Kind().String()) {
						meta.Type = "number"
					} else if regexp.MustCompile(`^(.*)?(float)(\d+)?`).MatchString(fieldType.Kind().String()) {
						meta.Type = "float"
					} else if _, ok := reflect.New(fieldType).Interface().(*time.Time); ok {
						meta.Type = "datetime"
					} else {
						if fieldType.Kind() == reflect.Struct {
							meta.Type = "single_edit"
						} else if fieldType.Kind() == reflect.Slice {
							refelectType := fieldType.Elem()
							for refelectType.Kind() == reflect.Ptr {
								refelectType = refelectType.Elem()
							}
							if refelectType.Kind() == reflect.Struct {
								meta.Type = "collection_edit"
							}
						}
					}
				}
			}
		} else {
			if relationship := meta.FieldStruct.Relationship; relationship != nil {
				if (relationship.Kind == "has_one" || relationship.Kind == "has_many") && meta.Meta.Setter == nil && (meta.Type == "select_one" || meta.Type == "select_many") {
					meta.SetSetter(func(resource interface{}, metaValue *resource.MetaValue, context *qor.Context) {
						scope := &gorm.Scope{Value: resource}
						reflectValue := reflect.Indirect(reflect.ValueOf(resource))
						field := reflectValue.FieldByName(meta.FieldName)

						if field.Kind() == reflect.Ptr {
							if field.IsNil() {
								field.Set(utils.NewValue(field.Type()).Elem())
							}

							for field.Kind() == reflect.Ptr {
								field = field.Elem()
							}
						}

						primaryKeys := utils.ToArray(metaValue.Value)
						if len(primaryKeys) > 0 {
							context.GetDB().Where(primaryKeys).Find(field.Addr().Interface())
						}

						if !scope.PrimaryKeyZero() {
							context.GetDB().Model(resource).Association(meta.FieldName).Replace(field.Interface())
							field.Set(reflect.Zero(field.Type()))
						}
					})
				}
			}
		}
	}

	{ // Set Meta Resource
		if hasColumn {
			if meta.Resource == nil {
				var result interface{}

				if fieldType.Kind() == reflect.Struct {
					result = reflect.New(fieldType).Interface()
				} else if fieldType.Kind() == reflect.Slice {
					refelectType := fieldType.Elem()
					for refelectType.Kind() == reflect.Ptr {
						refelectType = refelectType.Elem()
					}
					if refelectType.Kind() == reflect.Struct {
						result = reflect.New(refelectType).Interface()
					}
				}

				if result != nil {
					res := meta.baseResource.GetAdmin().NewResource(result)
					meta.Resource = res
					meta.Meta.Permission = meta.Meta.Permission.Concat(res.Config.Permission)
				}
			}

			if meta.Resource != nil {
				permission := meta.Resource.Permission.Concat(meta.Meta.Permission)
				meta.Resource.Permission = permission
				meta.SetPermission(permission)
				meta.setBaseResource(meta.baseResource)
			}
		}
	}

	meta.FieldName = meta.GetFieldName()

	// run meta configors
	if baseResource := meta.baseResource; baseResource != nil {
		for key, fc := range baseResource.GetAdmin().metaConfigorMaps {
			if key == meta.Type {
				fc(meta)
			}
		}
	}

	// call meta config's ConfigureMetaInterface
	if meta.Config != nil {
		meta.Config.ConfigureQorMeta(meta)
	}

	// call field's ConfigureMetaInterface
	if meta.FieldStruct != nil {
		if injector, ok := reflect.New(meta.FieldStruct.Struct.Type).Interface().(resource.ConfigureMetaInterface); ok {
			injector.ConfigureQorMeta(meta)
		}
	}
}
