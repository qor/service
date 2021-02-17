package admin

import (
	"testing"

	"github.com/qor/qor"
)

func generateResourceMenu(resource *Resource) *Menu {
	return &Menu{RelativePath: resource.ToParam(), Name: resource.Name}
}

func TestMenu(t *testing.T) {
	admin := New(&qor.Config{})
	admin.router.Prefix = "/admin"

	menu := &Menu{Name: "Dashboard", Link: "/link1"}
	admin.AddMenu(menu)

	if menu.URL() != "/link1" {
		t.Errorf("menu's URL should be correct")
	}

	if admin.GetMenu("Dashboard") == nil {
		t.Errorf("menu %v not added", "Dashboard")
	}

	menu2 := &Menu{Name: "Dashboard", RelativePath: "/link2"}
	admin.AddMenu(menu2)
	if menu2.URL() != "/admin/link2" {
		t.Errorf("menu's URL should be correct")
	}

	type Res struct{}
	admin.AddResource(&Res{})

	if menu := admin.GetMenu("Res"); menu == nil {
		t.Errorf("menu %v not added", "Res")
	} else if menu.URL() != "/admin/res" {
		t.Errorf("menu %v' URL should be correct, got %v", "Res", menu.URL())
	}

	admin.AddResource(&Res{}, &Config{Name: "Res2", Menu: []string{"management"}})

	if menu := admin.GetMenu("Res2"); menu == nil {
		t.Errorf("menu %v not added", "Res2")
	} else if menu.URL() != "/admin/res2" {
		t.Errorf("menu %v' URL should be correct, got %v", "Res2", menu.URL())
	} else if len(menu.Ancestors) != 1 || menu.Ancestors[0] != "management" {
		t.Errorf("menu %v' ancestors should be correct", "Res2")
	}

	if menu := admin.GetMenu("management", "Res2"); menu == nil {
		t.Errorf("menu %v not added", "Res2")
	} else if menu.URL() != "/admin/res2" {
		t.Errorf("menu %v' URL should be correct, got %v", "Res2", menu.URL())
	} else if len(menu.Ancestors) != 1 || menu.Ancestors[0] != "management" {
		t.Errorf("menu %v' ancestors should be correct", "Res2")
	}

	if menu := admin.GetMenu("management", "Res"); menu != nil {
		t.Errorf("menu management>Res should not found")
	}
}

func TestMenuPriority(t *testing.T) {
	admin := New(&qor.Config{})
	admin.router.Prefix = "/admin"

	admin.AddMenu(&Menu{Name: "Name1", Priority: 2})
	admin.AddMenu(&Menu{Name: "Name2", Priority: -1})
	admin.AddMenu(&Menu{Name: "Name3", Priority: 3})
	admin.AddMenu(&Menu{Name: "Name4", Priority: 4})
	admin.AddMenu(&Menu{Name: "Name5", Priority: 1})
	admin.AddMenu(&Menu{Name: "Name6", Priority: 0})
	admin.AddMenu(&Menu{Name: "Name7", Priority: -2})
	admin.AddMenu(&Menu{Name: "SubName1", Ancestors: []string{"Name5"}, Priority: 1})
	admin.AddMenu(&Menu{Name: "SubName2", Ancestors: []string{"Name5"}, Priority: 3})
	admin.AddMenu(&Menu{Name: "SubName3", Ancestors: []string{"Name5"}, Priority: -1})
	admin.AddMenu(&Menu{Name: "SubName4", Ancestors: []string{"Name5"}, Priority: 4})
	admin.AddMenu(&Menu{Name: "SubName5", Ancestors: []string{"Name5"}, Priority: -1})
	admin.AddMenu(&Menu{Name: "SubName1", Ancestors: []string{"Name1"}})
	admin.AddMenu(&Menu{Name: "SubName2", Ancestors: []string{"Name1"}, Priority: 2})
	admin.AddMenu(&Menu{Name: "SubName3", Ancestors: []string{"Name1"}, Priority: -2})
	admin.AddMenu(&Menu{Name: "SubName4", Ancestors: []string{"Name1"}, Priority: 1})
	admin.AddMenu(&Menu{Name: "SubName5", Ancestors: []string{"Name1"}, Priority: -1})

	menuNames := []string{"Name5", "Name1", "Name3", "Name4", "Name6", "Name7", "Name2"}
	submenuNames := []string{"SubName1", "SubName2", "SubName4", "SubName3", "SubName5"}
	submenuNames2 := []string{"SubName4", "SubName2", "SubName1", "SubName3", "SubName5"}
	for idx, menu := range admin.GetMenus() {
		if menuNames[idx] != menu.Name {
			t.Errorf("#%v menu should be %v, but got %v", idx, menuNames[idx], menu.Name)
		}

		if menu.Name == "Name5" {
			subMenus := menu.GetSubMenus()
			if len(subMenus) != 5 {
				t.Errorf("Should have 5 subMenus for Name5")
			}

			for idx, menu := range subMenus {
				if submenuNames[idx] != menu.Name {
					t.Errorf("#%v menu should be %v, but got %v", idx, submenuNames[idx], menu.Name)
				}
			}
		}

		if menu.Name == "Name1" {
			subMenus := menu.GetSubMenus()
			if len(subMenus) != 5 {
				t.Errorf("Should have 5 subMenus for Name1")
			}

			for idx, menu := range subMenus {
				if submenuNames2[idx] != menu.Name {
					t.Errorf("#%v menu should be %v, but got %v", idx, submenuNames2[idx], menu.Name)
				}
			}
		}
	}
}

func TestGetSelfMenuTree(t *testing.T) {
	admin := New(&qor.Config{})
	admin.router.Prefix = "/admin"

	gfMenu := admin.AddMenu(&Menu{Name: "Grandfather"})
	admin.AddMenu(&Menu{Name: "Father1", Ancestors: []string{"Grandfather"}})
	f2Menu := admin.AddMenu(&Menu{Name: "Father2", Ancestors: []string{"Grandfather"}})
	c1Menu := admin.AddMenu(&Menu{Name: "Child1", Ancestors: []string{"Grandfather", "Father1"}})

	f3Menu := admin.AddMenu(&Menu{Name: "Father3"})
	c2Menu := admin.AddMenu(&Menu{Name: "Child2", Ancestors: []string{"Father3"}})

	// return all end offsprings menus
	result := GetSelfMenuTree(gfMenu)
	if want, got := 2, len(result); want != got {
		t.Errorf("expect to have %d offspring menu but got %d", want, got)
	}

	for _, r := range result {
		if r != c1Menu && r != f2Menu {
			t.Error("expected c1 and f2 not all included in the result")
		}
	}

	// returns it only sub menu
	f3Result := GetSelfMenuTree(f3Menu)
	if want, got := 1, len(f3Result); want != got {
		t.Errorf("expect to have %d offspring menu but got %d", want, got)
	}
	if f3Result[0] != c2Menu {
		t.Error("expected c2 not included in the result")
	}

	// for an end offspring menu, return itself
	c2Result := GetSelfMenuTree(c2Menu)
	if want, got := 1, len(c2Result); want != got {
		t.Errorf("expect to have %d offspring menu but got %d", want, got)
	}
	if c2Result[0] != c2Menu {
		t.Error("expected c2 not included in the result")
	}
}
