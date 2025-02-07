package authentication

import "context"

// Repo định nghĩa interface cho authentication repository
type Repo interface {
	// Register đăng ký người dùng mới
	Register(ctx context.Context, firstName, lastName, email, password string) error

	// Login xác thực người dùng và trả về thông tin nếu thành công
	Login(ctx context.Context, req LoginRequest) (*User, bool, error)

	// Logout đăng xuất người dùng bằng cách xóa session
	Logout(ctx context.Context, userID uint64) (bool, error)

	// Csrf tạo và lưu trữ CSRF token
	Csrf(ctx context.Context) (string, error)
}
