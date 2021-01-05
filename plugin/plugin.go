package plugin

import (
	"fmt"
	"strings"

	pb "github.com/neophenix/protoc-gen-validation"

	"github.com/gogo/protobuf/gogoproto"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
)

type Plugin struct {
	gen        *generator.Generator
	imp        generator.PluginImports
	regexPkg   generator.Single
	stringsPkg generator.Single
	mailPkg    generator.Single
	timePkg    generator.Single
	uuidPkg    generator.Single
	strconvPkg generator.Single
}

func New() generator.Plugin {
	return &Plugin{}
}

func (p *Plugin) Name() string {
	return "validator"
}

func (p *Plugin) Init(g *generator.Generator) {
	p.gen = g
	p.imp = generator.NewPluginImports(p.gen)
}

// Here is where all the work is done, this is called to generate the main guts of the code
func (p *Plugin) Generate(file *generator.FileDescriptor) {
	if len(file.FileDescriptorProto.Service) == 0 {
		return
	}

	// output our error type
	p.generateErrorType()

	p.regexPkg = p.imp.NewImport("regexp")
	p.stringsPkg = p.imp.NewImport("strings")
	p.mailPkg = p.imp.NewImport("net/mail")
	p.timePkg = p.imp.NewImport("time")
	p.uuidPkg = p.imp.NewImport("github.com/google/uuid")
	p.strconvPkg = p.imp.NewImport("strconv")
	for _, msg := range file.Messages() {
		if msg.DescriptorProto.GetOptions().GetMapEntry() {
			continue
		}

		if gogoproto.IsProto3(file.FileDescriptorProto) {
			p.generateProto3(file, msg)
		}
	}

	// Helper funcs we can just generate even if we don't use them
	p.generateHelperFunctions()
}

// Remember that this is called last, so that we can mark imports as used in Generate and then they get output here.
// So don't go adding things here that expect to be first
func (p *Plugin) GenerateImports(file *generator.FileDescriptor) {
	if len(file.FileDescriptorProto.Service) == 0 {
		return
	}
	p.imp.GenerateImports(file)
}

// I lied above, this is actaully where all the code gets generated, at least for proto3
func (p *Plugin) generateProto3(file *generator.FileDescriptor, message *generator.Descriptor) {
	// begin Validate for this message
	p.P("func (m *%s) Validate() error {", message.GetName())
	p.P("err := ValidationErrors{Errors: []*ValidationError{}}")
	// if the message is nil, we can't validate it.  This should be ok to do here and will only be for "top level" messages
	// any embedded messages we already check to make sure they aren't nil before we call Validate on them down below
	p.P("if m == nil {")
	p.P(`err.Errors = []*ValidationError{&ValidationError{Field: "message", ErrorMessage: "message is nil, validation can not proceed"}}`)
	p.P(`return &err`)
	p.P("}")

	mv := getMessageValidation(message)

	for _, field := range message.Field {
		v := getFieldValidation(field)

		// Do not validate if this is set
		if v != nil && v.DoNotValidate != nil {
			continue
		}

		if field.IsMessage() {
			if isWKT(field.GetTypeName()) {
				if v != nil {
					p.P("if m.%s != nil {", generator.CamelCase(field.GetName()))
					p.generateValidationCode(field, v, mv)
					p.P("}")
				}
			} else if p.gen.IsMap(field) {
				p.P("// field:[%s] - maps not supported yet", field.GetName())
			} else {
				if field.IsRepeated() {
					p.P("for i, v := range m.%s {", generator.CamelCase(field.GetName()))
					p.P("msgerr := v.Validate()")
					p.P("if msgerr != nil {")
					p.P("if msgvalerr, ok := msgerr.(*ValidationErrors); ok {")
					p.generateErrorCode(generator.CamelCase(field.GetName()), "", "error in repeated value {field}", v, mv, field, "msgvalerr")
					p.P("}")
					p.P("}")
					p.P("}")
				} else {
					p.P("if m.%s != nil { ", generator.CamelCase(field.GetName()))
					p.P("msgerr := m.%s.Validate()", generator.CamelCase(field.GetName()))
					p.P("if msgerr != nil {")
					p.P("if msgvalerr, ok := msgerr.(*ValidationErrors); ok {")
					p.generateErrorCode(generator.CamelCase(field.GetName()), "", "error in {field}", v, mv, field, "msgvalerr")
					p.P("}")
					p.P("}")
					p.P("}")
				}
			}
		} else {
			if field.IsRepeated() && v != nil {
				p.P("for i, _ := range m.%s {", generator.CamelCase(field.GetName()))
				p.generateValidationCode(field, v, mv)
				p.P("}")
			} else {
				p.generateValidationCode(field, v, mv)
			}
		}
	}
	// return any error and close Validate for this message
	// but only return errors here if we aren't returning on individual errors as defined by message options
	if mv == nil || mv.ReturnOnError == nil || !mv.GetReturnOnError() {
		p.P("if len(err.Errors) != 0 { return &err }")
	}
	p.P("return nil")
	p.P("}")
}

// P forwards to p.gen.P after a Sprintf
func (p *Plugin) P(s string, args ...interface{}) { p.gen.P(fmt.Sprintf(s, args...)) }

