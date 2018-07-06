package admin

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"math/rand"
	"net/url"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime/debug"
	"sort"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/inflection"
	"github.com/qor/qor"
	"github.com/qor/qor/utils"
	"github.com/qor/roles"
	"github.com/qor/session"
)

// FuncMap funcs map for current context
func (context *Context) FuncMap() template.FuncMap {
	funcMap := template.FuncMap{
		"current_user":         func() qor.CurrentUser { return context.CurrentUser },
		"get_resource":         context.Admin.GetResource,
		"new_resource_context": context.NewResourceContext,
		"is_new_record":        context.isNewRecord,
		"is_equal":             context.isEqual,
		"is_included":          context.isIncluded,
		"primary_key_of":       context.primaryKeyOf,
		"unique_key_of":        context.uniqueKeyOf,
		"formatted_value_of":   context.FormattedValueOf,
		"raw_value_of":         context.RawValueOf,

		"t": context.t,
		"flashes": func() []session.Message {
			return context.Admin.SessionManager.Flashes(context.Writer, context.Request)
		},
		"pagination": context.Pagination,
		"escape":     html.EscapeString,
		"raw":        func(str string) template.HTML { return template.HTML(utils.HTMLSanitizer.Sanitize(str)) },
		"unsafe_raw": func(str string) template.HTML { return template.HTML(str) },
		"equal":      equal,
		"stringify":  utils.Stringify,
		"lower": func(value interface{}) string {
			return strings.ToLower(fmt.Sprint(value))
		},
		"plural": func(value interface{}) string {
			return inflection.Plural(fmt.Sprint(value))
		},
		"singular": func(value interface{}) string {
			return inflection.Singular(fmt.Sprint(value))
		},
		"marshal": func(v interface{}) template.JS {
			switch value := v.(type) {
			case string:
				return template.JS(value)
			case template.HTML:
				return template.JS(value)
			default:
				byt, _ := json.Marshal(v)
				return template.JS(byt)
			}
		},

		"to_map": func(values ...interface{}) map[string]interface{} {
			results := map[string]interface{}{}
			for i := 0; i < len(values)-1; i += 2 {
				results[fmt.Sprint(values[i])] = values[i+1]
			}
			return results
		},
		"render":      context.Render,
		"render_text": context.renderText,
		"render_with": context.renderWith,
		"render_form": context.renderForm,
		"render_meta": func(value interface{}, meta *Meta, types ...string) template.HTML {
			var (
				result = bytes.NewBufferString("")
				typ    = "index"
			)

			for _, t := range types {
				typ = t
			}

			context.renderMeta(meta, value, []string{}, typ, result)
			return template.HTML(result.String())
		},
		"render_filter": context.renderFilter,
		"saved_filters": context.savedFilters,
		"has_filter": func() bool {
			query := context.Request.URL.Query()
			for key := range query {
				if regexp.MustCompile("filter[(\\w+)]").MatchString(key) && query.Get(key) != "" {
					return true
				}
			}
			return false
		},
		"page_title": context.pageTitle,
		"meta_label": func(meta *Meta) template.HTML {
			key := fmt.Sprintf("%v.attributes.%v", meta.baseResource.ToParam(), meta.Label)
			return context.Admin.T(context.Context, key, meta.Label)
		},
		"meta_placeholder": func(meta *Meta, context *Context, placeholder string) template.HTML {
			if getPlaceholder, ok := meta.Config.(interface {
				GetPlaceholder(*Context) (template.HTML, bool)
			}); ok {
				if str, ok := getPlaceholder.GetPlaceholder(context); ok {
					return str
				}
			}

			key := fmt.Sprintf("%v.attributes.%v.placeholder", meta.baseResource.ToParam(), meta.Label)
			return context.Admin.T(context.Context, key, placeholder)
		},

		"url_for":            context.URLFor,
		"link_to":            context.linkTo,
		"patch_current_url":  context.patchCurrentURL,
		"patch_url":          context.patchURL,
		"join_current_url":   context.joinCurrentURL,
		"join_url":           context.joinURL,
		"logout_url":         context.logoutURL,
		"search_center_path": func() string { return path.Join(context.Admin.router.Prefix, "!search") },
		"new_resource_path":  context.newResourcePath,
		"defined_resource_show_page": func(res *Resource) bool {
			if res != nil {
				if r := context.Admin.GetResource(utils.ModelType(res.Value).String()); r != nil {
					return r.sections.ConfiguredShowAttrs
				}
			}

			return false
		},

		"get_menus":                 context.getMenus,
		"get_scopes":                context.getScopes,
		"get_filters":               context.getFilters,
		"get_formatted_errors":      context.getFormattedErrors,
		"load_actions":              context.loadActions,
		"allowed_actions":           context.AllowedActions,
		"is_sortable_meta":          context.isSortableMeta,
		"index_sections":            context.indexSections,
		"show_sections":             context.showSections,
		"new_sections":              context.newSections,
		"edit_sections":             context.editSections,
		"convert_sections_to_metas": context.convertSectionToMetas,

		"has_create_permission": context.hasCreatePermission,
		"has_read_permission":   context.hasReadPermission,
		"has_update_permission": context.hasUpdatePermission,
		"has_delete_permission": context.hasDeletePermission,
		"has_change_permission": context.hasChangePermission,

		"qor_theme_class":        context.themesClass,
		"javascript_tag":         context.javaScriptTag,
		"stylesheet_tag":         context.styleSheetTag,
		"load_theme_stylesheets": context.loadThemeStyleSheets,
		"load_theme_javascripts": context.loadThemeJavaScripts,
		"load_admin_stylesheets": context.loadAdminStyleSheets,
		"load_admin_javascripts": context.loadAdminJavaScripts,
	}

	for key, value := range context.Admin.funcMaps {
		funcMap[key] = value
	}

	for key, value := range context.funcMaps {
		funcMap[key] = value
	}

	return funcMap
}

