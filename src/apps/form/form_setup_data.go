package form

type FormSetupData struct {
	PageTitle             string
	ErrorMessage          string
	Email                 string
	EmailError            bool
	UserExistsError       bool
	EmailInvalidError     bool
	PasswordMatchError    bool
	PasswordStrengthError bool
	PasswordCrackTime     string
}
