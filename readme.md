# JSONSchema 验证器

`jsonschema-validator` 是一个用于验证 JSON 数据和 Go 结构体的 Go 库，支持 JSON Schema 规范。它提供严格和宽松验证模式、自定义验证规则、递归结构体验证，以及用于性能优化的缓存功能。该库设计灵活、可扩展，并且易于集成到 Go 应用程序中。

## 特性

- **JSON Schema 验证**：根据 JSON Schema 定义验证 JSON 数据。
- **结构体验证**：使用结构体标签或自定义验证函数验证 Go 结构体。
- **自定义验证**：为特定字段或模式添加自定义验证逻辑。
- **验证模式**：支持 `严格`、`宽松` 和 `警告` 模式，实现灵活验证。
- **缓存**：缓存已编译的模式以提高性能。
- **递归验证**：自动验证嵌套结构体。
- **并发安全**：可安全用于并发应用程序。
- **详细错误报告**：提供包含路径和标签的全面错误信息。

## 安装

要在 Go 项目中使用 `jsonschema-validator`，确保已安装 Go 1.16 或更高版本。然后，使用以下命令安装库：

```bash
go get github.com/songzhibin97/jsonschema-validator
```

在您的 Go 代码中添加以下导入：

```go
import "github.com/songzhibin97/jsonschema-validator/validator"
```

## 使用方法

该库提供了几种验证数据的方法：

- 根据模式验证 JSON
- 使用标签验证 Go 结构体
- 使用自定义模式映射进行验证
- 自定义验证函数
- 并发验证

以下是每个用例的详细示例。

### 示例 1：根据模式验证 JSON

根据 JSON Schema 字符串验证 JSON 数据。

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
        fmt.Printf("错误: %v\n", err)
        return
    }

    if result.Valid {
        fmt.Println("JSON 有效！")
    } else {
        fmt.Println("JSON 无效:")
        for _, e := range result.Errors {
            fmt.Printf("- %s: %s\n", e.Path, e.Message)
        }
    }
}
```

输出：
```
JSON 有效！
```

无效示例：
```go
jsonData := `{"name":"John","age":16,"extra":true}`
```

输出：
```
JSON 无效:
- root.age: 值 16 小于最小值 18
- root.extra: 未知字段
```

### 示例 2：使用标签验证 Go 结构体

使用结构体标签定义验证规则，验证 Go 结构体。

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
        fmt.Printf("验证失败: %v\n", err)
        return
    }

    fmt.Println("结构体有效！")
}
```

输出：
```
结构体有效！
```

无效示例：
```go
user := User{
    Name: "",
    Age:  15,
    Role: "guest",
}
```

输出：
```
验证失败: 验证失败，存在以下错误:
[1] 验证错误: 字段是必需的 (路径: Name)
[2] 验证错误: 值 15 小于最小值 18 (路径: Age)
[3] 验证错误: 值必须是以下之一: admin, user (路径: Role)
```

### 示例 3：使用自定义模式映射进行验证

根据以编程方式定义的模式映射验证数据。

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
        fmt.Printf("错误: %v\n", err)
        return
    }

    if result.Valid {
        fmt.Println("数据有效！")
    } else {
        fmt.Println("数据无效:")
        for _, e := range result.Errors {
            fmt.Printf("- %s: %s\n", e.Path, e.Message)
        }
    }
}
```

输出：
```
数据有效！
```

无效示例：
```go
data := map[string]interface{}{
    "name": 123,
    "age":  15,
}
```

输出：
```
数据无效:
- root.name: 值类型为 int，预期为 string
- root.age: 值 15 小于最小值 18
```

### 示例 4：自定义验证函数

为特定字段添加自定义验证逻辑，例如检查字符串是否以前缀开头。

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
        fmt.Printf("验证失败: %v\n", err)
        return
    }

    fmt.Println("用户角色有效！")

    // 无效示例
    user = User{Role: "user"}
    err = v.Struct(user)
    if err != nil {
        fmt.Printf("验证失败: %v\n", err)
    }
}
```

输出：
```
用户角色有效！
验证失败: 验证失败，存在以下错误:
[1] 验证错误: 值必须以 'ADMIN_' 开头 (路径: Role)
```

### 示例 5：并发验证