// NewResourceContext new context with resource
func (context *Context) NewResourceContext(name ...interface{}) *Context {
	clone := &Context{Context: context.Context.Clone(), Admin: context.Admin, Result: context.Result, Action: context.Action}
	if len(name) > 0 {
		if str, ok := name[0].(string); ok {
			clone.setResource(context.Admin.GetResource(str))
		} else if res, ok := name[0].(*Resource); ok {
			clone.setResource(res)
		}
	} else {
		clone.setResource(context.Resource)
	}
	return clone
}

// URLFor generate url for resource value
//     context.URLFor(&Product{})
//     context.URLFor(&Product{ID: 111})
//     context.URLFor(productResource)
func (context *Context) URLFor(value interface{}, resources ...*Resource) string {
	getPrefix := func(res *Resource) string {
		var params string
		for res.ParentResource != nil {
			params = path.Join(res.ParentResource.ToParam(), res.ParentResource.GetPrimaryValue(context.Request), params)
			res = res.ParentResource
		}
		return path.Join(res.GetAdmin().router.Prefix, params)
	}

	if admin, ok := value.(*Admin); ok {
		return admin.router.Prefix
	} else if res, ok := value.(*Resource); ok {
		return path.Join(getPrefix(res), res.ToParam())
	} else {
		var res *Resource

		if len(resources) > 0 {
			res = resources[0]
		}

		if res == nil {
			res = context.Admin.GetResource(reflect.Indirect(reflect.ValueOf(value)).Type().String())
		}

		if res != nil {
			if res.Config.Singleton {
				return path.Join(getPrefix(res), res.ToParam())
			}

			var (
				scope         = context.GetDB().NewScope(value)
				primaryFields []string
				primaryValues = map[string]string{}
			)

			for _, primaryField := range res.PrimaryFields {
				if field, ok := scope.FieldByName(primaryField.Name); ok {
					primaryFields = append(primaryFields, fmt.Sprint(field.Field.Interface())) // TODO improve me
				}
			}

			for _, field := range scope.PrimaryFields() {
				useAsPrimaryField := false
				for _, primaryField := range res.PrimaryFields {
					if field.DBName == primaryField.DBName {
						useAsPrimaryField = true
						break
					}
				}

				if !useAsPrimaryField {
					primaryValues[fmt.Sprintf("primary_key[%v_%v]", scope.TableName(), field.DBName)] = fmt.Sprint(reflect.Indirect(field.Field).Interface())
				}
			}

			urlPath := path.Join(getPrefix(res), res.ToParam(), strings.Join(primaryFields, ","))

			if len(primaryValues) > 0 {
				var primaryValueParams []string
				for key, value := range primaryValues {
					primaryValueParams = append(primaryValueParams, fmt.Sprintf("%v=%v", key, url.QueryEscape(value)))
				}
				urlPath = urlPath + "?" + strings.Join(primaryValueParams, "&")
			}
			return urlPath
		}
	}
	return ""
}

// RawValueOf return raw value of a meta for current resource
func (context *Context) RawValueOf(value interface{}, meta *Meta) interface{} {
	return context.valueOf(meta.GetValuer(), value, meta)
}

// FormattedValueOf return formatted value of a meta for current resource
func (context *Context) FormattedValueOf(value interface{}, meta *Meta) interface{} {
	result := context.valueOf(meta.GetFormattedValuer(), value, meta)
	if resultValuer, ok := result.(driver.Valuer); ok {
		if result, err := resultValuer.Value(); err == nil {
			return result
		}
	}

	return result
}

var visiblePageCount = 8

// Page contain pagination information
type Page struct {
	Page       int
	Current    bool
	IsPrevious bool
	IsNext     bool
	IsFirst    bool
	IsLast     bool
}

// PaginationResult pagination result struct
type PaginationResult struct {
	Pagination Pagination
	Pages      []Page
}

