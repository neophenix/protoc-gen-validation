# protoc-gen-validation

A protobuf v3 validation code generator.  Supports (some of) google's Well Known Types as well as handing back all errors
that occur during validation, not just the first one.

## Usage
```
func (s *server) MyRPC(ctx context.Context, req *pb.Request) (*pb.Response, error) {
    err = req.Validate()
    if err != nil {
        // do something with the errors
    }
    ...
}
```

## Supported Options
### Common
* error: string - override predefined error messages.  You can use {field} and {value} as macros that get replaced with the
field name and the required value.

### String
* not_empty_string: bool - make sure a string isn't ""
* matches: string - must match this value exactly, why, I don't know
* contains: string - must contain this string, simpler regex really
* regex: string - must match this regex
* min_len: int - must be at least this long
* max_len: int - must be at most this long
* eq_len: int - must be exactly this long
* is_uuid: bool - uses github.com/google/uuid to validate the value is a uuid
* is_email: bool - uses net/mail ParseAddress to validate this is an email address
* trim: bool - uses strings.Trim(value, " ") to remove whitespace before comparing value with other rules
* lc: bool - uses strings.ToLower before comparing with other rules
* uc: bool - uses strings.ToUpper before comparing with other rules

### Ints
* int_lte: int - must be <= this value
* int_gte: int - must be >= this value
* int_eq: int - must equal this value

### Float
* float_lte: double - must be <= this value
* float_gte: double - must be >= this value
* float_eq: double - must equal this value

### Message Options
* return_on_error: bool - returns when we encounter an error instead of collecting all of them

## Errors
Each Validate function returns a typical error, but underneath that error is a ValidationErrors struct.  This contains a slice 
of ValidationError pointers.  Each ValidationError has a Field that will be the name of the field that caused the error, and
an ErrorMessage that is the human readable message.  Each ValidationError can then also contain an Errors array, if this is a 
message in a message in a message and we need some structure to see where the problems were.

The errors are defined as such
```
type ValidationError struct {
    Field string
    ErrorMessage string
    Errors []*ValidationError
}

type ValidationErrors struct {
    Errors []*ValidationError
}
```

Usage example:
```
err = req.Validate()
if err != nil {
    if verr, ok := err.(*ValidationErrors); ok {
        for _, v := range verr.Errors {
	    fmt.Printf("%s\n", v.ErrorMessage)
        }
    }
}
```

## Example Protobuf Definition
This was just a copy + past from a test proto I was messing with, so there are repeated examples I'm sure
```
import "validation.proto";
import "google/protobuf/wrappers.proto";

message Inner {
	option (validation.message).return_on_error = true;
	string hello = 1 [(validation.field) = {not_empty_string: true}];
}

message InnerArEl {
	string element = 1;
}

message TestRequest {
	string foo = 1 [(validation.field) = {regex: "^[a-zA-Z]{2}$"}];
	string bar = 2;
	google.protobuf.StringValue baz = 3 [(validation.field) = {not_empty_string: true}];
	Inner inner = 4;
	string my_field = 5 [(validation.field) = {contains: "foo"}];
	string uuid = 6 [(validation.field) = {is_uuid: true}];
	string email = 7 [(validation.field) = {is_email: true}];
	google.protobuf.StringValue other_uuid = 8 [(validation.field) = {is_uuid: true}];
	repeated string array = 9 [(validation.field) = {not_empty_string: true}];
	repeated InnerArEl elements = 10;
}

message TestResponse {
	int64 baz = 1 [(validation.field) = {int_lte: 10, int_gte: 5}];
	float qux = 2 [(validation.field) = {float_eq: 3.14}];
}

message StringTests {
	string min = 1 [(validation.field) = {min_len: 10}];
	string max = 2 [(validation.field) = {max_len: 10}];
	string eq = 3 [(validation.field) = {eq_len: 10}];
}
```

## But Why?
This was heavily inspired / cloned in places from https://github.com/mwitkow/go-proto-validators 
and https://github.com/envoyproxy/protoc-gen-validate I could not find good documentation on how to build a validator so 
much of the time was spent going through those projects, seeing how they did things and trying to mimic it.  I'm sure there
is some code lifted from either where I just gave up and pasted something in and it worked.

### So why not just use one of them?
I ran into some showstoppers in both that I felt I couldn't easily overcome.  Mostly I didn't want to have to modify the
generated code that much to deal with the issues.  mwitkow's does not, to my knowledge, support the well known types.  Working
around that lead to ugly translations from protobuf <-> JSON.  Lyft's is still alpha as documented in their repo, and I've 
found for my use cases it generates invalid code in at least 2 instances (both have open issues already).

Despite their shortcomings for my cases, the other repos support more features and my assumption is that eventually when they
resolve the issues, I'll just switch back to using one of them as I'm sure they will be better maintained.
