package admin

import (
	"crypto/md5"
	"fmt"
	"mime"
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/qor/responder"
)

// Controller admin controller
type Controller struct {
	*Admin
	action *Action
}

// HTTPUnprocessableEntity error status code
const HTTPUnprocessableEntity = 422

// Dashboard render dashboard page
func (ac *Controller) Dashboard(context *Context) {
	context.Execute("dashboard", nil)
}

// Index render index page
func (ac *Controller) Index(context *Context) {
	result, err := context.FindMany()
	context.AddError(err)

	responder.With("html", func() {
		context.Execute("index", result)
	}).With([]string{"json", "xml"}, func() {
		context.Encode("index", result)
	}).Respond(context.Request)
}

// SearchCenter render search center page
func (ac *Controller) SearchCenter(context *Context) {
	type Result struct {
		Context  *Context
		Resource *Resource
		Results  interface{}
	}
	var searchResults []Result

	for _, res := range context.GetSearchableResources() {
		var (
			resourceName = context.Request.URL.Query().Get("resource_name")
			ctx          = context.clone().setResource(res)
			searchResult = Result{Context: ctx, Resource: res}
		)

		if resourceName == "" || res.ToParam() == resourceName {
			searchResult.Results, _ = ctx.FindMany()
		}
		searchResults = append(searchResults, searchResult)
	}
	context.Execute("search_center", searchResults)
}

// New render new page
func (ac *Controller) New(context *Context) {
	context.Execute("new", context.Resource.NewStruct())
}

// Create create data
func (ac *Controller) Create(context *Context) {
	res := context.Resource
	status := http.StatusCreated
	result := res.NewStruct()
	if context.AddError(res.Decode(context.Context, result)); !context.HasError() {
		context.AddError(res.CallSave(result, context.Context))
	}

	if context.HasError() {
		responder.With("html", func() {
			context.Writer.WriteHeader(HTTPUnprocessableEntity)
			context.Execute("new", result)
		}).With([]string{"json", "xml"}, func() {
			context.Writer.WriteHeader(HTTPUnprocessableEntity)
			context.Encode("index", map[string]interface{}{"errors": context.GetErrors()})
		}).Respond(context.Request)
	} else {
		responder.With("html", func() {
			context.Flash(string(context.t("qor_admin.form.successfully_created", "{{.Name}} was successfully created", res)), "success")
			http.Redirect(context.Writer, context.Request, context.URLFor(result, res), http.StatusFound)
		}).With([]string{"json", "xml"}, func() {
			context.Writer.WriteHeader(status)
			context.Encode("show", result)
		}).Respond(context.Request)
	}
}

func (ac *Controller) renderSingleton(context *Context) (interface{}, bool, error) {
	var result interface{}
	var err error

	if context.Resource.Config.Singleton {
		result = context.Resource.NewStruct()
		if err = context.Resource.CallFindMany(result, context.Context); err == gorm.ErrRecordNotFound {
			context.Execute("new", result)
			return nil, true, nil
		}
	} else {
		result, err = context.FindOne()
	}

	if err == gorm.ErrRecordNotFound {
		context.Writer.WriteHeader(http.StatusNotFound)
	}
	return result, false, err
}

// Show render show page
func (ac *Controller) Show(context *Context) {
	result, rendered, err := ac.renderSingleton(context)
	if rendered {
		return
	}

	context.AddError(err)

	responder.With("html", func() {
		context.Execute("show", result)
	}).With([]string{"json", "xml"}, func() {
		context.Encode("show", result)
	}).Respond(context.Request)
}

// Edit render edit page
func (ac *Controller) Edit(context *Context) {
	result, rendered, err := ac.renderSingleton(context)
	if rendered {
		return
	}

	context.AddError(err)
	responder.With("html", func() {
		context.Execute("edit", result)
	}).With([]string{"json", "xml"}, func() {
		context.Encode("edit", result)
	}).Respond(context.Request)
}

