package admin

import (
	"encoding/xml"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/jinzhu/inflection"
	"github.com/qor/qor/utils"
	"github.com/qor/roles"
)

// XMLTransformer xml transformer
type XMLTransformer struct{}

// CouldEncode check if encodable
func (XMLTransformer) CouldEncode(encoder Encoder) bool {
	return true
}

// Encode encode encoder to writer as XML
func (XMLTransformer) Encode(writer io.Writer, encoder Encoder) error {
	xmlMarshaler := XMLStruct{
		Action:   encoder.Action,
		Resource: encoder.Resource,
		Context:  encoder.Context,
		Result:   encoder.Result,
	}

	xmlMarshalResult, err := xml.MarshalIndent(xmlMarshaler, "", "\t")

	if err != nil {
		xmlMarshaler.Result = map[string]string{"error": err.Error()}
		xmlMarshalResult, _ = xml.MarshalIndent(xmlMarshaler, "", "\t")
	}

	_, err = writer.Write([]byte(xml.Header + string(xmlMarshalResult)))
	return err
}

// XMLStruct used to decode resource to xml
type XMLStruct struct {
	Action   string
	Resource *Resource
	Context  *Context
	Result   interface{}
}

// Initialize initialize a resource to XML Transformer
func (xmlStruct XMLStruct) Initialize(value interface{}, res *Resource) XMLStruct {
	return XMLStruct{
		Resource: res,
		Action:   xmlStruct.Action,
		Context:  xmlStruct.Context,
		Result:   value,
	}
}

// MarshalXML implement MarshalXMLInterface
func (xmlStruct XMLStruct) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return XMLMarshalDefaultHandler(xmlStruct, e, start)
}

// XMLMarshalDefaultHandler default xml marshal handler, allow developers overwrite it
var XMLMarshalDefaultHandler = func(xmlStruct XMLStruct, e *xml.Encoder, start xml.StartElement) error {
	defaultStartElement := xml.StartElement{Name: xml.Name{Local: "XMLStruct"}}
	reflectValue := reflect.Indirect(reflect.ValueOf(xmlStruct.Result))
	res := xmlStruct.Resource
	context := xmlStruct.Context

	switch reflectValue.Kind() {
	case reflect.Map:
		// Write Start Element
		if start.Name.Local == defaultStartElement.Name.Local {
			start.Name.Local = "response"
		}

		if err := e.EncodeToken(start); err != nil {
			return err
		}

		mapKeys := reflectValue.MapKeys()
		for _, mapKey := range mapKeys {
			var (
				err       error
				mapValue  = reflectValue.MapIndex(mapKey)
				startElem = xml.StartElement{
					Name: xml.Name{Space: "", Local: fmt.Sprint(mapKey.Interface())},
					Attr: []xml.Attr{},
				}
			)

			mapValue = reflect.Indirect(reflect.ValueOf(mapValue.Interface()))
			if mapValue.Kind() == reflect.Map {
				err = e.EncodeElement(xmlStruct.Initialize(mapValue.Interface(), xmlStruct.Resource), startElem)
			} else {
				err = e.EncodeElement(fmt.Sprint(reflectValue.MapIndex(mapKey).Interface()), startElem)
			}

			if err != nil {
				return err
			}
		}
	case reflect.Slice:
		// Write Start Element
		if start.Name.Local == defaultStartElement.Name.Local {
			modelType := utils.ModelType(xmlStruct.Result)
			if xmlStruct.Resource != nil && modelType == utils.ModelType(xmlStruct.Resource.Value) {
				start.Name.Local = inflection.Plural(strings.Replace(xmlStruct.Resource.Name, " ", "", -1))
			} else {
				start.Name.Local = "responses"
			}
		}

		if err := e.EncodeToken(start); err != nil {
			return err
		}

		for i := 0; i < reflectValue.Len(); i++ {
			if err := e.EncodeElement(xmlStruct.Initialize(reflect.Indirect(reflectValue.Index(i)).Interface(), xmlStruct.Resource), defaultStartElement); err != nil {
				return err
			}
		}
	case reflect.Struct:
		// Write Start Element
		if xmlStruct.Resource == nil || utils.ModelType(xmlStruct.Result) != utils.ModelType(xmlStruct.Resource.Value) {
			if err := e.EncodeElement(fmt.Sprint(xmlStruct.Result), start); err != nil {
				return err
			}
		} else {
			if start.Name.Local == defaultStartElement.Name.Local {
				start.Name.Local = strings.Replace(xmlStruct.Resource.Name, " ", "", -1)
			}

			if err := e.EncodeToken(start); err != nil {
				return err
			}

			metas := []*Meta{}
			switch xmlStruct.Action {
			case "index":
				metas = res.ConvertSectionToMetas(res.allowedSections(res.IndexAttrs(), context, roles.Read))
			case "edit":
				metas = res.ConvertSectionToMetas(res.allowedSections(res.EditAttrs(), context, roles.Update))
			case "show":
				metas = res.ConvertSectionToMetas(res.allowedSections(res.ShowAttrs(), context, roles.Read))
			}

			for _, meta := range metas {
				if meta.HasPermission(roles.Read, context.Context) {
					metaStart := xml.StartElement{
						Name: xml.Name{
							Space: "",
							Local: strings.Replace(meta.Label, " ", "", -1),
						},
					}

					// has_one, has_many checker to avoid dead loop
					if meta.Resource != nil && (meta.FieldStruct != nil && meta.FieldStruct.Relationship != nil && (meta.FieldStruct.Relationship.Kind == "has_one" || meta.FieldStruct.Relationship.Kind == "has_many" || meta.Type == "single_edit" || meta.Type == "collection_edit")) {
						if err := e.EncodeElement(xmlStruct.Initialize(context.RawValueOf(xmlStruct.Result, meta), meta.Resource), metaStart); err != nil {
							return err
						}
					} else {
						if err := e.EncodeElement(context.FormattedValueOf(xmlStruct.Result, meta), metaStart); err != nil {
							return err
						}
					}
				}
			}
		}
	default:
		if reflectValue.IsValid() {
			if err := e.EncodeElement(fmt.Sprint(reflectValue.Interface()), start); err != nil {
				return err
			}
		} else {
			return nil
		}
	}

	// Write End Element
	if err := e.EncodeToken(xml.EndElement{Name: start.Name}); err != nil {
		return err
	}
	return nil
}
