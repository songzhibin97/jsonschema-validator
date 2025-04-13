# JSONSchema Validator

`jsonschema-validator` is a Go library for validating JSON data and Go structs against JSON Schema specifications. It supports both strict and loose validation modes, custom validation rules, recursive struct validation, and caching for performance optimization. The library is designed to be flexible, extensible, and easy to integrate into Go applications.

## Features

- **JSON Schema Validation**: Validate JSON data against JSON Schema definitions.
- **Struct Validation**: Validate Go structs using struct tags or custom validation functions.
- **Custom Validation**: Add custom validation logic for specific fields or schemas.
- **Validation Modes**: Supports `Strict`, `Loose`, and `Warn` modes for flexible validation.
- **Caching**: Cache compiled schemas for improved performance.
- **Recursive Validation**: Validate nested structs automatically.
- **Concurrent Safety**: Safe for use in concurrent applications.
- **Detailed Error Reporting**: Provides comprehensive error messages with paths and tags.

## Installation

To use `jsonschema-validator` in your Go project, ensure you have Go 1.16 or later installed. Then, install the library using:

```bash
go get github.com/songzhibin97/jsonschema-validator
```

Add the following import to your Go code:

```go
import "github.com/songzhibin97/jsonschema-validator/validator"
```

## Usage

The library provides several ways to validate data:

- Validate JSON Against a Schema
- Validate Go Structs with Tags
- Validate with Custom Schema Maps
- Custom Validation Functions
- Concurrent Validation

Below are detailed examples for each use case.

### Example 1: Validate JSON Against a Schema

Validate JSON data against a JSON Schema string.

```go
package main

import (
    "fmt"
    "github.com/songzhibin97/jsonschema-validator/validator"
    "github.com/songzhibin97/jsonschema-validator/schema"
)

func main() {
    v := validator.New(validator.WithValidationMode(schema.ModeStrict))

    jsonData := `{"name":"John","age":30}`
    schemaJSON := `{
        "type":"object",
        "properties":{
            "name":{"type":"string"},
            "age":{"type":"integer","minimum":18}
        },
        "required":["name"],
        "additionalProperties":false
    }`

    result, err := v.ValidateJSON(jsonData, schemaJSON)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }

    if result.Valid {
        fmt.Println("JSON is valid!")
    } else {
        fmt.Println("JSON is invalid:")
        for _, e := range result.Errors {
            fmt.Printf("- %s: %s\n", e.Path, e.Message)
        }
    }
}
```

Output:
```
JSON is valid!
```

Invalid Example:
```go
jsonData := `{"name":"John","age":16,"extra":true}`
```

Output:
```
JSON is invalid:
- root.age: value 16 is less than minimum 18
- root.extra: unknown field
```

### Example 2: Validate Go Structs with Tags

Validate a Go struct using struct tags to define validation rules.

```go
package main

import (
    "fmt"
    "github.com/songzhibin97/jsonschema-validator/validator"
)

type User struct {
    Name string `validate:"required,type=string"`
    Age  int    `validate:"minimum=18"`
    Role string `validate:"enum=admin|user"`
}

func main() {
    v := validator.New(validator.WithTagName("validate"))

    user := User{
        Name: "John",
        Age:  30,
        Role: "admin",
    }

    err := v.Struct(user)
    if err != nil {
        fmt.Printf("Validation failed: %v\n", err)
        return
    }

    fmt.Println("Struct is valid!")
}
```

Output:
```
Struct is valid!
```

Invalid Example:
```go
user := User{
    Name: "",
    Age:  15,
    Role: "guest",
}
```

Output:
```
Validation failed: validation failed with the following errors:
[1] validation error: field is required (path: Name)
[2] validation error: value 15 is less than minimum 18 (path: Age)
[3] validation error: value must be one of: admin, user (path: Role)
```

### Example 3: Validate with Custom Schema Maps

Validate data against a programmatically defined schema map.

