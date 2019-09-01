package plugin

import (
	"strings"
)

// all WKTs come from the google proto package
const wktBasePath = ".google.protobuf."

const (
	message = 1
	enum    = 2
)

// not sure what else to store as a val, so use something simple
var wktLookup = map[string]int8{
	"Any":               message,
	"Api":               message,
	"BoolValue":         message,
	"BytesValue":        message,
	"DoubleValue":       message,
	"Duration":          message,
	"Empty":             message,
	"Enum":              message,
	"EnumValue":         message,
	"Field":             message,
	"Field.Cardinality": enum,
	"Field.Kind":        enum,
	"FieldMask":         message,
	"FloatValue":        message,
	"Int32Value":        message,
	"Int64Value":        message,
	"ListValue":         message,
	"Method":            message,
	"Mixin":             message,
	"NullValue":         enum,
	"Option":            message,
	"SourceContext":     message,
	"StringValue":       message,
	"Struct":            message,
	"Syntax":            enum,
	"Timestamp":         message,
	"Type":              message,
	"UInt32Value":       message,
	"UInt64Value":       message,
	"Value":             message,
}

func isWKT(typeName string) bool {
	if strings.Contains(typeName, wktBasePath) {
		typeName = strings.ReplaceAll(typeName, wktBasePath, "")
		if _, ok := wktLookup[typeName]; ok {
			return true
		}
	}
	return false
}

func isWKTString(typeName string) bool {
	if typeName == wktBasePath+"StringValue" {
		return true
	}
	return false
}

func isWKTInt(typeName string) bool {
	if strings.Contains(typeName, "Int") {
		return true
	}
	return false
}

func isWKTFloat(typeName string) bool {
	if typeName == wktBasePath+"DoubleValue" {
		return true
	}
	if typeName == wktBasePath+"FloatValue" {
		return true
	}
	return false
}