// Pagination return pagination information
// Keep visiblePageCount's pages visible, exclude prev and next link
// Assume there are 12 pages in total.
// When current page is 1
// [current, 2, 3, 4, 5, 6, 7, 8, next]
// When current page is 6
// [prev, 2, 3, 4, 5, current, 7, 8, 9, 10, next]
// When current page is 10
// [prev, 5, 6, 7, 8, 9, current, 11, 12]
// If total page count less than VISIBLE_PAGE_COUNT, always show all pages
func (context *Context) Pagination() *PaginationResult {
	var (
		pages      []Page
		pagination = context.Searcher.Pagination
		pageCount  = pagination.PerPage
	)

	if pageCount == 0 {
		if context.Resource != nil && context.Resource.Config.PageCount != 0 {
			pageCount = context.Resource.Config.PageCount
		} else {
			pageCount = PaginationPageCount
		}
	}

	if pagination.Total <= pageCount && pagination.CurrentPage <= 1 {
		return nil
	}

	start := pagination.CurrentPage - visiblePageCount/2
	if start < 1 {
		start = 1
	}

	end := start + visiblePageCount - 1 // -1 for "start page" itself
	if end > pagination.Pages {
		end = pagination.Pages
	}

	if (end-start) < visiblePageCount && start != 1 {
		start = end - visiblePageCount + 1
	}
	if start < 1 {
		start = 1
	}

	// Append prev link
	if start > 1 {
		pages = append(pages, Page{Page: 1, IsFirst: true})
		pages = append(pages, Page{Page: pagination.CurrentPage - 1, IsPrevious: true})
	}

	for i := start; i <= end; i++ {
		pages = append(pages, Page{Page: i, Current: pagination.CurrentPage == i})
	}

	// Append next link
	if end < pagination.Pages {
		pages = append(pages, Page{Page: pagination.CurrentPage + 1, IsNext: true})
		pages = append(pages, Page{Page: pagination.Pages, IsLast: true})
	}

	return &PaginationResult{Pagination: pagination, Pages: pages}
}

func (context *Context) primaryKeyOf(value interface{}) interface{} {
	if reflect.Indirect(reflect.ValueOf(value)).Kind() == reflect.Struct {
		scope := &gorm.Scope{Value: value}
		return fmt.Sprint(scope.PrimaryKeyValue())
	}
	return fmt.Sprint(value)
}

func (context *Context) uniqueKeyOf(value interface{}) interface{} {
	if reflect.Indirect(reflect.ValueOf(value)).Kind() == reflect.Struct {
		scope := &gorm.Scope{Value: value}
		var primaryValues []string
		for _, primaryField := range scope.PrimaryFields() {
			primaryValues = append(primaryValues, fmt.Sprint(primaryField.Field.Interface()))
		}
		primaryValues = append(primaryValues, fmt.Sprint(rand.Intn(1000)))
		return utils.ToParamString(url.QueryEscape(strings.Join(primaryValues, "_")))
	}
	return fmt.Sprint(value)
}

func (context *Context) isNewRecord(value interface{}) bool {
	if value == nil {
		return true
	}
	return context.GetDB().NewRecord(value)
}

func (context *Context) newResourcePath(res *Resource) string {
	return path.Join(context.URLFor(res), "new")
}

func (context *Context) linkTo(text interface{}, link interface{}) template.HTML {
	text = reflect.Indirect(reflect.ValueOf(text)).Interface()
	if linkStr, ok := link.(string); ok {
		return template.HTML(fmt.Sprintf(`<a href="%v">%v</a>`, linkStr, text))
	}
	return template.HTML(fmt.Sprintf(`<a href="%v">%v</a>`, context.URLFor(link), text))
}

func (context *Context) valueOf(valuer func(interface{}, *qor.Context) interface{}, value interface{}, meta *Meta) interface{} {
	if valuer != nil {
		reflectValue := reflect.ValueOf(value)
		if reflectValue.Kind() != reflect.Ptr {
			if !reflectValue.IsValid() {
				return nil
			}
			reflectPtr := reflect.New(reflectValue.Type())
			reflectPtr.Elem().Set(reflectValue)
			value = reflectPtr.Interface()
		}

		result := valuer(value, context.Context)

		if reflectValue := reflect.ValueOf(result); reflectValue.IsValid() {
			if reflectValue.Kind() == reflect.Ptr {
				if reflectValue.IsNil() || !reflectValue.Elem().IsValid() {
					return nil
				}

				result = reflectValue.Elem().Interface()
			}

			if meta.Type == "number" || meta.Type == "float" {
				if context.isNewRecord(value) && equal(reflect.Zero(reflect.TypeOf(result)).Interface(), result) {
					return nil
				}
			}
			return result
		}
		return nil
	}

	utils.ExitWithMsg(fmt.Sprintf("No valuer found for meta %v of resource %v", meta.Name, meta.baseResource.Name))
	return nil
}

func (context *Context) renderForm(value interface{}, sections []*Section) template.HTML {
	var result = bytes.NewBufferString("")
	context.renderSections(value, sections, []string{"QorResource"}, result, "form")
	return template.HTML(result.String())
}

