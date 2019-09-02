package plugin

import (
	"fmt"

	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	pb "github.com/neophenix/protoc-gen-validation"
)

func (p *Plugin) generateStringValidationCode(fieldName string, fieldValue string, v *pb.FieldValidation, field *descriptor.FieldDescriptorProto) {
	if v.NotEmptyString != nil {
		p.P(`if %s == "" {`, fieldValue)
		p.generateErrorCode(fieldName, "", "{field} can not be an empty string", v, field, "")
		p.P(`}`)
	}
	if v.Matches != nil {
		p.P(`if %s != "%s" {`, fieldValue, v.GetMatches())
		p.generateErrorCode(fieldName, v.GetMatches(), "{field} must equal {value}", v, field, "")
		p.P(`}`)
	}
	if v.Contains != nil {
		p.P(`if !%s.Contains(%s, "%s") {`, p.stringsPkg.Use(), fieldValue, v.GetContains())
		p.generateErrorCode(fieldName, v.GetContains(), "{field} must contain {value}", v, field, "")
		p.P(`}`)
	}
	if v.Regex != nil {
		p.P(`if !%s.MustCompile("%s").MatchString(%s) {`, p.regexPkg.Use(), v.GetRegex(), fieldValue)
		p.generateErrorCode(fieldName, v.GetRegex(), "{field} must match regex {value}", v, field, "")
		p.P(`}`)
	}
	if v.MinLen != nil {
		p.P(`if len(%s) < %d {`, fieldValue, v.GetMinLen())
		p.generateErrorCode(fieldName, fmt.Sprintf("%d", v.GetMinLen()), "{field} must be at least {value} characters long", v, field, "")
		p.P(`}`)
	}
	if v.MaxLen != nil {
		p.P(`if len(%s) > %d {`, fieldValue, v.GetMaxLen())
		p.generateErrorCode(fieldName, fmt.Sprintf("%d", v.GetMaxLen()), "{field} must be no more than {value} characters long", v, field, "")
		p.P(`}`)
	}
	if v.EqLen != nil {
		p.P(`if len(%s) != %d {`, fieldValue, v.GetEqLen())
		p.generateErrorCode(fieldName, fmt.Sprintf("%d", v.GetEqLen()), "{field} must be exactly {value} characters long", v, field, "")
		p.P(`}`)
	}
	if v.IsUuid != nil {
		p.P(`if !isValidUUID(%s) {`, fieldValue)
		p.generateErrorCode(fieldName, "", "{field} be a valid UUID", v, field, "")
		p.P(`}`)
	}
	if v.IsEmail != nil {
		p.P(`if !isValidEmail(%s) {`, fieldValue)
		p.generateErrorCode(fieldName, "", "{field} must be a valid email address", v, field, "")
		p.P(`}`)
	}
}

func isString(field *descriptor.FieldDescriptorProto) bool {
	if field.GetType() == descriptor.FieldDescriptorProto_TYPE_STRING {
		return true
	}
	if isWKTString(field.GetTypeName()) {
		return true
	}
	return false
}
