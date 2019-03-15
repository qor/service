package admin

import (
	"html/template"
	"path/filepath"
	"reflect"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/inflection"
	"github.com/qor/assetfs"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/utils"
	"github.com/qor/session"
	"github.com/qor/session/manager"
	"github.com/theplant/cldr"
)

// AdminConfig admin config struct
type AdminConfig struct {
	// SiteName set site's name, the name will be used as admin HTML title and admin interface will auto load javascripts, stylesheets files based on its value
	SiteName        string
	DB              *gorm.DB
	Auth            Auth
	AssetFS         assetfs.Interface
	SessionManager  session.ManagerInterface
	SettingsStorage SettingsStorageInterface
	I18n            I18n
	*Transformer
}

// Admin is a struct that used to generate admin/api interface
type Admin struct {
	*AdminConfig
	menus            []*Menu
	resources        []*Resource
	searchResources  []*Resource
	router           *Router
	funcMaps         template.FuncMap
	metaConfigorMaps map[string]func(*Meta)
}

// New new admin with configuration
func New(config interface{}) *Admin {
	admin := Admin{
		funcMaps:         make(template.FuncMap),
		router:           newRouter(),
		metaConfigorMaps: defaultMetaConfigorMaps,
	}

	if c, ok := config.(*qor.Config); ok {
		admin.AdminConfig = &AdminConfig{DB: c.DB}
	} else if c, ok := config.(*AdminConfig); ok {
		admin.AdminConfig = c
	} else {
		admin.AdminConfig = &AdminConfig{}
	}

	if admin.SessionManager == nil {
		admin.SessionManager = manager.SessionManager
	}

	if admin.Transformer == nil {
		admin.Transformer = DefaultTransformer
	}

	if admin.AssetFS == nil {
		admin.AssetFS = assetfs.AssetFS().NameSpace("admin")
	}

	if admin.SettingsStorage == nil {
		admin.SettingsStorage = newSettings(admin.AdminConfig.DB)
	}

	admin.SetAssetFS(admin.AssetFS)

	admin.registerCompositePrimaryKeyCallback()
	return &admin
}

// SetSiteName set site's name, the name will be used as admin HTML title and admin interface will auto load javascripts, stylesheets files based on its value
// For example, if you named it as `Qor Demo`, admin will look up `qor_demo.js`, `qor_demo.css` in QOR view paths, and load them if found
func (admin *Admin) SetSiteName(siteName string) {
	admin.SiteName = siteName
}

// SetAuth set admin's authorization gateway
func (admin *Admin) SetAuth(auth Auth) {
	admin.Auth = auth
}

// SetAssetFS set AssetFS for admin
func (admin *Admin) SetAssetFS(assetFS assetfs.Interface) {
	admin.AssetFS = assetFS
	globalAssetFSes = append(globalAssetFSes, assetFS)

	admin.AssetFS.RegisterPath(filepath.Join(utils.AppRoot, "app/views/qor"))
	admin.RegisterViewPath("github.com/qor/admin/views")

	for _, viewPath := range globalViewPaths {
		admin.RegisterViewPath(viewPath)
	}
}

// RegisterViewPath register view path for admin
func (admin *Admin) RegisterViewPath(pth string) {
	if admin.AssetFS.RegisterPath(filepath.Join(utils.AppRoot, "vendor", pth)) != nil {
		for _, gopath := range utils.GOPATH() {
			if admin.AssetFS.RegisterPath(filepath.Join(gopath, "src", pth)) == nil {
				break
			}
		}
	}
}

// RegisterMetaConfigor register configor for a kind, it will be called when register those kind of metas
func (admin *Admin) RegisterMetaConfigor(kind string, fc func(*Meta)) {
	admin.metaConfigorMaps[kind] = fc
}

// RegisterFuncMap register view funcs, it could be used in view templates
func (admin *Admin) RegisterFuncMap(name string, fc interface{}) {
	admin.funcMaps[name] = fc
}

