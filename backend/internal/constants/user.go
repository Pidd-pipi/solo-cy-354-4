package constants

// UserRole defines user role enum values shared with the frontend.
const (
	UserRoleStudent = "student"
	UserRoleAdmin   = "admin"
)

// UserRoles lists all valid user roles.
var UserRoles = []string{UserRoleStudent, UserRoleAdmin}

// IsUserRole reports whether the given role is valid.
func IsUserRole(r string) bool {
	for _, v := range UserRoles {
		if v == r {
			return true
		}
	}
	return false
}

// UserRoleText returns the Chinese label of a user role.
func UserRoleText(r string) string {
	switch r {
	case UserRoleStudent:
		return "学生"
	case UserRoleAdmin:
		return "管理员"
	default:
		return "未知"
	}
}

// CreditLevelText maps credit score to a Chinese credit grade.
func CreditLevelText(score int) string {
	switch {
	case score >= 200:
		return "极佳"
	case score >= 150:
		return "优秀"
	case score >= 100:
		return "良好"
	case score >= 60:
		return "一般"
	default:
		return "待提升"
	}
}
