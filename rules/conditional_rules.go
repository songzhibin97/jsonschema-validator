package rules

import (
	"context"
	"fmt"

	"github.com/songzhibin97/jsonschema-validator/errors"
)

func registerConditionalRules(registry ValidatorRegistry) {
	registry.RegisterValidator("if", validateIf)
	registry.RegisterValidator("then", validateThen)
	registry.RegisterValidator("else", validateElse)
	registry.RegisterValidator("conditional", func(ctx context.Context, value interface{}, schemaValue interface{}, path string) (bool, error) {
		schema, ok := schemaValue.(map[string]interface{})
		if !ok {
			return false, &errors.ValidationError{
				Path:    path,
				Message: "conditional must be an object",
				Value:   schemaValue,
				Tag:     "conditional",
			}
		}
		return ValidateConditional(ctx, value, schema, path)
	})
}

func validateIf(ctx context.Context, value interface{}, schemaValue interface{}, path string) (bool, error) {
	schema, ok := schemaValue.(map[string]interface{})
	if !ok {
		return false, &errors.ValidationError{Path: path, Message: "if must be an object", Value: schemaValue, Tag: "if"}
	}
	registry, ok := ctx.Value("validator").(ValidatorRegistry)
	if !ok {
		return false, &errors.ValidationError{Path: path, Message: "validator not found in context", Tag: "if"}
	}
	isValid := true
	for keyword, keywordValue := range schema {
		if keyword == "title" || keyword == "description" || keyword == "default" || keyword == "examples" {
			continue
		}
		validator := registry.GetValidator(keyword)
		if validator == nil {
			continue
		}
		valid, err := validator(ctx, value, keywordValue, path)
		if err != nil || !valid {
			isValid = false
			break
		}
	}
	ctx = context.WithValue(ctx, "ifConditionMet", isValid)
	return true, nil
}

func validateThen(ctx context.Context, value interface{}, schemaValue interface{}, path string) (bool, error) {
	// 先检查 validator 是否存在
	registry, ok := ctx.Value("validator").(ValidatorRegistry)
	if !ok {
		return false, &errors.ValidationError{Path: path, Message: "validator not found in context", Tag: "then"}
	}

	// 再检查 ifConditionMet
	conditionMet, ok := ctx.Value("ifConditionMet").(bool)
	if !ok || !conditionMet {
		return true, nil
	}

	schema, ok := schemaValue.(map[string]interface{})
	if !ok {
		return false, &errors.ValidationError{Path: path, Message: "then must be an object", Value: schemaValue, Tag: "then"}
	}

	for keyword, keywordValue := range schema {
		if keyword == "title" || keyword == "description" || keyword == "default" || keyword == "examples" {
			continue
		}
		validator := registry.GetValidator(keyword)
		if validator == nil {
			continue
		}
		valid, err := validator(ctx, value, keywordValue, path)
		if err != nil || !valid {
			return false, &errors.ValidationError{
				Path:    path,
				Message: fmt.Sprintf("validation failed against then schema for keyword '%s'", keyword),
				Value:   value,
				Tag:     keyword,
			}
		}
	}
	return true, nil
}

func validateElse(ctx context.Context, value interface{}, schemaValue interface{}, path string) (bool, error) {
	// 先检查 validator 是否存在
	registry, ok := ctx.Value("validator").(ValidatorRegistry)
	if !ok {
		return false, &errors.ValidationError{Path: path, Message: "validator not found in context", Tag: "else"}
	}

	// 再检查 ifConditionMet
	conditionMet, ok := ctx.Value("ifConditionMet").(bool)
	if !ok || conditionMet {
		return true, nil
	}

	schema, ok := schemaValue.(map[string]interface{})
	if !ok {
		return false, &errors.ValidationError{Path: path, Message: "else must be an object", Value: schemaValue, Tag: "else"}
	}

	for keyword, keywordValue := range schema {
		if keyword == "title" || keyword == "description" || keyword == "default" || keyword == "examples" {
			continue
		}
		validator := registry.GetValidator(keyword)
		if validator == nil {
			continue
		}
		valid, err := validator(ctx, value, keywordValue, path)
		if err != nil || !valid {
			return false, &errors.ValidationError{
				Path:    path,
				Message: fmt.Sprintf("validation failed against else schema for keyword '%s'", keyword),
				Value:   value,
				Tag:     keyword,
			}
		}
	}
	return true, nil
}

func ValidateConditional(ctx context.Context, value interface{}, conditionalSchema map[string]interface{}, path string) (bool, error) {
	ifSchema, hasIf := conditionalSchema["if"]
	thenSchema, hasThen := conditionalSchema["then"]
	elseSchema, hasElse := conditionalSchema["else"]

	updatedCtx := ctx

	if hasIf {
		valid, err := validateIf(updatedCtx, value, ifSchema, path+".if")
		if !valid || err != nil {
			return false, err
		}
		isValid := true
		schema, ok := ifSchema.(map[string]interface{})
		if ok {
			registry, ok := updatedCtx.Value("validator").(ValidatorRegistry)
			if ok {
				for keyword, keywordValue := range schema {
					if keyword == "title" || keyword == "description" || keyword == "default" || keyword == "examples" {
						continue
					}
					validator := registry.GetValidator(keyword)
					if validator != nil {
						valid, err := validator(updatedCtx, value, keywordValue, path+".if")
						if err != nil || !valid {
							isValid = false
							break
						}
					}
				}
			}
		}
		updatedCtx = context.WithValue(updatedCtx, "ifConditionMet", isValid)
	}

	if hasThen {
		valid, err := validateThen(updatedCtx, value, thenSchema, path+".then")
		if !valid || err != nil {
			return false, err
		}
	}

	if hasElse {
		valid, err := validateElse(updatedCtx, value, elseSchema, path+".else")
		if !valid || err != nil {
			return false, err
		}
	}

	return true, nil
}
