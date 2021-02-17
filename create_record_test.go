package admin_test

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/qor/admin"
	. "github.com/qor/admin/tests/dummy"
	qorTestUtils "github.com/qor/qor/test/utils"
)

func TestCreateRecord(t *testing.T) {
	form := url.Values{
		"QorResource.Name": {"create_record"},
	}

	if req, err := http.PostForm(server.URL+"/admin/users", form); err == nil {
		if req.StatusCode != 200 {
			t.Errorf("Create request should be processed successfully")
		}

		if db.First(&User{}, "name = ?", "create_record").RecordNotFound() {
			t.Errorf("User should be created successfully")
		}
	} else {
		t.Errorf(err.Error())
	}
}

func TestCreateBelongsToRecord(t *testing.T) {
	name := "create_belongs_to_record"
	form := url.Values{
		"QorResource.Name":              {name},
		"QorResource.Role":              {Role_system_administrator},
		"QorResource.CreditCard.Number": {"1234567890"},
		"QorResource.CreditCard.Issuer": {"Visa"},
	}

	if req, err := http.PostForm(server.URL+"/admin/users", form); err == nil {
		if req.StatusCode != 200 {
			t.Errorf("Create request should be processed successfully")
		}

		var user User
		if db.First(&user, "name = ?", name).RecordNotFound() {
			t.Errorf("User should be created successfully")
		}

		if db.Model(&user).Related(&user.CreditCard).RecordNotFound() || user.CreditCard.Number != "1234567890" {
			t.Errorf("Embedded struct should be created successfully")
		}
	} else {
		t.Errorf(err.Error())
	}
}

func TestCreateHasManyRecord(t *testing.T) {
	name := "create_record_and_has_many"
	form := url.Values{
		"QorResource.Name":                  {name},
		"QorResource.Role":                  {Role_system_administrator},
		"QorResource.Addresses[0].Address1": {"address_1"},
		"QorResource.Addresses[1].Address1": {"address_2"},
		"QorResource.Addresses[2].ID":       {"0"},
		"QorResource.Addresses[2].Address1": {""},
	}

	if req, err := http.PostForm(server.URL+"/admin/users", form); err == nil {
		if req.StatusCode != 200 {
			t.Errorf("Create request should be processed successfully")
		}

		var user User
		if db.First(&user, "name = ?", name).RecordNotFound() {
			t.Errorf("User should be created successfully")
		}

		if db.First(&Address{}, "user_id = ? and address1 = ?", user.ID, "address_1").RecordNotFound() {
			t.Errorf("Address 1 should be created successfully")
		}

		if db.First(&Address{}, "user_id = ? and address1 = ?", user.ID, "address_2").RecordNotFound() {
			t.Errorf("Address 2 should be created successfully")
		}

		var addresses []Address
		if db.Find(&addresses, "user_id = ?", user.ID); len(addresses) != 2 {
			t.Errorf("Blank address should not be created")
		}
	} else {
		t.Errorf(err.Error())
	}
}

func TestCreateHasManyRecordWithOrder(t *testing.T) {
	name := "create_record_and_has_many_with_order"
	form := url.Values{
		"QorResource.Name":                   {name},
		"QorResource.Role":                   {Role_system_administrator},
		"QorResource.Addresses[0].Address1":  {"address_0"},
		"QorResource.Addresses[1].Address1":  {"address_1"},
		"QorResource.Addresses[2].Address1":  {"address_2"},
		"QorResource.Addresses[11].Address1": {"address_11"},
	}

	if req, err := http.PostForm(server.URL+"/admin/users", form); err == nil {
		if req.StatusCode != 200 {
			t.Errorf("Create request should be processed successfully")
		}

		var user User
		if db.First(&user, "name = ?", name).RecordNotFound() {
			t.Errorf("User should be created successfully")
		}

		var address0, address1, address2, address11 Address
		if db.First(&address0, "user_id = ? and address1 = ?", user.ID, "address_0").RecordNotFound() {
			t.Errorf("Address 0 should be created successfully")
		}

		if db.First(&address1, "user_id = ? and address1 = ?", user.ID, "address_1").RecordNotFound() {
			t.Errorf("Address 1 should be created successfully")
		}

		if db.First(&address2, "user_id = ? and address1 = ?", user.ID, "address_2").RecordNotFound() {
			t.Errorf("Address 2 should be created successfully")
		}

		if db.First(&address11, "user_id = ? and address1 = ?", user.ID, "address_11").RecordNotFound() {
			t.Errorf("Address 11 should be created successfully")
		}

		if address11.ID < address2.ID || address2.ID < address1.ID || address1.ID < address0.ID {
			t.Errorf("Address should be created in order")
		}

		var addresses []Address
		if db.Find(&addresses, "user_id = ?", user.ID); len(addresses) != 4 {
			t.Errorf("There should be only %v addresses created", 4)
		}
	} else {
		t.Errorf(err.Error())
	}
}