func (context *Context) renderSections(value interface{}, sections []*Section, prefix []string, writer *bytes.Buffer, kind string) {
	for _, section := range sections {
		var rows []struct {
			Length      int
			ColumnsHTML template.HTML
		}

		for _, column := range section.Rows {
			columnsHTML := bytes.NewBufferString("")
			for _, col := range column {
				meta := section.Resource.GetMeta(col)
				if meta != nil {
					context.renderMeta(meta, value, prefix, kind, columnsHTML)
				}
			}

			rows = append(rows, struct {
				Length      int
				ColumnsHTML template.HTML
			}{
				Length:      len(column),
				ColumnsHTML: template.HTML(string(columnsHTML.Bytes())),
			})
		}

		if len(rows) > 0 {
			var data = map[string]interface{}{
				"Section": section,
				"Title":   template.HTML(section.Title),
				"Rows":    rows,
			}
			if content, err := context.Asset("metas/section.tmpl"); err == nil {
				if tmpl, err := template.New("section").Funcs(context.FuncMap()).Parse(string(content)); err == nil {
					tmpl.Execute(writer, data)
				}
			}
		}
	}
}

func (context *Context) renderFilter(filter *Filter) template.HTML {
	var (
		err     error
		content []byte
		result  = bytes.NewBufferString("")
	)

	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
			debug.PrintStack()
			result.WriteString(fmt.Sprintf("Get error when render template for filter %v (%v): %v", filter.Name, filter.Type, r))
		}
	}()

	if content, err = context.Asset(fmt.Sprintf("metas/filter/%v.tmpl", filter.Type)); err == nil {
		tmpl := template.New(filter.Type + ".tmpl").Funcs(context.FuncMap())
		if tmpl, err = tmpl.Parse(string(content)); err == nil {
			var data = map[string]interface{}{
				"Filter":          filter,
				"Label":           filter.Label,
				"InputNamePrefix": fmt.Sprintf("filters[%v]", filter.Name),
				"Context":         context,
				"Resource":        context.Resource,
			}

			err = tmpl.Execute(result, data)
		}
	}

	if err != nil {
		result.WriteString(fmt.Sprintf("got error when render filter template for %v(%v):%v", filter.Name, filter.Type, err))
	}

	return template.HTML(result.String())
}

func (context *Context) savedFilters() (filters []SavedFilter) {
	context.Admin.SettingsStorage.Get("saved_filters", &filters, context)
	return
}

func (context *Context) renderMeta(meta *Meta, value interface{}, prefix []string, metaType string, writer *bytes.Buffer) {
	var (
		err      error
		funcsMap = context.FuncMap()
	)
	prefix = append(prefix, meta.Name)

	var generateNestedRenderSections = func(kind string) func(interface{}, []*Section, int) template.HTML {
		return func(value interface{}, sections []*Section, index int) template.HTML {
			var result = bytes.NewBufferString("")
			var newPrefix = append([]string{}, prefix...)

			if index >= 0 {
				last := newPrefix[len(newPrefix)-1]
				newPrefix = append(newPrefix[:len(newPrefix)-1], fmt.Sprintf("%v[%v]", last, index))
			}

			if len(sections) > 0 {
				for _, field := range context.GetDB().NewScope(value).PrimaryFields() {
					if meta := sections[0].Resource.GetMeta(field.Name); meta != nil {
						context.renderMeta(meta, value, newPrefix, kind, result)
					}
				}

				context.renderSections(value, sections, newPrefix, result, kind)
			}

			return template.HTML(result.String())
		}
	}

	funcsMap["has_change_permission"] = func(permissioner HasPermissioner) bool {
		if context.GetDB().NewScope(value).PrimaryKeyZero() {
			return context.hasCreatePermission(permissioner)
		}
		return context.hasUpdatePermission(permissioner)
	}
	funcsMap["render_nested_form"] = generateNestedRenderSections("form")

	defer func() {
		if r := recover(); r != nil {
			debug.PrintStack()
			writer.WriteString(fmt.Sprintf("Get error when render template for meta %v (%v): %v", meta.Name, meta.Type, r))
		}
	}()

	var (
		tmpl    = template.New(meta.Type + ".tmpl").Funcs(funcsMap)
		content []byte
	)

	switch {
	case meta.Config != nil:
		if templater, ok := meta.Config.(interface {
			GetTemplate(context *Context, metaType string) ([]byte, error)
		}); ok {
			if content, err = templater.GetTemplate(context, metaType); err == nil {
				tmpl, err = tmpl.Parse(string(content))
				break
			}
		}
		fallthrough
	default:
		if content, err = context.Asset(fmt.Sprintf("%v/metas/%v/%v.tmpl", meta.baseResource.ToParam(), metaType, meta.Name), fmt.Sprintf("metas/%v/%v.tmpl", metaType, meta.Type)); err == nil {
			tmpl, err = tmpl.Parse(string(content))
		} else if metaType == "index" {
			tmpl, err = tmpl.Parse("{{.Value}}")
		} else {
			err = fmt.Errorf("haven't found %v template for meta %v", metaType, meta.Name)
		}
	}

	if err == nil {
		var scope = context.GetDB().NewScope(value)
		var data = map[string]interface{}{
			"Context":       context,
			"BaseResource":  meta.baseResource,
			"Meta":          meta,
			"ResourceValue": value,
			"Value":         context.FormattedValueOf(value, meta),
			"Label":         meta.Label,
			"InputName":     strings.Join(prefix, "."),
		}

		if !scope.PrimaryKeyZero() {
			data["InputId"] = utils.ToParamString(fmt.Sprintf("%v_%v_%v", scope.GetModelStruct().ModelType.Name(), scope.PrimaryKeyValue(), meta.Name))
		}

		data["CollectionValue"] = func() [][]string {
			fmt.Printf("%v: Call .CollectionValue from views already Deprecated, get the value with `.Meta.Config.GetCollection .ResourceValue .Context`", meta.Name)
			return meta.Config.(interface {
				GetCollection(value interface{}, context *Context) [][]string
			}).GetCollection(value, context)
		}

		err = tmpl.Execute(writer, data)
	}

	if err != nil {
		msg := fmt.Sprintf("got error when render %v template for %v(%v): %v", metaType, meta.Name, meta.Type, err)
		fmt.Fprint(writer, msg)
		utils.ExitWithMsg(msg)
	}
}

