package auth

type SignUpInput struct {
	Email    string
	Username string
	Password string
}

type SignInInput struct {
	Email    string
	Password string
}

type RefreshInput struct {
	RefreshToken string
}

type LogoutInput struct {
	RefreshToken string
}
