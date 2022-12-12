package types

type VerifyData struct {
	Otp   string
	Email string
}

type LoginDetails struct {
	Password string
	Account  string
}

type VerifyPhoneData struct {
	Phone string
	Otp   string
}

type UpdatePassword struct {
	CurrentPassword string
	NewPassword     string
}