func (context *Context) isEqual(value interface{}, hasValue interface{}) bool {
	var result string

	if (value == nil || hasValue == nil) && (value != hasValue) {
		return false
	}

	if reflect.Indirect(reflect.ValueOf(hasValue)).Kind() == reflect.Struct {
		scope := &gorm.Scope{Value: hasValue}
		result = fmt.Sprint(scope.PrimaryKeyValue())
	} else {
		result = fmt.Sprint(hasValue)
	}

	reflectValue := reflect.Indirect(reflect.ValueOf(value))
	if reflectValue.Kind() == reflect.Struct {
		scope := &gorm.Scope{Value: value}
		return fmt.Sprint(scope.PrimaryKeyValue()) == result
	} else if reflectValue.Kind() == reflect.String {
		// type UserType string, alias type will panic if do
		// return reflectValue.Interface().(string) == result
		return fmt.Sprint(reflectValue.Interface()) == result
	} else {
		return fmt.Sprint(reflectValue.Interface()) == result
	}
}

func (context *Context) isIncluded(value interface{}, hasValue interface{}) bool {
	var result string
	if reflect.Indirect(reflect.ValueOf(hasValue)).Kind() == reflect.Struct {
		scope := &gorm.Scope{Value: hasValue}
		result = fmt.Sprint(scope.PrimaryKeyValue())
	} else {
		result = fmt.Sprint(hasValue)
	}

	primaryKeys := []interface{}{}
	reflectValue := reflect.Indirect(reflect.ValueOf(value))

	if reflectValue.Kind() == reflect.Slice {
		for i := 0; i < reflectValue.Len(); i++ {
			if value := reflectValue.Index(i); value.IsValid() {
				if reflect.Indirect(value).Kind() == reflect.Struct {
					scope := &gorm.Scope{Value: reflectValue.Index(i).Interface()}
					primaryKeys = append(primaryKeys, scope.PrimaryKeyValue())
				} else {
					primaryKeys = append(primaryKeys, reflect.Indirect(reflectValue.Index(i)).Interface())
				}
			}
		}
	} else if reflectValue.Kind() == reflect.Struct {
		scope := &gorm.Scope{Value: value}
		primaryKeys = append(primaryKeys, scope.PrimaryKeyValue())
	} else if reflectValue.Kind() == reflect.String {
		return strings.Contains(reflectValue.Interface().(string), result)
	} else if reflectValue.IsValid() {
		primaryKeys = append(primaryKeys, reflect.Indirect(reflectValue).Interface())
	}

	for _, key := range primaryKeys {
		if fmt.Sprint(key) == result {
			return true
		}
	}
	return false
}

func (context *Context) getResource(resources ...*Resource) *Resource {
	for _, res := range resources {
		return res
	}
	return context.Resource
}

func (context *Context) indexSections(resources ...*Resource) []*Section {
	res := context.getResource(resources...)
	return res.allowedSections(res.IndexAttrs(), context, roles.Read)
}

func (context *Context) editSections(resources ...*Resource) []*Section {
	res := context.getResource(resources...)
	return res.allowedSections(res.EditAttrs(), context, roles.Read)
}

func (context *Context) newSections(resources ...*Resource) []*Section {
	res := context.getResource(resources...)
	return res.allowedSections(res.NewAttrs(), context, roles.Create)
}

func (context *Context) showSections(resources ...*Resource) []*Section {
	res := context.getResource(resources...)
	return res.allowedSections(res.ShowAttrs(), context, roles.Read)
}

type menu struct {
	*Menu
	Active   bool
	SubMenus []*menu
}

