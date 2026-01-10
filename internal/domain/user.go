package domain

type User struct {
	ID             int    `json:"id"`
	Email          string `json:"email"`
	Phone          string `json:"phone"`
	PasswordHashed string `json:"password_hashed"`
	Role           int    `json:"role"`
}

type UserUnsafe struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
	Role     int    `json:"role"`
}
