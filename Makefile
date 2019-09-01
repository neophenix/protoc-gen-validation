build:
	go build -o protoc-gen-validation cmd/protoc-gen-validation/main.go
	mv protoc-gen-validation ~/go/bin/

validation:
	protoc \
		-I . \
		--gogofast_out=Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor:. \
		validation.proto