func (context *Context) getMenus() (menus []*menu) {
	var (
		globalMenu        = &menu{}
		mostMatchedMenu   *menu
		mostMatchedLength int
		addMenu           func(*menu, []*Menu)
	)

	addMenu = func(parent *menu, menus []*Menu) {
		for _, m := range menus {
			url := m.URL()
			if m.HasPermission(roles.Read, context.Context) {
				var menu = &menu{Menu: m}
				if strings.HasPrefix(context.Request.URL.Path, url) && len(url) > mostMatchedLength {
					mostMatchedMenu = menu
					mostMatchedLength = len(url)
				}

				addMenu(menu, menu.GetSubMenus())
				if len(menu.SubMenus) > 0 || menu.URL() != "" {
					parent.SubMenus = append(parent.SubMenus, menu)
				}
			}
		}
	}

	addMenu(globalMenu, context.Admin.GetMenus())

	if context.Action != "search_center" && mostMatchedMenu != nil {
		mostMatchedMenu.Active = true
	}

	return globalMenu.SubMenus
}

type scope struct {
	*Scope
	Active bool
}

type scopeMenu struct {
	Group  string
	Scopes []*scope
}

// getScopes get scopes from current context
func (context *Context) getScopes() (menus []*scopeMenu) {
	if context.Resource == nil {
		return
	}

	activatedScopeNames := context.Request.URL.Query()["scopes"]
OUT:
	for _, s := range context.Resource.scopes {
		if s.Visible != nil && !s.Visible(s.Name, context) {
			continue
		}

		menu := &scope{Scope: s}

		for _, s := range activatedScopeNames {
			if s == menu.Name {
				menu.Active = true
			}
		}

		if menu.Group != "" {
			for _, m := range menus {
				if m.Group == menu.Group {
					m.Scopes = append(m.Scopes, menu)
					continue OUT
				}
			}
			menus = append(menus, &scopeMenu{Group: menu.Group, Scopes: []*scope{menu}})
		} else if !menu.Default {
			menus = append(menus, &scopeMenu{Group: menu.Group, Scopes: []*scope{menu}})
		}
	}

	for _, menu := range menus {
		hasActivedScope, hasDefaultScope := false, false
		for _, scope := range menu.Scopes {
			if scope.Active {
				hasActivedScope = true
			}
			if scope.Default {
				hasDefaultScope = true
			}
		}

		if hasDefaultScope && !hasActivedScope {
			for _, scope := range menu.Scopes {
				if scope.Default {
					scope.Active = true
				}
			}
		}
	}
	return menus
}

// getFilters get filters from current context
func (context *Context) getFilters() (filters []*Filter) {
	if context.Resource == nil {
		return
	}

	for _, f := range context.Resource.filters {
		if f.Visible == nil || f.Visible(context) {
			filters = append(filters, f)
		}
	}
	return
}

func (context *Context) hasCreatePermission(permissioner HasPermissioner) bool {
	return permissioner.HasPermission(roles.Create, context.Context)
}

func (context *Context) hasReadPermission(permissioner HasPermissioner) bool {
	return permissioner.HasPermission(roles.Read, context.Context)
}

func (context *Context) hasUpdatePermission(permissioner HasPermissioner) bool {
	return permissioner.HasPermission(roles.Update, context.Context)
}

func (context *Context) hasDeletePermission(permissioner HasPermissioner) bool {
	return permissioner.HasPermission(roles.Delete, context.Context)
}

func (context *Context) hasChangePermission(permissioner HasPermissioner) bool {
	if context.Action == "new" {
		return context.hasCreatePermission(permissioner)
	}
	return context.hasUpdatePermission(permissioner)
}

// PatchCurrentURL is a convinent wrapper for qor/utils.PatchURL
func (context *Context) patchCurrentURL(params ...interface{}) (patchedURL string, err error) {
	return utils.PatchURL(context.Request.URL.String(), params...)
}

// PatchURL is a convinent wrapper for qor/utils.PatchURL
func (context *Context) patchURL(url string, params ...interface{}) (patchedURL string, err error) {
	return utils.PatchURL(url, params...)
}

// JoinCurrentURL is a convinent wrapper for qor/utils.JoinURL
func (context *Context) joinCurrentURL(params ...interface{}) (joinedURL string, err error) {
	return utils.JoinURL(context.Request.URL.String(), params...)
}

// JoinURL is a convinent wrapper for qor/utils.JoinURL
func (context *Context) joinURL(url string, params ...interface{}) (joinedURL string, err error) {
	return utils.JoinURL(url, params...)
}

func (context *Context) themesClass() (result string) {
	var results = map[string]bool{}
	if context.Resource != nil {
		for _, theme := range context.Resource.Config.Themes {
			if strings.HasPrefix(theme.GetName(), "-") {
				results[strings.TrimPrefix(theme.GetName(), "-")] = false
			} else if _, ok := results[theme.GetName()]; !ok {
				results[theme.GetName()] = true
			}
		}
	}

	var names []string
	for name, enabled := range results {
		if enabled {
			names = append(names, "qor-theme-"+name)
		}
	}
	return strings.Join(names, " ")
}

