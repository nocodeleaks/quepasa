package form

import "net/http"

// LogoutHandler renders route GET "/logout"
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{
		Name:     "jwt",
		Value:    "",
		MaxAge:   0,
		Path:     "/",
		HttpOnly: true,
	}

	http.SetCookie(w, cookie)

	RedirectToLogin(w, r)
}
