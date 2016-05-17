package admin

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/qor/qor"
	"github.com/qor/qor/utils"
)

// Context admin context, which is used for admin controller
type Context struct {
	*qor.Context
	*Searcher
	Flashes  []Flash
	Resource *Resource
	Admin    *Admin
	Content  template.HTML
	Action   string
	Result   interface{}
	Keys     map[string]interface{}
}

// NewContext new admin context
func (admin *Admin) NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{Context: &qor.Context{Config: admin.Config, Request: r, Writer: w}, Admin: admin}
}

// Set is used to store a new key/value pair exclusively for this context.
// It also lazily initializes context.Keys if it was not used previously.
func (context *Context) Set(key string, value interface{}) {
	if context.Keys == nil {
		context.Keys = make(map[string]interface{})
	}
	context.Keys[key] = value
}

// Get returns the value for the given key, ie: (value, true).
// If the value does not exists it returns (nil, false).
func (context *Context) Get(key string) (value interface{}, exists bool) {
	if context.Keys != nil {
		value, exists = context.Keys[key]
	}
	return
}

// MustGet returns the value for the given key if it exists, otherwise it panics.
func (context *Context) MustGet(key string) interface{} {
	if value, exists := context.Get(key); exists {
		return value
	}
	panic("Key \"" + key + "\" does not exist")
}

func (context *Context) clone() *Context {
	return &Context{
		Context:  context.Context,
		Searcher: context.Searcher,
		Flashes:  context.Flashes,
		Resource: context.Resource,
		Admin:    context.Admin,
		Result:   context.Result,
		Content:  context.Content,
		Action:   context.Action,
	}
}

func (context *Context) resourcePath() string {
	if context.Resource == nil {
		return ""
	}
	return context.Resource.ToParam()
}

func (context *Context) setResource(res *Resource) *Context {
	if res != nil {
		context.Resource = res
		context.ResourceID = res.GetPrimaryValue(context.Request)
	}
	context.Searcher = &Searcher{Context: context}
	return context
}

func (context *Context) Asset(layouts ...string) ([]byte, error) {
	var prefixes, themes []string

	if context.Request != nil {
		if theme := context.Request.URL.Query().Get("theme"); theme != "" {
			themes = append(themes, theme)
		}
	}

	if len(themes) == 0 && context.Resource != nil {
		themes = append(themes, context.Resource.Config.Themes...)
	}

	if resourcePath := context.resourcePath(); resourcePath != "" {
		for _, theme := range themes {
			prefixes = append(prefixes, filepath.Join("themes", theme, resourcePath))
		}
		prefixes = append(prefixes, resourcePath)
	}

	for _, theme := range themes {
		prefixes = append(prefixes, filepath.Join("themes", theme))
	}

	for _, layout := range layouts {
		for _, prefix := range prefixes {
			if content, err := context.Admin.AssetFS.Asset(filepath.Join(prefix, layout)); err == nil {
				return content, nil
			}
		}

		if content, err := context.Admin.AssetFS.Asset(layout); err == nil {
			return content, nil
		}
	}

	return []byte(""), fmt.Errorf("template not found: %v", layouts)
}

// Render render template based on context
func (context *Context) Render(name string, results ...interface{}) template.HTML {
	var (
		err     error
		content []byte
	)

	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("Get error when render file %v: %v", name, r))
			utils.ExitWithMsg(err)
		}
	}()

	if content, err = context.Asset(name + ".tmpl"); err == nil {
		var clone = context.clone()
		var result = bytes.NewBufferString("")

		if len(results) > 0 {
			clone.Result = results[0]
		}

		var tmpl *template.Template
		if tmpl, err = template.New(filepath.Base(name)).Funcs(clone.FuncMap()).Parse(string(content)); err == nil {
			if err = tmpl.Execute(result, clone); err == nil {
				return template.HTML(result.String())
			}
		}
	}

	return template.HTML(err.Error())
}

// Execute execute template with layout
func (context *Context) Execute(name string, result interface{}) {
	var tmpl *template.Template

	if name == "show" && !context.Resource.isSetShowAttrs {
		name = "edit"
	}

	if context.Action == "" {
		context.Action = name
	}

	if content, err := context.Asset("layout.tmpl"); err == nil {
		if tmpl, err = template.New("layout").Funcs(context.FuncMap()).Parse(string(content)); err == nil {
			for _, name := range []string{"header", "footer"} {
				if tmpl.Lookup(name) == nil {
					if content, err := context.Asset(name + ".tmpl"); err == nil {
						tmpl.Parse(string(content))
					}
				} else {
					utils.ExitWithMsg(err)
				}
			}
		} else {
			utils.ExitWithMsg(err)
		}
	}

	context.Result = result
	context.Content = context.Render(name, result)
	if err := tmpl.Execute(context.Writer, context); err != nil {
		utils.ExitWithMsg(err)
	}
}

// JSON generate json outputs for action
func (context *Context) JSON(action string, result interface{}) {
	if action == "show" && !context.Resource.isSetShowAttrs {
		action = "edit"
	}

	js, _ := json.MarshalIndent(context.Resource.convertObjectToJSONMap(context, result, action), "", "\t")
	context.Writer.Header().Set("Content-Type", "application/json")
	context.Writer.Write(js)
}
