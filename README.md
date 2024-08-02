# encoding/json with "optional" and "nullable" struct field tags
This is a copy of Go's encoding/json with modifications to add "optional" and "nullable" struct field tags.

## Intention
Differentiate between JSON fields that are undefined (meaning not present), null, or are Go zero-values.

Currently, Go's `encoding/json` package lumps all three cases into the Go zero-value, which makes it impossible to
differentiate between them without heavy, clunky workarounds.

Use cases:
- [JSON Schema](https://json-schema.org/specification-links#2020-12), [OpenAPI](https://swagger.io/specification/), etc
  types/code generation
- [JSON Merge Patch](https://datatracker.ietf.org/doc/html/rfc7396)
- anything else where zero-valued, null, and undefined/omitted fields can have different meanings

## Usage
#### Example
```go
type MyStruct struct {
	BasicInt            int   `json:""`
	BasicIntPtr         *int  `json:""`
	NullableInt         *int  `json:",nullable"`
	OptionalInt         *int  `json:",optional"`
	OptionalNullableInt **int `json:",optional,nullable"`  // NOTE: the order of optional and nullable does not matter
}
```

#### Marshalling
In the above example:
- `BasicInt` and `BasicIntPtr` are handled as vanilla `encoding/json` would handle them.
- `NullableInt` will be set to `null` if the field is `nil`.
- `NullableInt` will be set to `0` if the field is `&(int(0))`.
- `OptionalInt` will be omitted from the JSON if it is `nil`.
- `OptionalInt` will be set to `0` if the field is `&(int(0))`.
- `OptionalNullableInt` will be omitted from the JSON if it is `nil`.
- `OptionalNullableInt` will be set to `null` if it is `&((*int)(nil))`.
- `OptionalNullableInt` will be set to `0` if the field is `&(&(int(0)))`.

#### Unmarshalling

In the above example:
- `BasicInt` and `BasicIntPtr` are handled as vanilla `encoding/json` would handle them.
- GOTCHA: `NullableInt` will return an error if the field is undefined.
  - This is for consistency: a `nil` `NullableInt` means the field was explicitly set to `null`.
  - Elsewhere, I refer to this as a nullable-but-not-optional field.
- `NullableInt` will be set to `nil` if the field is `null`.
- `NullableInt` will be set to `&(int(0))` if the field is `0`.
- `OptionalInt` will be set to `nil` if the field is undefined.
- `OptionalInt` will be set to `&(int(0))` if the field is `0`.
- `OptionalNullableInt` will be set to `nil` if the field is undefined.
- `OptionalNullableInt` will be set to `&((*int)(nil))` if the field is `null`.
- `OptionalNullableInt` will be set to `&(&(int(0)))` if the field is `0`.

You can differentiate between a field that is undefined/omitted, a field that is `null`, and a zero-valued field as follows:
```go
x := MyStruct{}
json.Unmarshal(data, &x)  // for example's sake, assume no error

// optional field
isUndefined := x.OptionalInt == nil
isZero := *x.OptionalInt == 0  // assuming x.OptionalInt is not nil

// nullable field
isNull := x.NullableInt == nil
isZero = *x.NullableInt == 0  // assuming x.NullableInt is not nil

// optional, nullable field
isUndefined = x.OptionalNullableInt == nil
isNull = *x.OptionalNullableInt == nil  // assuming x.OptionalNullableInt is not nil
isZero = **x.OptionalNullableInt == 0   // assuming x.OptionalNullableInt and *x.OptionalNullableInt are not nil
```

## Gotchas
- The `optional` and `nullable` tags are not compatible with the `omitempty` tag and will return an error at
  marshal/unmarshal time if used together.
- `optional` and `nullable` tags each require an additional level of indirection for the field.
  - For example, for a base type `T`:
    - ``*T `json:",nullable"` ``
    - ``*T `json:",optional"` ``
    - ``**T `json:",optional,nullable"` ``
    - this includes when `T` is a pointer type: ``***APtrType `json:",optional,nullable"` ``
  - An insufficient level of indirection will return an error at marshal/unmarshal time.
- As mentioned in the Usage > Unmarshalling section, nullable-but-not-optional fields will raise an error if the field
  is undefined.
- To be clear, the absence of an `optional` tag does not imply that a field is required (expect for the
  nullable-but-not-optional case that was just mentioned). It only means we will not perform any special `optional`
  handling for the field.

## Future ideas
- Instead of (or in addition to) using pointers for `optional` and `nullable` fields, use custom types.
  - Perhaps `json.Nullable[T any]`, `json.Optional[T any]`, and `json.OptionalNullable[T any]` types that expose
    `IsSet()` and `IsNull()` methods.
  - Pro: less pointer dereferencing.
  - Con: it more strongly couples struct types with this JSON package. Keeping struct definitions and
    JSON logic separate is a good thing. (But we're already adding pointers to accommodate JSON, so...?)
  - Con: it is more verbose.