```go
package main

import (
    "fmt"
    "github.com/songzhibin97/jsonschema-validator/validator"
    "github.com/songzhibin97/jsonschema-validator/schema"
)

func main() {
    v := validator.New(validator.WithValidationMode(schema.ModeLoose))

    schemaMap := map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "name": map[string]interface{}{"type": "string"},
            "age":  map[string]interface{}{"type": "integer", "minimum": 18},
        },
        "required": []interface{}{"name"},
    }

    data := map[string]interface{}{
        "name": "John",
        "age":  30,
    }

    result, err := v.ValidateWithSchema(data, schemaMap, "root")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }

    if result.Valid {
        fmt.Println("Data is valid!")
    } else {
        fmt.Println("Data is invalid:")
        for _, e := range result.Errors {
            fmt.Printf("- %s: %s\n", e.Path, e.Message)
        }
    }
}
```

Output:
```
Data is valid!
```

Invalid Example:
```go
data := map[string]interface{}{
    "name": 123,
    "age":  15,
}
```

Output:
```
Data is invalid:
- root.name: value is of type int, expected string
- root.age: value 15 is less than minimum 18
```

### Example 4: Custom Validation Functions

Add custom validation logic for specific fields, such as checking if a string starts with a prefix.

```go
package main

import (
    "context"
    "fmt"
    "strings"
    "github.com/songzhibin97/jsonschema-validator/validator"
)

type User struct {
    Role string `validate:"required"`
}

func main() {
    v := validator.New()
    v.SetCustomValidateFunc(func(ctx context.Context, value interface{}, path string) (bool, error) {
        if str, ok := value.(string); ok && strings.HasPrefix(strings.ToUpper(str), "ADMIN_") {
            return true, nil
        }
        return false, nil
    })

    user := User{Role: "ADMIN_user"}
    err := v.Struct(user)
    if err != nil {
        fmt.Printf("Validation failed: %v\n", err)
        return
    }

    fmt.Println("User role is valid!")

    // Invalid case
    user = User{Role: "user"}
    err = v.Struct(user)
    if err != nil {
        fmt.Printf("Validation failed: %v\n", err)
    }
}
```

Output:
```
User role is valid!
Validation failed: validation failed with the following errors:
[1] validation error: value must start with 'ADMIN_' (path: Role)
```

### Example 5: Concurrent Validation

Validate multiple structs concurrently, leveraging the library's thread-safe design.

```go
package main

import (
    "fmt"
    "sync"
    "github.com/songzhibin97/jsonschema-validator/validator"
)

type User struct {
    Name string `validate:"required,type=string"`
}

func main() {
    v := validator.New()
    var wg sync.WaitGroup

    for i := 0; i < 5; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            user := User{Name: fmt.Sprintf("User%d", id)}
            err := v.Struct(user)
            if err != nil {
                fmt.Printf("Validation failed for User%d: %v\n", id, err)
            } else {
                fmt.Printf("User%d is valid\n", id)
            }
        }(i)
    }

    wg.Wait()
}
```

Output:
```
User0 is valid
User1 is valid
User2 is valid
User3 is valid
User4 is valid
```

### Example 6: Nested Struct Validation

Validate nested structs with recursive validation enabled.

```go
package main

import (
    "fmt"
    "github.com/songzhibin97/jsonschema-validator/validator"
)

type Address struct {
    City string `validate:"required"`
}

type User struct {
    Name    string  `validate:"required"`
    Age     int     `validate:"minimum=18"`
    Address Address `validate:"required"`
}

func main() {
    v := validator.New(validator.WithRecursiveValidation(true))

    user := User{
        Name: "John",
        Age:  30,
        Address: Address{
            City: "New York",
        },
    }

    err := v.Struct(user)
    if err != nil {
        fmt.Printf("Validation failed: %v\n", err)
        return
    }

    fmt.Println("Nested struct is valid!")
}
```

