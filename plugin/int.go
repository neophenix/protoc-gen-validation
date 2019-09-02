package plugin

import (
	"fmt"

	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	pb "github.com/neophenix/protoc-gen-validation"
)

func (p *Plugin) generateIntValidationCode(fieldName string, fieldValue string, v *pb.FieldValidation, mv *pb.MessageValidation, field *descriptor.FieldDescriptorProto) {
	if v.IntEq != nil {
		p.P(`if %s != %d {`, fieldValue, v.GetIntEq())
		p.generateErrorCode(fieldName, fmt.Sprintf("%d", v.GetIntEq()), "{field} must equal {value}", v, mv, field, "")
		p.P(`}`)
	}
	if v.IntLte != nil {
		p.P(`if %s > %d {`, fieldValue, v.GetIntLte())
		p.generateErrorCode(fieldName, fmt.Sprintf("%d", v.GetIntLte()), "{field} must be less than or equal to  {value}", v, mv, field, "")
		p.P(`}`)
	}
	if v.IntGte != nil {
		p.P(`if %s < %d {`, fieldValue, v.GetIntGte())
		p.generateErrorCode(fieldName, fmt.Sprintf("%d", v.GetIntGte()), "{field} must be greater than or equal to  {value}", v, mv, field, "")
		p.P(`}`)
	}
}

func isInt(field *descriptor.FieldDescriptorProto) bool {
	switch field.GetType() {
	case descriptor.FieldDescriptorProto_TYPE_INT32:
		return true
	case descriptor.FieldDescriptorProto_TYPE_INT64:
		return true
	case descriptor.FieldDescriptorProto_TYPE_SINT32:
		return true
	case descriptor.FieldDescriptorProto_TYPE_SINT64:
		return true
	case descriptor.FieldDescriptorProto_TYPE_UINT32:
		return true
	case descriptor.FieldDescriptorProto_TYPE_UINT64:
		return true
	case descriptor.FieldDescriptorProto_TYPE_FIXED32:
		return true
	case descriptor.FieldDescriptorProto_TYPE_FIXED64:
		return true
	}
	if isWKTInt(field.GetTypeName()) {
		return true
	}
	return false
}