并发验证多个结构体，利用库的线程安全设计。

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
                fmt.Printf("User%d 验证失败: %v\n", id, err)
            } else {
                fmt.Printf("User%d 有效\n", id)
            }
        }(i)
    }

    wg.Wait()
}
```

输出：
```
User0 有效
User1 有效
User2 有效
User3 有效
User4 有效
```

### 示例 6：嵌套结构体验证

启用递归验证以验证嵌套结构体。

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
        fmt.Printf("验证失败: %v\n", err)
        return
    }

    fmt.Println("嵌套结构体有效！")
}
```

输出：
```
嵌套结构体有效！
```

无效示例：
```go
user := User{
    Name: "John",
    Age:  15,
    Address: Address{
        City: "",
    },
}
```

输出：
```
验证失败: 验证失败，存在以下错误:
[1] 验证错误: 值 15 小于最小值 18 (路径: Age)
[2] 验证错误: 字段是必需的 (路径: Address.City)
```

### 示例 7：模式编译与缓存

编译并缓存模式以进行重复验证，提高性能。

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

    // 编译模式一次
    _, err := v.CompileSchema(schemaJSON)
    if err != nil {
        fmt.Printf("编译模式错误: %v\n", err)
        return
    }

    // 多次验证（使用缓存模式）
    jsonData := `{"name":"John"}`
    result, err := v.ValidateJSON(jsonData, schemaJSON)
    if err != nil {
        fmt.Printf("错误: %v\n", err)
        return
    }

    if result.Valid {
        fmt.Println("JSON 有效（缓存模式）！")
    }

    // 需要时清除缓存
    v.ClearCache()
}
```

输出：
```
JSON 有效（缓存模式）！
```

## 配置选项

使用以下选项自定义验证器：

- `WithTagName (string)`：设置用于验证规则的结构体标签名称（默认：`"validate"`）。
- `WithValidationMode (schema.ValidationMode)`：设置验证模式（`ModeStrict`、`ModeLoose`、`ModeWarn`）。
- `WithErrorFormattingMode (errors.FormattingMode)`：设置错误格式化模式（`FormattingModeDetailed`、`FormattingModeSimple`）。
- `WithCaching (bool)`：启用模式缓存以提高性能（默认：`false`）。
- `WithStopOnFirstError (bool)`：在第一个错误处停止验证（默认：`false`）。
- `WithRecursiveValidation (bool)`：为嵌套结构体启用递归验证（默认：`false`）。
- `WithAllowUnknownFields (bool)`：允许 JSON 对象中的未知字段（默认：`false`）。

示例：
```go
v := validator.New(
    validator.WithTagName("custom"),
    validator.WithValidationMode(schema.ModeLoose),
    validator.WithStopOnFirstError(true),
    validator.WithRecursiveValidation(true),
)
```

## 错误处理

验证错误以 `errors.ValidationErrors` 返回，包含一个 `errors.ValidationError` 切片，其中包含详细信息：

- `Path`：无效字段的路径（例如，`"root.name"`）。
- `Message`：错误消息（例如，`"预期为字符串"`）。
- `Tag`：失败的验证规则（例如，`"type"`）。
- `Value`：无效值（可选）。

示例：
```go
if err != nil {
    if ve, ok := err.(errors.ValidationErrors); ok {
        for _, e := range ve {
            fmt.Printf("错误位于 %s: %s (标签: %s)\n", e.Path, e.Message, e.Tag)
        }
    }
}
```

## 支持的验证关键字

该库支持标准 JSON Schema 关键字，包括但不限于：

- `type`（例如，`"string"`、`"integer"`、`"object"`、`"array"`）
- `required`（必需属性名称的数组）
- `minimum` / `maximum`（用于数字）
- `minLength` / `maxLength`（用于字符串）
- `enum`（允许值的数组）
- `properties`（对象属性）
- `items`（数组项）
- `additionalProperties`（控制未知字段）

可以使用 `RegisterValidator` 注册自定义关键字。

## 贡献

欢迎贡献！要贡献：

1. Fork 仓库。
2. 创建特性分支（`git checkout -b feature/my-feature`）。
3. 提交更改（`git commit -m "Add my feature"`）。
4. 推送到分支（`git push origin feature/my-feature`）。
5. 开启拉取请求。

请为新特性包含测试，并确保现有测试通过。

## 许可证

该项目根据 MIT 许可证授权。有关详细信息，请参阅许可证文件。