func TestCreateManyToManyRecord(t *testing.T) {
	name := "create_record_many_to_many"
	var languageCN Language
	var languageEN Language
	db.FirstOrCreate(&languageCN, Language{Name: "CN"})
	db.FirstOrCreate(&languageEN, Language{Name: "EN"})

	form := url.Values{
		"QorResource.Name":      {name},
		"QorResource.Role":      {Role_system_administrator},
		"QorResource.Languages": {fmt.Sprintf("%d", languageCN.ID), fmt.Sprintf("%d", languageEN.ID)},
	}

	if req, err := http.PostForm(server.URL+"/admin/users", form); err == nil {
		if req.StatusCode != 200 {
			t.Errorf("Create request should be processed successfully")
		}

		var user User
		if db.First(&user, "name = ?", name).RecordNotFound() {
			t.Errorf("User should be created successfully")
		}

		var languages []Language
		db.Model(&user).Related(&languages, "Languages")

		if len(languages) != 2 {
			t.Errorf("User should have two languages after create")
		}
	} else {
		t.Errorf(err.Error())
	}
}

func TestUploadAttachment(t *testing.T) {
	name := "create_record_upload_attachment"

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if attachment, err := filepath.Abs("tests/qor.png"); err == nil {
		if part, err := writer.CreateFormFile("QorResource.Avatar", filepath.Base(attachment)); err == nil {
			if file, err := os.Open(attachment); err == nil {
				io.Copy(part, file)
			}
		}
		form := url.Values{
			"QorResource.Name": {name},
			"QorResource.Role": {Role_system_administrator},
		}
		for key, val := range form {
			_ = writer.WriteField(key, val[0])
		}
		writer.Close()

		if req, err := http.Post(server.URL+"/admin/users", writer.FormDataContentType(), body); err == nil {
			if req.StatusCode != 200 {
				t.Errorf("Create request should be processed successfully")
			}

			var user User
			if db.First(&user, "name = ?", name).RecordNotFound() {
				t.Errorf("User should be created successfully")
			}

			if user.Avatar.URL() == "" {
				t.Error("Avatar should be saved, but its URL is blank")
			}
		}
	}
}

func TestCreateRecordWithJSON(t *testing.T) {
	qorTestUtils.ResetDBTables(db, &admin.Group{}, &Language{}, &Address{}, &CreditCard{}, &User{})
	createLoggedInUser()

	name := "api_create_record"

	var languageCN Language
	var languageEN Language
	db.FirstOrCreate(&languageCN, Language{Name: "CN"})
	db.FirstOrCreate(&languageEN, Language{Name: "EN"})

	json := fmt.Sprintf(`{"Name":"api_create_record",
	                      "Role":"admin",
                          "CreditCard": {"Number": "987654321", "Issuer": "Visa"},
                          "Addresses": [{"Address1": "address_1"}, {"Address1": "address_2"}, {"_id": "0"}],
                          "Languages": [%v, %v]
                       }`, languageCN.ID, languageEN.ID)

	buf := strings.NewReader(json)

	if resp, err := http.Post(server.URL+"/admin/users", "application/json", buf); err == nil {
		if resp.StatusCode != 200 {
			t.Errorf("Create request should be processed successfully")
		}

		var user User
		if db.First(&user, "name = ?", name).RecordNotFound() {
			t.Errorf("User should be created successfully")
		}

		if db.Model(&user).Related(&user.CreditCard).RecordNotFound() || user.CreditCard.Number != "987654321" {
			t.Errorf("Embedded struct should be created successfully")
		}

		if db.First(&Address{}, "user_id = ? and address1 = ?", user.ID, "address_1").RecordNotFound() {
			t.Errorf("Address 1 should be created successfully")
		}

		if db.First(&Address{}, "user_id = ? and address1 = ?", user.ID, "address_2").RecordNotFound() {
			t.Errorf("Address 2 should be created successfully")
		}

		var addresses []Address
		if db.Find(&addresses, "user_id = ?", user.ID); len(addresses) != 2 {
			t.Errorf("Blank address should not be created")
		}

		var languages []Language
		db.Model(&user).Related(&languages, "Languages")

		if len(languages) != 2 {
			t.Errorf("User should have two languages after create")
		}
	}
}
