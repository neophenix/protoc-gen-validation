package plugin

import (
	"fmt"

	pb "github.com/deelawn/protoc-gen-validation"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
)

func (p *Plugin) generateStringValidationCode(fieldName string, fieldValue string, v *pb.FieldValidation, mv *pb.MessageValidation, field *descriptor.FieldDescriptorProto) {

	if v.DoNotValidate != nil && *v.DoNotValidate {
		return
	}

	closeBrackets := 0

	if (v.Trim != nil && *v.Trim) || (mv != nil && mv.TrimStrings != nil && *mv.TrimStrings) {
		p.P(`%s = %s.Trim(%s, " ")`, fieldValue, p.stringsPkg.Use(), fieldValue)
	}
	if v.Lc != nil && *v.Lc {
		p.P(`%s = %s.ToLower(%s)`, fieldValue, p.stringsPkg.Use(), fieldValue)
	}
	if v.Uc != nil && *v.Uc {
		p.P(`%s = %s.ToUpper(%s)`, fieldValue, p.stringsPkg.Use(), fieldValue)
	}
	if v.NotEmptyString != nil {
		// For empty string checks, there is no point in doing furhter validation if we have an empty string, so
		// while it makes this code a bit uglier, try to build a decent looking if around further validation
		if *v.NotEmptyString {
			p.P(`if %s == "" {`, fieldValue)
			p.generateErrorCode(fieldName, "", "{field} can not be an empty string", v, mv, field, "")

			if getNumberOfValidationOptions(v) > 1 {
				// we will close this out at the end of this function
				p.P(`} else {`)
				closeBrackets++
			} else {
				p.P(`}`)
			}
		} else {
			if getNumberOfValidationOptions(v) > 1 {
				// we will close this out at the end of this function
				p.P(`if %s != "" {`, fieldValue)
				closeBrackets++
			}
		}
	}
	if v.Matches != nil {
		p.P(`if %s != "%s" {`, fieldValue, v.GetMatches())
		p.generateErrorCode(fieldName, v.GetMatches(), "{field} must equal {value}", v, mv, field, "")
		p.P(`}`)
	}
	if v.Contains != nil {
		p.P(`if !%s.Contains(%s, "%s") {`, p.stringsPkg.Use(), fieldValue, v.GetContains())
		p.generateErrorCode(fieldName, v.GetContains(), "{field} must contain {value}", v, mv, field, "")
		p.P(`}`)
	}
	if v.Regex != nil {
		p.P(`if !%s.MustCompile("%s").MatchString(%s) {`, p.regexPkg.Use(), v.GetRegex(), fieldValue)
		p.generateErrorCode(fieldName, v.GetRegex(), "{field} must match regex {value}", v, mv, field, "")
		p.P(`}`)
	}
	if v.MinLen != nil {
		p.P(`if len(%s) < %d {`, fieldValue, v.GetMinLen())
		p.generateErrorCode(fieldName, fmt.Sprintf("%d", v.GetMinLen()), "{field} must be at least {value} characters long", v, mv, field, "")
		p.P(`}`)
	}
	if v.MaxLen != nil {
		p.P(`if len(%s) > %d {`, fieldValue, v.GetMaxLen())
		p.generateErrorCode(fieldName, fmt.Sprintf("%d", v.GetMaxLen()), "{field} must be no more than {value} characters long", v, mv, field, "")
		p.P(`}`)
	}
	if v.EqLen != nil {
		p.P(`if len(%s) != %d {`, fieldValue, v.GetEqLen())
		p.generateErrorCode(fieldName, fmt.Sprintf("%d", v.GetEqLen()), "{field} must be exactly {value} characters long", v, mv, field, "")
		p.P(`}`)
	}
	if v.IsUuid != nil && *v.IsUuid {
		p.P(`if !isValidUUID(%s) {`, fieldValue)
		p.generateErrorCode(fieldName, "", "{field} must be a valid UUID", v, mv, field, "")
		p.P(`}`)
	}
	if v.IsEmail != nil && *v.IsEmail {
		p.P(`if !isValidEmail(%s) {`, fieldValue)
		p.generateErrorCode(fieldName, "", "{field} must be a valid email address", v, mv, field, "")
		p.P(`}`)
	}
	if v.IsIso8601Date != nil && *v.IsIso8601Date {
		p.P(`if !isValidDate("%s", %s) {`, "2006-01-02", fieldValue)
		p.generateErrorCode(fieldName, "", "{field} must be a date in the format YYYY-MM-DD", v, mv, field, "")
		p.P(`}`)
	}

	for i := 0; i < closeBrackets; i++ {
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

// getNumberOfValidationOptions figures out how many options are set for this field, useful if we need to adjust
// logic depending on if this is the only option
func getNumberOfValidationOptions(v *pb.FieldValidation) int {
	count := 0
	// we can use this as false too, so don't check for true
	if v.NotEmptyString != nil {
		count++
	}
	if v.Matches != nil {
		count++
	}
	if v.Contains != nil {
		count++
	}
	if v.Regex != nil {
		count++
	}
	if v.MinLen != nil {
		count++
	}
	if v.MaxLen != nil {
		count++
	}
	if v.EqLen != nil {
		count++
	}
	if v.IsUuid != nil && *v.IsUuid {
		count++
	}
	if v.IsEmail != nil && *v.IsEmail {
		count++
	}
	if v.IsIso8601Date != nil && *v.IsIso8601Date {
		count++
	}
	return count
}
