package admin_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/now"
	"github.com/qor/admin"
	. "github.com/qor/admin/tests/dummy"
	qorTestUtils "github.com/qor/qor/test/utils"
	"github.com/theplant/testingutils"
)

func TestJSONTransformerEncode(t *testing.T) {
	qorTestUtils.ResetDBTables(db, &admin.Group{}, &Language{}, &Profile{}, &CreditCard{}, &User{})

	ctx := Admin.NewContext(nil, nil)
	ctx.Context.Roles = []string{Role_system_administrator}

	var (
		buffer          bytes.Buffer
		registeredAt    = now.MustParse("2017-01-01")
		jsonTransformer = &admin.JSONTransformer{}
		encoder         = admin.Encoder{
			Action:   "show",
			Resource: Admin.GetResource("User"),
			Context:  ctx,
			Result: &User{
				Active:       true,
				Model:        gorm.Model{ID: 1},
				Name:         "jinzhu",
				Role:         "admin",
				RegisteredAt: &registeredAt,
				CreditCard: CreditCard{
					Number: "411111111111",
					Issuer: "visa",
				},
				Profile: Profile{
					Name: "jinzhu",
					Phone: Phone{
						Num: "110",
					},
					Sex: "male",
				},
			},
		}
	)

	if err := jsonTransformer.Encode(&buffer, encoder); err != nil {
		t.Errorf("no error should returned when encode object to JSON")
	}

	var response, expect json.RawMessage
	json.Unmarshal(buffer.Bytes(), &response)

	jsonResponse := `{
        "Active": true,
        "Addresses": [],
        "Age": 0,
        "Avatar": "",
        "Company": "",
        "CreditCard": {
                "ID": 0,
                "Issuer": "visa",
                "Number": "411111111111"
        },
        "ID": 1,
        "Languages": null,
        "Name": "jinzhu",
        "Profile": {
                "ID": 0,
                "Name": "jinzhu",
                "Phone": {
                        "ID": 0,
                        "Num": "110"
                },
                "Sex": "male"
        },
        "RegisteredAt": "2017-01-01 00:00",
        "Role": "admin"
}`

	json.Unmarshal([]byte(jsonResponse), &expect)

	diff := testingutils.PrettyJsonDiff(expect, response)
	if len(diff) > 0 {
		t.Errorf("Got %v\n\n\n\n%v", string(buffer.Bytes()), diff)
	}
}

func TestJSONTransformerEncodeMap(t *testing.T) {
	var (
		buffer          bytes.Buffer
		jsonTransformer = &admin.JSONTransformer{}
		encoder         = admin.Encoder{
			Result: map[string]interface{}{"error": []error{errors.New("error1"), errors.New("error2")}},
		}
	)

	jsonTransformer.Encode(&buffer, encoder)

	except := "{\n\t\"error\": [\n\t\t\"error1\",\n\t\t\"error2\"\n\t]\n}"
	if except != buffer.String() {
		t.Errorf("Failed to decode errors map to JSON, except: %v, but got %v", except, buffer.String())
	}
}
