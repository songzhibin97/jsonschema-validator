package rules

import (
	"encoding/json"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// 数值转换函数

// toFloat64 尝试将值转换为float64
func toFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case json.Number:
		f, err := v.Float64()
		if err != nil {
			return 0, false
		}
		return f, true
	case string:
		var f float64
		_, err := fmt.Sscanf(v, "%f", &f)
		if err != nil {
			return 0, false
		}
		return f, true
	default:
		return 0, false
	}
}

// toInt 尝试将值转换为int
func toInt(value interface{}) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int8:
		return int(v), true
	case int16:
		return int(v), true
	case int32:
		return int(v), true
	case int64:
		return int(v), true
	case uint:
		return int(v), true
	case uint8:
		return int(v), true
	case uint16:
		return int(v), true
	case uint32:
		return int(v), true
	case uint64:
		if v <= uint64(^uint(0)>>1) {
			return int(v), true
		}
	case float32:
		if float32(int(v)) == v {
			return int(v), true
		}
	case float64:
		if float64(int(v)) == v {
			return int(v), true
		}
	case string:
		if i, err := strconv.Atoi(v); err == nil {
			return i, true
		}
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return int(i), true
		}
	}
	return 0, false
}

type Stringer interface {
	String() string
}

// toString 尝试将值转换为字符串
func toString(value interface{}) (string, bool) {
	switch v := value.(type) {
	case string:
		return v, true
	case []byte:
		return string(v), true
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v), true
	case float32, float64:
		return fmt.Sprintf("%v", v), true
	case Stringer:
		return v.String(), true
	case error:
		return v.Error(), true
	default:
		return "", false
	}
}

// toBool 尝试将值转换为布尔值
func toBool(value interface{}) (bool, bool) {
	switch v := value.(type) {
	case bool:
		return v, true
	case string:
		return v == "true" || v == "1" || v == "yes" || v == "y", true
	case int:
		return v != 0, true
	case float64:
		return v != 0, true
	default:
		return false, false
	}
}

// 格式验证函数

// validateEmail 验证邮箱格式
func validateEmail(str string) bool {
	_, err := mail.ParseAddress(str)
	return err == nil
}

// validateDateTime 验证日期时间格式（RFC3339）
func validateDateTime(str string) bool {
	_, err := time.Parse(time.RFC3339, str)
	return err == nil
}

// validateDate 验证日期格式（YYYY-MM-DD）
func validateDate(str string) bool {
	_, err := time.Parse("2006-01-02", str)
	return err == nil
}

// validateTime 验证时间格式（HH:MM:SS）
func validateTime(str string) bool {
	_, err := time.Parse("15:04:05", str)
	return err == nil
}

// validateURI 验证URI格式
func validateURI(str string) bool {
	_, err := url.ParseRequestURI(str)
	return err == nil
}

// validateHostname 验证主机名格式
func validateHostname(str string) bool {
	if len(str) > 255 {
		return false
	}
	if str == "" {
		return false
	}

	patternStr := `^([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])(\.([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]{0,61}[a-zA-Z0-9]))*$`
	pattern := regexp.MustCompile(patternStr)
	return pattern.MatchString(str)
}

// validateIPv4 验证IPv4地址格式
func validateIPv4(str string) bool {
	ip := net.ParseIP(str)
	return ip != nil && strings.Contains(str, ".")
}

// validateIPv6 验证IPv6地址格式
func validateIPv6(str string) bool {
	ip := net.ParseIP(str)
	return ip != nil && strings.Contains(str, ":")
}

// validateUUID 验证UUID格式
func validateUUID(str string) bool {
	pattern := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	return pattern.MatchString(strings.ToLower(str))
}

// 集合操作函数

// Contains 检查数组是否包含指定元素
func Contains(arr []interface{}, val interface{}) bool {
	for _, item := range arr {
		if reflect.DeepEqual(item, val) {
			return true
		}
	}
	return false
}

// Intersection 计算两个数组的交集
func Intersection(a, b []interface{}) []interface{} {
	result := make([]interface{}, 0)
	for _, item := range a {
		if Contains(b, item) {
			result = append(result, item)
		}
	}
	return result
}

// Union 计算两个数组的并集
func Union(a, b []interface{}) []interface{} {
	result := make([]interface{}, len(a))
	copy(result, a)

	for _, item := range b {
		if !Contains(result, item) {
			result = append(result, item)
		}
	}
	return result
}

// Difference 计算两个数组的差集（a - b）
func Difference(a, b []interface{}) []interface{} {
	result := make([]interface{}, 0)
	for _, item := range a {
		if !Contains(b, item) {
			result = append(result, item)
		}
	}
	return result
}

// 对象操作函数

// GetObjectKeys 获取对象的所有键
func GetObjectKeys(obj map[string]interface{}) []string {
	result := make([]string, 0, len(obj))
	for k := range obj {
		result = append(result, k)
	}
	return result
}

// HasKey 检查对象是否包含指定键
func HasKey(obj map[string]interface{}, key string) bool {
	_, exists := obj[key]
	return exists
}

// MergeObjects 合并两个对象
func MergeObjects(a, b map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// 复制第一个对象的所有键值
	for k, v := range a {
		result[k] = v
	}

	// 添加或覆盖第二个对象的键值
	for k, v := range b {
		result[k] = v
	}

	return result
}
