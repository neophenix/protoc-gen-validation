package plugin

import (
	"fmt"
	"strings"
	"time"

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

	p.P("// Generated on: %v", time.Now().String())
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
	p.P("func (m *%s) Validate() *ValidationError {", message.GetName())
	p.P("err := ValidationError{}")

	for _, field := range message.Field {
		v := getFieldValidation(field)
		if field.IsMessage() {
			if isWKT(field.GetTypeName()) {
				if v != nil {
					p.P("if m.%s != nil {", generator.CamelCase(field.GetName()))
					p.generateValidationCode(field, v)
					p.P("}")
				}
			} else {
				p.P("if m.%s != nil { ", generator.CamelCase(field.GetName()))
				p.P("msgerr := m.%s.Validate()", generator.CamelCase(field.GetName()))
				p.P(`if len(err.Fields) != 0 {
					err.Fields = append(err.Fields, msgerr.Fields...)
					err.Errors = append(err.Errors, msgerr.Errors...)
				}`)
				p.P("}")
			}
		} else {
			if field.IsRepeated() {
				p.P("for i, _ := range m.%s {", generator.CamelCase(field.GetName()))
				p.generateValidationCode(field, v)
				p.P("}")
			} else {
				p.generateValidationCode(field, v)
			}
		}
	}
	// return any error and close Validate for this message
	p.P("if len(err.Fields) != 0 { return &err }")
	p.P("return nil")
	p.P("}")
}

// P forwards to p.gen.P after a Sprintf
func (p *Plugin) P(s string, args ...interface{}) { p.gen.P(fmt.Sprintf(s, args...)) }

func (p *Plugin) generateValidationCode(field *descriptor.FieldDescriptorProto, v *pb.FieldValidation) {
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

	if isString(field) {
		p.generateStringValidationCode(fieldName, fieldValueAccessor, v, field)
	} else if isInt(field) {
		p.generateIntValidationCode(fieldName, fieldValueAccessor, v, field)
	} else if isFloat(field) {
		p.generateFloatValidationCode(fieldName, fieldValueAccessor, v, field)
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
}

func (p *Plugin) generateErrorType() {
	ourErrorDef := `type ValidationError struct {
		Fields []string
		Errors []string
	}

	// Error will just return the first error we encountered, inspect the actual object for more details
	func (e ValidationError) Error() string {
		if len(e.Errors) >= 1 {
			return e.Errors[0]
		}
		return ""
	}
	`
	p.P(ourErrorDef)
}

func (p *Plugin) generateErrorCode(fieldName string, requiredValue string, errorMsg string, v *pb.FieldValidation, field *descriptor.FieldDescriptorProto) {
	/*
		p.P(`err.Fields = append(err.Fields, "%s")`, fieldName)
		if v.Error != nil {
			p.P(`err.Errors = append(err.Errors, "%s")`, v.GetError())
		} else {
			p.P(`err.Errors = append(err.Errors, "%s")`, errorMsg)
		}
	*/

	if v.Error != nil {
		errorMsg = v.GetError()
	}
	errorMsg = strings.ReplaceAll(errorMsg, "{value}", requiredValue)

	if field.IsRepeated() {
		errorMsg = strings.ReplaceAll(errorMsg, "{field}", `" + fieldName + "`)
		p.strconvPkg.Use()
		p.P(`fieldName := "%s["+strconv.Itoa(i)+"]"`, fieldName)
		p.P(`err.Fields = append(err.Fields, fieldName)`)
		p.P(`err.Errors = append(err.Errors, "%s")`, errorMsg)
	} else {
		errorMsg = strings.ReplaceAll(errorMsg, "{field}", fieldName)
		p.P(`err.Fields = append(err.Fields, "%s")`, fieldName)
		p.P(`err.Errors = append(err.Errors, "%s")`, errorMsg)
	}
}