func (context *Context) javaScriptTag(names ...string) template.HTML {
	var results []string
	for _, name := range names {
		name = path.Join(context.Admin.GetRouter().Prefix, "assets", "javascripts", name+".js")
		results = append(results, fmt.Sprintf(`<script src="%s"></script>`, name))
	}
	return template.HTML(strings.Join(results, ""))
}

func (context *Context) styleSheetTag(names ...string) template.HTML {
	var results []string
	for _, name := range names {
		name = path.Join(context.Admin.GetRouter().Prefix, "assets", "stylesheets", name+".css")
		results = append(results, fmt.Sprintf(`<link type="text/css" rel="stylesheet" href="%s">`, name))
	}
	return template.HTML(strings.Join(results, ""))
}

func (context *Context) getThemeNames() (themes []string) {
	themesMap := map[string]bool{}

	if context.Resource != nil {
		for _, theme := range context.Resource.Config.Themes {
			if _, ok := themesMap[theme.GetName()]; !ok {
				themes = append(themes, theme.GetName())
			}
		}
	}

	return
}

func (context *Context) loadThemeStyleSheets() template.HTML {
	var results []string
	for _, themeName := range context.getThemeNames() {
		var file = path.Join("themes", themeName, "assets", "stylesheets", themeName+".css")
		if _, err := context.Asset(file); err == nil {
			results = append(results, fmt.Sprintf(`<link type="text/css" rel="stylesheet" href="%s?theme=%s">`, path.Join(context.Admin.GetRouter().Prefix, "assets", "stylesheets", themeName+".css"), themeName))
		}
	}

	return template.HTML(strings.Join(results, " "))
}

func (context *Context) loadThemeJavaScripts() template.HTML {
	var results []string
	for _, themeName := range context.getThemeNames() {
		var file = path.Join("themes", themeName, "assets", "javascripts", themeName+".js")
		if _, err := context.Asset(file); err == nil {
			results = append(results, fmt.Sprintf(`<script src="%s?theme=%s"></script>`, path.Join(context.Admin.GetRouter().Prefix, "assets", "javascripts", themeName+".js"), themeName))
		}
	}

	return template.HTML(strings.Join(results, " "))
}

func (context *Context) loadAdminJavaScripts() template.HTML {
	var siteName = context.Admin.SiteName
	if siteName == "" {
		siteName = "application"
	}

	var file = path.Join("assets", "javascripts", strings.ToLower(strings.Replace(siteName, " ", "_", -1))+".js")
	if _, err := context.Asset(file); err == nil {
		return template.HTML(fmt.Sprintf(`<script src="%s"></script>`, path.Join(context.Admin.GetRouter().Prefix, file)))
	}
	return ""
}

func (context *Context) loadAdminStyleSheets() template.HTML {
	var siteName = context.Admin.SiteName
	if siteName == "" {
		siteName = "application"
	}

	var file = path.Join("assets", "stylesheets", strings.ToLower(strings.Replace(siteName, " ", "_", -1))+".css")
	if _, err := context.Asset(file); err == nil {
		return template.HTML(fmt.Sprintf(`<link type="text/css" rel="stylesheet" href="%s">`, path.Join(context.Admin.GetRouter().Prefix, file)))
	}
	return ""
}

func (context *Context) loadActions(action string) template.HTML {
	var (
		actionPatterns, actionKeys, actionFiles []string
		actions                                 = map[string]string{}
	)

	switch action {
	case "index", "show", "edit", "new":
		actionPatterns = []string{filepath.Join("actions", action, "*.tmpl"), "actions/*.tmpl"}

		if !context.Resource.sections.ConfiguredShowAttrs && action == "edit" {
			actionPatterns = []string{filepath.Join("actions", "show", "*.tmpl"), "actions/*.tmpl"}
		}
	case "global":
		actionPatterns = []string{"actions/*.tmpl"}
	default:
		actionPatterns = []string{filepath.Join("actions", action, "*.tmpl")}
	}

	for _, pattern := range actionPatterns {
		for _, themeName := range context.getThemeNames() {
			if resourcePath := context.resourcePath(); resourcePath != "" {
				if matches, err := context.Admin.AssetFS.Glob(filepath.Join("themes", themeName, resourcePath, pattern)); err == nil {
					actionFiles = append(actionFiles, matches...)
				}
			}

			if matches, err := context.Admin.AssetFS.Glob(filepath.Join("themes", themeName, pattern)); err == nil {
				actionFiles = append(actionFiles, matches...)
			}
		}

		if resourcePath := context.resourcePath(); resourcePath != "" {
			if matches, err := context.Admin.AssetFS.Glob(filepath.Join(resourcePath, pattern)); err == nil {
				actionFiles = append(actionFiles, matches...)
			}
		}

		if matches, err := context.Admin.AssetFS.Glob(pattern); err == nil {
			actionFiles = append(actionFiles, matches...)
		}
	}

	// before files have higher priority
	for _, actionFile := range actionFiles {
		base := regexp.MustCompile("^\\d+\\.").ReplaceAllString(path.Base(actionFile), "")

		if _, ok := actions[base]; !ok {
			actionKeys = append(actionKeys, path.Base(actionFile))
			actions[base] = actionFile
		}
	}

	sort.Strings(actionKeys)

	var result = bytes.NewBufferString("")
	for _, key := range actionKeys {
		defer func() {
			if r := recover(); r != nil {
				err := fmt.Sprintf("Get error when render action %v: %v", key, r)
				utils.ExitWithMsg(err)
				result.WriteString(err)
			}
		}()

		base := regexp.MustCompile("^\\d+\\.").ReplaceAllString(key, "")
		if content, err := context.Asset(actions[base]); err == nil {
			if tmpl, err := template.New(filepath.Base(actions[base])).Funcs(context.FuncMap()).Parse(string(content)); err == nil {
				if err := tmpl.Execute(result, context); err != nil {
					result.WriteString(err.Error())
					utils.ExitWithMsg(err)
				}
			} else {
				result.WriteString(err.Error())
				utils.ExitWithMsg(err)
			}
		}
	}

	return template.HTML(strings.TrimSpace(result.String()))
}

