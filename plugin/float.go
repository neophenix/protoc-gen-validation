package plugin

import (
	"fmt"

	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	pb "github.com/neophenix/protoc-gen-validation"
)

func (p *Plugin) generateFloatValidationCode(fieldName string, fieldValue string, v *pb.FieldValidation, mv *pb.MessageValidation, field *descriptor.FieldDescriptorProto) {
	if v.FloatEq != nil {
		p.P(`if %s != %f {`, fieldValue, v.GetFloatEq())
		p.generateErrorCode(fieldName, fmt.Sprintf("%f", v.GetFloatEq()), "{field} must equal {value}", v, mv, field, "")
		p.P(`}`)
	}
	if v.FloatLte != nil {
		p.P(`if %s > %f {`, fieldValue, v.GetFloatLte())
		p.generateErrorCode(fieldName, fmt.Sprintf("%f", v.GetFloatLte()), "{field} must be less than or equal to {value}", v, mv, field, "")
		p.P(`}`)
	}
	if v.FloatGte != nil {
		p.P(`if %s < %f {`, fieldValue, v.GetFloatGte())
		p.generateErrorCode(fieldName, fmt.Sprintf("%f", v.GetFloatGte()), "{field} must be greater than or equal to {value}", v, mv, field, "")
		p.P(`}`)
	}
}

func isFloat(field *descriptor.FieldDescriptorProto) bool {
	switch field.GetType() {
	case descriptor.FieldDescriptorProto_TYPE_FLOAT:
		return true
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
		return true
	}
	if isWKTFloat(field.GetTypeName()) {
		return true
	}
	return false
}
