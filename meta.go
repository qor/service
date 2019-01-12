package admin

import (
	"database/sql"
	"math/big"
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

// MetaConfigInterface meta config interface
type MetaConfigInterface interface {
	resource.MetaConfigInterface
}

// Meta meta struct definition
type Meta struct {
	*resource.Meta
	Name            string
	FieldName       string
	Label           string
	Type            string
	Setter          func(record interface{}, metaValue *resource.MetaValue, context *qor.Context)
	Valuer          func(record interface{}, context *qor.Context) (result interface{})
	FormattedValuer func(record interface{}, context *qor.Context) (result interface{})
	Permission      *roles.Permission
	Config          MetaConfigInterface
	Collection      interface{}
	Resource        *Resource

	metas        []resource.Metaor
	baseResource *Resource
	processors   []*MetaProcessor
}

// SetPermission set meta's permission
func (meta *Meta) SetPermission(permission *roles.Permission) {
	meta.Permission = permission
	meta.Meta.Permission = permission
	if meta.Resource != nil {
		meta.Resource.Permission = permission
	}
}

// HasPermission check has permission or not
func (meta Meta) HasPermission(mode roles.PermissionMode, context *qor.Context) bool {
	var roles = []interface{}{}
	for _, role := range context.Roles {
		roles = append(roles, role)
	}
	if meta.Permission != nil {
		return meta.Permission.HasPermission(mode, roles...)
	}

	if meta.Resource != nil {
		return meta.Resource.HasPermission(mode, context)
	}

	if meta.baseResource != nil {
		return meta.baseResource.HasPermission(mode, context)
	}

	return true
}

// GetResource get resource from meta
func (meta *Meta) GetResource() resource.Resourcer {
	if meta.Resource == nil {
		return nil
	}
	return meta.Resource
}

// GetMetas get sub metas
func (meta *Meta) GetMetas() []resource.Metaor {
	if len(meta.metas) > 0 {
		return meta.metas
	} else if meta.Resource == nil {
		return []resource.Metaor{}
	} else {
		return meta.Resource.GetMetas([]string{})
	}
}

// MetaProcessor meta processor which will be run each time update Meta
type MetaProcessor struct {
	Name    string
	Handler func(*Meta)
}

// AddProcessor add meta processors, it will be run when add them and each time update Meta
func (meta *Meta) AddProcessor(processor *MetaProcessor) {
	if processor != nil && processor.Handler != nil {
		processor.Handler(meta)

		for idx, p := range meta.processors {
			if p.Name == processor.Name {
				meta.processors[idx] = processor
				return
			}
		}

		meta.processors = append(meta.processors, processor)
	}
}

func (meta *Meta) configure() {
	if meta.Meta == nil {
		meta.Meta = &resource.Meta{
			Name:            meta.Name,
			FieldName:       meta.FieldName,
			Setter:          meta.Setter,
			Valuer:          meta.Valuer,
			FormattedValuer: meta.FormattedValuer,
			BaseResource:    meta.baseResource,
			Resource:        meta.Resource,
			Permission:      meta.Permission,
			Config:          meta.Config,
		}
	} else {
		meta.Meta.Name = meta.Name
		meta.Meta.FieldName = meta.FieldName
		meta.Meta.Setter = meta.Setter
		meta.Meta.Valuer = meta.Valuer
		meta.Meta.FormattedValuer = meta.FormattedValuer
		meta.Meta.BaseResource = meta.baseResource
		meta.Meta.Resource = meta.Resource
		meta.Meta.Permission = meta.Permission
		meta.Meta.Config = meta.Config
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
					var tags = meta.FieldStruct.TagSettings
					if size, ok := tags["SIZE"]; ok {
						if i, _ := strconv.Atoi(size); i > 255 {
							meta.Type = "text"
						} else {
							meta.Type = "string"
						}
					} else if text, ok := tags["TYPE"]; ok && text == "text" {
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
					} else if _, ok := reflect.New(fieldType).Interface().(**big.Int); ok {
						meta.Type = "float"
					} else if _, ok := reflect.New(fieldType).Interface().(*big.Int); ok {
						meta.Type = "float"
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
							// set current field value to blank and replace it with new value
							field.Set(reflect.Zero(field.Type()))
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
					res := meta.baseResource.NewResource(result)
					meta.Resource = res
					meta.Meta.Permission = meta.Meta.Permission.Concat(res.Config.Permission)
				}
			}

			if meta.Resource != nil {
				permission := meta.Resource.Permission.Concat(meta.Meta.Permission)
				meta.Meta.Resource = meta.Resource
				meta.Resource.Permission = permission
				meta.SetPermission(permission)
			}
		}
	}

	meta.FieldName = meta.GetFieldName()

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

	// run meta configors
	if baseResource := meta.baseResource; baseResource != nil {
		for key, fc := range baseResource.GetAdmin().metaConfigorMaps {
			if key == meta.Type {
				fc(meta)
			}
		}
	}

	for _, processor := range meta.processors {
		processor.Handler(meta)
	}
}

// DBName get meta's db name, used in index page for sorting
func (meta *Meta) DBName() string {
	if meta.FieldStruct != nil {
		return meta.FieldStruct.DBName
	}
	return ""
}
