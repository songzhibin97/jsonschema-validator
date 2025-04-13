package rules

// 注册对象相关规则
func registerObjectRules(registry ValidatorRegistry) {
	// 属性验证
	registry.RegisterValidator("required", validateRequired)
	registry.RegisterValidator("properties", validateProperties)

	// 约束验证
	registry.RegisterValidator("minProperties", validateMinProperties)
	registry.RegisterValidator("maxProperties", validateMaxProperties)

	// 模式属性验证
	registry.RegisterValidator("patternProperties", validatePatternProperties)
	registry.RegisterValidator("additionalProperties", validateAdditionalProperties)

	// 依赖关系验证
	registry.RegisterValidator("dependencies", validateDependencies)
}