// GetRouter get router from admin
func (admin *Admin) GetRouter() *Router {
	return admin.router
}

func (admin *Admin) newResource(value interface{}, config ...*Config) *Resource {
	var configuration *Config
	if len(config) > 0 {
		configuration = config[0]
	}

	if configuration == nil {
		configuration = &Config{}
	}

	res := &Resource{
		Resource: resource.New(value),
		Config:   configuration,
		admin:    admin,
	}

	res.Permission = configuration.Permission

	if configuration.Name != "" {
		res.Name = configuration.Name
	} else if namer, ok := value.(ResourceNamer); ok {
		res.Name = namer.ResourceName()
	}

	// Configure resource when initializing
	modelType := utils.ModelType(res.Value)
	for i := 0; i < modelType.NumField(); i++ {
		if fieldStruct := modelType.Field(i); fieldStruct.Anonymous {
			if injector, ok := reflect.New(fieldStruct.Type).Interface().(resource.ConfigureResourceBeforeInitializeInterface); ok {
				injector.ConfigureQorResourceBeforeInitialize(res)
			}
		}
	}

	if injector, ok := res.Value.(resource.ConfigureResourceBeforeInitializeInterface); ok {
		injector.ConfigureQorResourceBeforeInitialize(res)
	}

	findOneHandler := res.FindOneHandler
	res.FindOneHandler = func(result interface{}, metaValues *resource.MetaValues, context *qor.Context) error {
		if context.ResourceID == "" {
			context.ResourceID = res.GetPrimaryValue(context.Request)
		}
		return findOneHandler(result, metaValues, context)
	}

	res.UseTheme("slideout")
	return res
}

// NewResource initialize a new qor resource, won't add it to admin, just initialize it
func (admin *Admin) NewResource(value interface{}, config ...*Config) *Resource {
	res := admin.newResource(value, config...)
	res.Config.Invisible = true
	res.configure()
	return res
}

// AddResource make a model manageable from admin interface
func (admin *Admin) AddResource(value interface{}, config ...*Config) *Resource {
	res := admin.newResource(value, config...)
	admin.resources = append(admin.resources, res)

	res.configure()

	if !res.Config.Invisible {
		res.Action(&Action{
			Name:   "Delete",
			Method: "DELETE",
			URL: func(record interface{}, context *Context) string {
				return context.URLFor(record, res)
			},
			Permission: res.Config.Permission,
			Modes:      []string{"menu_item"},
		})

		menuName := res.Name
		if !res.Config.Singleton {
			menuName = inflection.Plural(res.Name)
		}
		admin.AddMenu(&Menu{Name: menuName, IconName: res.Config.IconName, Permissioner: res, Priority: res.Config.Priority, Ancestors: res.Config.Menu, RelativePath: res.ToParam()})

		admin.RegisterResourceRouters(res, "create", "update", "read", "delete")
	}

	return res
}

// GetResources get defined resources from admin
func (admin *Admin) GetResources() []*Resource {
	return admin.resources
}

// GetResource get resource with name
func (admin *Admin) GetResource(name string) (resource *Resource) {
	for _, res := range admin.resources {
		modelType := utils.ModelType(res.Value)
		// find with defined name first
		if res.ToParam() == name || res.Name == name || modelType.String() == name {
			return res
		}

		// if failed to find, use its model name
		if modelType.Name() == name {
			resource = res
		}
	}

	return
}

// AddSearchResource make a resource searchable from search center
func (admin *Admin) AddSearchResource(resources ...*Resource) {
	admin.searchResources = append(admin.searchResources, resources...)
}

// T call i18n backend to translate
func (admin *Admin) T(context *qor.Context, key string, value string, values ...interface{}) template.HTML {
	locale := utils.GetLocale(context)

	if admin.I18n == nil {
		if result, err := cldr.Parse(locale, value, values...); err == nil {
			return template.HTML(result)
		}
		return template.HTML(key)
	}

	return admin.I18n.Default(value).T(locale, key, values...)
}