func (context *Context) logoutURL() string {
	if context.Admin.Auth != nil {
		return context.Admin.Auth.LogoutURL(context)
	}
	return ""
}

func (context *Context) t(values ...interface{}) template.HTML {
	switch len(values) {
	case 1:
		return context.Admin.T(context.Context, fmt.Sprint(values[0]), fmt.Sprint(values[0]))
	case 2:
		return context.Admin.T(context.Context, fmt.Sprint(values[0]), fmt.Sprint(values[1]))
	case 3:
		return context.Admin.T(context.Context, fmt.Sprint(values[0]), fmt.Sprint(values[1]), values[2:]...)
	default:
		utils.ExitWithMsg("passed wrong params for T")
	}
	return ""
}

func (context *Context) isSortableMeta(meta *Meta) bool {
	for _, attr := range context.Resource.SortableAttrs() {
		if attr == meta.Name && meta.FieldStruct != nil && meta.FieldStruct.IsNormal && meta.FieldStruct.DBName != "" {
			return true
		}
	}
	return false
}

func (context *Context) convertSectionToMetas(res *Resource, sections []*Section) []*Meta {
	return res.ConvertSectionToMetas(sections)
}

type formatedError struct {
	Label  string
	Errors []string
}

func (context *Context) getFormattedErrors() (formatedErrors []formatedError) {
	type labelInterface interface {
		Label() string
	}

	for _, err := range context.GetErrors() {
		if labelErr, ok := err.(labelInterface); ok {
			var found bool
			label := labelErr.Label()
			for _, formatedError := range formatedErrors {
				if formatedError.Label == label {
					formatedError.Errors = append(formatedError.Errors, err.Error())
				}
			}
			if !found {
				formatedErrors = append(formatedErrors, formatedError{Label: label, Errors: []string{err.Error()}})
			}
		} else {
			formatedErrors = append(formatedErrors, formatedError{Errors: []string{err.Error()}})
		}
	}
	return
}

// AllowedActions return allowed actions based on context
func (context *Context) AllowedActions(actions []*Action, mode string, records ...interface{}) []*Action {
	var allowedActions []*Action
	for _, action := range actions {
		for _, m := range action.Modes {
			if m == mode || (m == "index" && mode == "batch") {
				var permission = roles.Update
				switch strings.ToUpper(action.Method) {
				case "POST":
					permission = roles.Create
				case "DELETE":
					permission = roles.Delete
				case "PUT":
					permission = roles.Update
				case "GET":
					permission = roles.Read
				}

				if action.isAllowed(permission, context, records...) {
					allowedActions = append(allowedActions, action)
					break
				}
			}
		}
	}
	return allowedActions
}

func (context *Context) pageTitle() template.HTML {
	if context.Action == "search_center" {
		return context.t("qor_admin.search_center.title", "Search Center")
	}

	if context.Resource == nil {
		return context.t("qor_admin.layout.title", "Admin")
	}

	if context.Action == "action" {
		if action, ok := context.Result.(*Action); ok {
			return context.t(fmt.Sprintf("%v.actions.%v", context.Resource.ToParam(), action.Label), action.Label)
		}
	}

	var (
		defaultValue string
		titleKey     = fmt.Sprintf("qor_admin.form.%v.title", context.Action)
		usePlural    bool
	)

	switch context.Action {
	case "new":
		defaultValue = "Add {{$1}}"
	case "edit":
		defaultValue = "Edit {{$1}}"
	case "show":
		defaultValue = "{{$1}} Details"
	default:
		defaultValue = "{{$1}}"
		if !context.Resource.Config.Singleton {
			usePlural = true
		}
	}

	var resourceName string
	if usePlural {
		resourceName = string(context.t(fmt.Sprintf("%v.name.plural", context.Resource.ToParam()), inflection.Plural(context.Resource.Name)))
	} else {
		resourceName = string(context.t(fmt.Sprintf("%v.name", context.Resource.ToParam()), context.Resource.Name))
	}

	return context.t(titleKey, defaultValue, resourceName)
}
