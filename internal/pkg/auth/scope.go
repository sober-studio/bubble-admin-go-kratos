package auth

var scopePriority = map[string]int{
	"ALL":      4,
	"DEPT_SUB": 3,
	"DEPT":     2,
	"SELF":     1,
}

// GetGreaterScope 比较两个范围，返回较大的那个
func GetGreaterScope(oldScope, newScope string) string {
	if scopePriority[newScope] > scopePriority[oldScope] {
		return newScope
	}
	return oldScope
}