// Update update data
func (ac *Controller) Update(context *Context) {
	var result interface{}
	var err error

	// If singleton Resource
	if context.Resource.Config.Singleton {
		result = context.Resource.NewStruct()
		context.Resource.CallFindMany(result, context.Context)
	} else {
		result, err = context.FindOne()
		context.AddError(err)
	}

	res := context.Resource
	if !context.HasError() {
		originalDB := context.DB
		tx := context.DB.Begin()
		context.SetDB(tx)
		if context.AddError(res.Decode(context.Context, result)); !context.HasError() {
			context.AddError(res.CallSave(result, context.Context))
		}
		if context.HasError() {
			tx.Rollback()
		} else {
			tx.Commit()
		}
		context.SetDB(originalDB)
	}

	if context.HasError() {
		context.Writer.WriteHeader(HTTPUnprocessableEntity)
		responder.With("html", func() {
			context.Execute("edit", result)
		}).With([]string{"json", "xml"}, func() {
			context.Encode("edit", map[string]interface{}{"errors": context.GetErrors()})
		}).Respond(context.Request)
	} else {
		responder.With("html", func() {
			context.Flash(string(context.t("qor_admin.form.successfully_updated", "{{.Name}} was successfully updated", res)), "success")
			context.Execute("show", result)
		}).With([]string{"json", "xml"}, func() {
			context.Encode("show", result)
		}).Respond(context.Request)
	}
}

// Delete delete data
func (ac *Controller) Delete(context *Context) {
	res := context.Resource
	status := http.StatusOK

	if context.AddError(res.CallDelete(res.NewStruct(), context.Context)); context.HasError() {
		context.Flash(string(context.t("qor_admin.form.failed_to_delete", "Failed to delete {{.Name}}", res)), "error")
		status = http.StatusNotFound
	}

	responder.With("html", func() {
		http.Redirect(context.Writer, context.Request, path.Join(ac.GetRouter().Prefix, res.ToParam()), http.StatusFound)
	}).With([]string{"json", "xml"}, func() {
		context.Writer.WriteHeader(status)
		context.Encode("OK", map[string]interface{}{"status": "ok"})
	}).Respond(context.Request)
}

// Action handle action related requests
func (ac *Controller) Action(context *Context) {
	var action = ac.action
	if context.Request.Method == "GET" {
		context.Execute("action", action)
	} else {
		var actionArgument = ActionArgument{
			PrimaryValues: context.Request.Form["primary_values[]"],
			Context:       context,
		}

		if primaryValue := context.Resource.GetPrimaryValue(context.Request); primaryValue != "" {
			actionArgument.PrimaryValues = append(actionArgument.PrimaryValues, primaryValue)
		}

		if action.Resource != nil {
			result := action.Resource.NewStruct()
			action.Resource.Decode(context.Context, result)
			actionArgument.Argument = result
		}

		context.AddError(action.Handler(&actionArgument))

		if !actionArgument.SkipDefaultResponse {
			if !context.HasError() {
				message := string(context.t("qor_admin.actions.executed_successfully", "Action {{.Name}}: Executed successfully", action))
				context.Flash(message, "success")
				responder.With("html", func() {
					http.Redirect(context.Writer, context.Request, context.Request.Referer(), http.StatusFound)
				}).With([]string{"json", "xml"}, func() {
					context.Encode("OK", map[string]interface{}{"message": message, "status": "ok"})
				}).Respond(context.Request)
			} else {
				context.Writer.WriteHeader(HTTPUnprocessableEntity)
				responder.With("html", func() {
					context.Execute("action", action)
				}).With([]string{"json", "xml"}, func() {
					var errs []string
					for _, err := range context.GetErrors() {
						errs = append(errs, err.Error())
					}
					context.Encode("OK", map[string]interface{}{"errors": errs, "status": "error"})
				}).Respond(context.Request)
			}
		}
	}
}

var (
	cacheSince = time.Now().Format(http.TimeFormat)
)

// Asset handle asset requests
func (ac *Controller) Asset(context *Context) {
	file := strings.TrimPrefix(context.Request.URL.Path, ac.GetRouter().Prefix)

	if context.Request.Header.Get("If-Modified-Since") == cacheSince {
		context.Writer.WriteHeader(http.StatusNotModified)
		return
	}
	context.Writer.Header().Set("Last-Modified", cacheSince)

	if content, err := context.Asset(file); err == nil {
		etag := fmt.Sprintf("%x", md5.Sum(content))
		if context.Request.Header.Get("If-None-Match") == etag {
			context.Writer.WriteHeader(http.StatusNotModified)
			return
		}

		if ctype := mime.TypeByExtension(filepath.Ext(file)); ctype != "" {
			context.Writer.Header().Set("Content-Type", ctype)
		}

		context.Writer.Header().Set("Cache-control", "private, must-revalidate, max-age=300")
		context.Writer.Header().Set("ETag", etag)
		context.Writer.Write(content)
	} else {
		http.NotFound(context.Writer, context.Request)
	}
}