Output:
```
Nested struct is valid!
```

Invalid Example:
```go
user := User{
    Name: "John",
    Age:  15,
    Address: Address{
        City: "",
    },
}
```

Output:
```
Validation failed: validation failed with the following errors:
[1] validation error: value 15 is less than minimum 18 (path: Age)
[2] validation error: field is required (path: Address.City)
```

### Example 7: Schema Compilation and Caching

Compile and cache schemas for repeated validation to improve performance.

```go
package main

import (
    "fmt"
    "github.com/songzhibin97/jsonschema-validator/validator"
)

func main() {
    v := validator.New(validator.WithCaching(true))

    schemaJSON := `{
        "type":"object",
        "properties":{"name":{"type":"string"}},
        "required":["name"]
    }`

    // Compile schema once
    _, err := v.CompileSchema(schemaJSON)
    if err != nil {
        fmt.Printf("Error compiling schema: %v\n", err)
        return
    }

    // Validate multiple times (uses cached schema)
    jsonData := `{"name":"John"}`
    result, err := v.ValidateJSON(jsonData, schemaJSON)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }

    if result.Valid {
        fmt.Println("JSON is valid (cached schema)!")
    }

    // Clear cache if needed
    v.ClearCache()
}
```

Output:
```
JSON is valid (cached schema)!
```

## Configuration Options

Customize the validator with the following options:

- `WithTagName (string)`: Set the struct tag name for validation rules (default: `"validate"`).
- `WithValidationMode (schema.ValidationMode)`: Set validation mode (`ModeStrict`, `ModeLoose`, `ModeWarn`).
- `WithErrorFormattingMode (errors.FormattingMode)`: Set error formatting (`FormattingModeDetailed`, `FormattingModeSimple`).
- `WithCaching (bool)`: Enable schema caching for performance (default: `false`).
- `WithStopOnFirstError (bool)`: Stop validation on the first error (default: `false`).
- `WithRecursiveValidation (bool)`: Enable recursive validation for nested structs (default: `false`).
- `WithAllowUnknownFields (bool)`: Allow unknown fields in JSON objects (default: `false`).

Example:
```go
v := validator.New(
    validator.WithTagName("custom"),
    validator.WithValidationMode(schema.ModeLoose),
    validator.WithStopOnFirstError(true),
    validator.WithRecursiveValidation(true),
)
```

## Error Handling

Validation errors are returned as `errors.ValidationErrors`, containing a slice of `errors.ValidationError` with details:

- `Path`: The path to the invalid field (e.g., `"root.name"`).
- `Message`: The error message (e.g., `"expected string"`).
- `Tag`: The validation rule that failed (e.g., `"type"`).
- `Value`: The invalid value (optional).

Example:
```go
if err != nil {
    if ve, ok := err.(errors.ValidationErrors); ok {
        for _, e := range ve {
            fmt.Printf("Error at %s: %s (tag: %s)\n", e.Path, e.Message, e.Tag)
        }
    }
}
```

## Supported Validation Keywords

The library supports standard JSON Schema keywords, including but not limited to:

- `type` (e.g., `"string"`, `"integer"`, `"object"`, `"array"`)
- `required` (array of required property names)
- `minimum` / `maximum` (for numbers)
- `minLength` / `maxLength` (for strings)
- `enum` (array of allowed values)
- `properties` (object properties)
- `items` (array items)
- `additionalProperties` (control unknown fields)

Custom keywords can be registered using `RegisterValidator`.

## Contributing

Contributions are welcome! To contribute:

1. Fork the repository.
2. Create a feature branch (`git checkout -b feature/my-feature`).
3. Commit your changes (`git commit -m "Add my feature"`).
4. Push to the branch (`git push origin feature/my-feature`).
5. Open a pull request.

Please include tests for new features and ensure existing tests pass.

## License

This project is licensed under the MIT License. See the LICENSE file for details.