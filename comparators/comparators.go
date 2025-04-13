package comparators

// CompareFunc 定义了比较函数的签名
// a, b: 要比较的两个值
// 返回: 比较结果（true表示满足比较条件）
type CompareFunc func(a, b interface{}) bool

// Comparator 接口定义了比较器的行为
type Comparator interface {
	// Name 返回比较器的名称
	Name() string

	// Compare 执行比较
	Compare(a, b interface{}) bool
}

// BaseComparator 是所有比较器的基础实现
type BaseComparator struct {
	name string
	fn   CompareFunc
}

// NewComparator 创建一个新的比较器
func NewComparator(name string, fn CompareFunc) Comparator {
	return &BaseComparator{
		name: name,
		fn:   fn,
	}
}

// Name 返回比较器的名称
func (c *BaseComparator) Name() string {
	return c.name
}

// Compare 执行比较
func (c *BaseComparator) Compare(a, b interface{}) bool {
	return c.fn(a, b)
}

func geComparator(a, b interface{}) bool {
	aFloat, ok := toFloat64(a)
	if !ok {
		return false
	}
	bFloat, ok := toFloat64(b)
	if !ok {
		return false
	}
	return aFloat >= bFloat
}

func GetGeComparator() CompareFunc {
	return geComparator
}

// ComparatorRegistry 接口定义了比较器注册表的行为
type ComparatorRegistry interface {
	// RegisterComparator 注册比较器
	RegisterComparator(name string, fn CompareFunc) error

	// GetComparator 获取比较器
	GetComparator(name string) CompareFunc
}
