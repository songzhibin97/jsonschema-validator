package rules

import (
	"context"
	"fmt"
	"regexp"

	"github.com/songzhibin97/jsonschema-validator/errors"
)

func validatePatternProperties(ctx context.Context, value interface{}, schemaValue interface{}, path string) (bool, error) {
	patternProps, ok := schemaValue.(map[string]interface{})
	if !ok {
		return false, &errors.ValidationError{Path: path, Message: "patternProperties must be an object", Value: schemaValue, Tag: "patternProperties"}
	}
	obj, ok := value.(map[string]interface{})
	if !ok {
		return false, &errors.ValidationError{Path: path, Message: "patternProperties can only be applied to objects", Value: value, Tag: "patternProperties"}
	}
	registry, ok := ctx.Value("validator").(ValidatorRegistry)
	if !ok {
		return false, &errors.ValidationError{Path: path, Message: "validator not found in context", Tag: "patternProperties"}
	}
	ctx = context.WithValue(ctx, "patternProperties", patternProps)
	patterns := make(map[string]*regexp.Regexp)
	for pattern := range patternProps {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return false, &errors.ValidationError{Path: path, Message: fmt.Sprintf("invalid pattern: %s", err.Error()), Value: pattern, Tag: "patternProperties"}
		}
		patterns[pattern] = re
	}
	for propName, propValue := range obj {
		for pattern, re := range patterns {
			if re.MatchString(propName) {
				propSchema, ok := patternProps[pattern]
				if !ok {
					continue
				}
				propSchemaObj, ok := propSchema.(map[string]interface{})
				if !ok {
					continue
				}
				propPath := fmt.Sprintf("%s.%s", path, propName)
				for keyword, keywordValue := range propSchemaObj {
					if keyword == "title" || keyword == "description" || keyword == "default" || keyword == "examples" {
						continue
					}
					validator := registry.GetValidator(keyword)
					if validator == nil {
						continue
					}
					isValid, err := validator(ctx, propValue, keywordValue, propPath)
					if !isValid || err != nil {
						return false, &errors.ValidationError{
							Path:    propPath,
							Message: fmt.Sprintf("property validation failed for keyword '%s'", keyword),
							Value:   propValue,
							Tag:     keyword,
						}
					}
				}
			}
		}
	}
	return true, nil
}

func validateAdditionalProperties(ctx context.Context, value interface{}, schemaValue interface{}, path string) (bool, error) {
	obj, ok := value.(map[string]interface{})
	if !ok {
		return false, &errors.ValidationError{Path: path, Message: "additionalProperties can only be applied to objects", Value: value, Tag: "additionalProperties"}
	}
	properties, _ := ctx.Value("properties").(map[string]interface{})
	patternProperties, _ := ctx.Value("patternProperties").(map[string]interface{})
	patterns := make([]*regexp.Regexp, 0)
	for pattern := range patternProperties {
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}
		patterns = append(patterns, re)
	}
	additionalProps := make(map[string]interface{})
	for propName, propValue := range obj {
		if _, exists := properties[propName]; exists {
			continue
		}
		matched := false
		for _, re := range patterns {
			if re.MatchString(propName) {
				matched = true
				break
			}
		}
		if matched {
			continue
		}
		additionalProps[propName] = propValue
	}
	if len(additionalProps) == 0 {
		return true, nil
	}
	switch schemaValue.(type) {
	case bool:
		allowed, _ := schemaValue.(bool)
		if !allowed {
			return false, &errors.ValidationError{Path: path, Message: "additional properties are not allowed", Value: additionalProps, Tag: "additionalProperties"}
		}
		return true, nil
	case map[string]interface{}:
		schema, _ := schemaValue.(map[string]interface{})
		registry, ok := ctx.Value("validator").(ValidatorRegistry)
		if !ok {
			return false, &errors.ValidationError{Path: path, Message: "validator not found in context", Tag: "additionalProperties"}
		}
		for propName, propValue := range additionalProps {
			propPath := fmt.Sprintf("%s.%s", path, propName)
			for keyword, keywordValue := range schema {
				if keyword == "title" || keyword == "description" || keyword == "default" || keyword == "examples" {
					continue
				}
				validator := registry.GetValidator(keyword)
				if validator == nil {
					continue
				}
				isValid, err := validator(ctx, propValue, keywordValue, propPath)
				if !isValid || err != nil {
					return false, &errors.ValidationError{
						Path:    propPath,
						Message: fmt.Sprintf("additional property validation failed for keyword '%s'", keyword),
						Value:   propValue,
						Tag:     keyword,
					}
				}
			}
		}
		return true, nil
	default:
		return false, &errors.ValidationError{Path: path, Message: "additionalProperties must be a boolean or an object", Value: schemaValue, Tag: "additionalProperties"}
	}
}