func (p *Plugin) generateValidationCode(field *descriptor.FieldDescriptorProto, v *pb.FieldValidation, mv *pb.MessageValidation) {
	if v == nil {
		return
	}

	fieldValueAccessor := "m." + generator.CamelCase(field.GetName())
	fieldName := field.GetName()
	if isWKT(field.GetTypeName()) {
		fieldValueAccessor = "m." + generator.CamelCase(field.GetName()) + ".Value"
	}
	if field.IsRepeated() {
		fieldValueAccessor = "m." + generator.CamelCase(field.GetName()) + "[i]"
	}

	if v.TransformFunc != nil {
		p.P("%s = %s(%s)", fieldValueAccessor, *v.TransformFunc, fieldValueAccessor)
	}

	if isString(field) {
		p.generateStringValidationCode(fieldName, fieldValueAccessor, v, mv, field)
	} else if isInt(field) {
		p.generateIntValidationCode(fieldName, fieldValueAccessor, v, mv, field)
	} else if isFloat(field) {
		p.generateFloatValidationCode(fieldName, fieldValueAccessor, v, mv, field)
	}
}

func getFieldValidation(field *descriptor.FieldDescriptorProto) *pb.FieldValidation {
	if field.Options != nil {
		v, err := proto.GetExtension(field.Options, pb.E_Field)
		if err == nil && v.(*pb.FieldValidation) != nil {
			return (v.(*pb.FieldValidation))
		}
	}
	return nil
}

func getMessageValidation(msg *generator.Descriptor) *pb.MessageValidation {
	if msg.Options != nil {
		v, err := proto.GetExtension(msg.Options, pb.E_Message)
		if err == nil && v.(*pb.MessageValidation) != nil {
			return (v.(*pb.MessageValidation))
		}
	}
	return nil
}

func (p *Plugin) generateHelperFunctions() {
	isValidUUID := `func isValidUUID(u string) bool {
		_, err := ` + p.uuidPkg.Use() + `.Parse(u)
		return err == nil
	}
	`
	p.P(isValidUUID)

	isValidEmail := `func isValidEmail(e string) bool {
		_, err := ` + p.mailPkg.Use() + `.ParseAddress(e)
		return err == nil
	}`
	p.P(isValidEmail)

	isValidDate := `func isValidDate(f string, d string) bool {
		_, err := ` + p.timePkg.Use() + `.Parse(f, d)
		return err == nil
	}`
	p.P(isValidDate)
}

func (p *Plugin) generateErrorType() {
	ourErrorDef := `type ValidationError struct {
		Field string
		ErrorMessage string
		Errors []*ValidationError
	}

	type ValidationErrors struct {
		Errors []*ValidationError
	}

	// Error will just return the first error we encountered, inspect the actual object for more details
	func (e *ValidationErrors) Error() string {
		if e != nil && e.Errors != nil && len(e.Errors) >= 1 {
			return e.Errors[0].ErrorMessage
		}
		return ""
	}

	func GetValidationErrors(err error) ([]string, []string) {
		fields := []string{}
		errorMessages := []string{}
		if err != nil {
			if verr, ok := err.(*ValidationErrors); ok {
				errors := verr.Errors
				for i := 0; i < len(errors); i++ {
					fields = append(fields, errors[i].Field)
					errorMessages = append(errorMessages, errors[i].ErrorMessage)
					if len(errors[i].Errors) != 0 {
						errors = append(errors, errors[i].Errors...)
					}
				}
			}
		}
		return fields, errorMessages
	}
	`
	p.P(ourErrorDef)
}

func (p *Plugin) generateErrorCode(fieldName string, requiredValue string, errorMsg string, v *pb.FieldValidation, mv *pb.MessageValidation, field *descriptor.FieldDescriptorProto, subErrorArray string) {
	if v != nil && v.Error != nil {
		errorMsg = v.GetError()
	}
	errorMsg = strings.ReplaceAll(errorMsg, "{value}", requiredValue)

	if subErrorArray != "" {
		p.P(`verr := ValidationError{Errors: make([]*ValidationError, len(msgvalerr.Errors))}`)
	} else {
		p.P(`verr := ValidationError{}`)
	}

	if field.IsRepeated() {
		errorMsg = strings.ReplaceAll(errorMsg, "{field}", `" + fieldName + "`)
		p.strconvPkg.Use()
		p.P(`fieldName := "%s["+strconv.Itoa(i)+"]"`, fieldName)
		p.P(`verr.Field = fieldName`)
		p.P(`verr.ErrorMessage = "%s"`, errorMsg)
	} else {
		errorMsg = strings.ReplaceAll(errorMsg, "{field}", fieldName)
		p.P(`verr.Field = "%s"`, fieldName)
		p.P(`verr.ErrorMessage = "%s"`, errorMsg)
	}
	if subErrorArray != "" {
		p.P(`copy(verr.Errors, %s.Errors)`, subErrorArray)
	}
	p.P(`err.Errors = append(err.Errors, &verr)`)
	if mv != nil && mv.ReturnOnError != nil && mv.GetReturnOnError() {
		p.P(`return &err`)
	}
}
