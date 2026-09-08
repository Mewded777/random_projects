package webui

import (
	"net/http"
	"time"
)

// SetUserSession sets simple plaintext cookies for demonstration.
// targetID is the domain.User's TargetID (the linked Member.ID for a
// member account, empty for staff accounts) — it's what lets the
// catalog page know which member is borrowing without the user having
// to type their own ID in every time.
func SetUserSession(w http.ResponseWriter, username string, role string, targetID string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_user",
		Value:    username,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true, // Prevents JavaScript XSS attacks
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "session_role",
		Value:    role,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "session_target",
		Value:    targetID,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	})
}

// ClearUserSession removes the cookies on logout
func ClearUserSession(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_user",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "session_role",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "session_target",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
}

// GetUserSession reads the active session strings: username, role, and the linked member ID (targetID, empty for staff accounts).
func GetUserSession(r *http.Request) (username string, role string, targetID string) {
	userCookie, errUser := r.Cookie("session_user")
	roleCookie, errRole := r.Cookie("session_role")
	if errUser != nil || errRole != nil {
		return "", "", ""
	}
	if targetCookie, err := r.Cookie("session_target"); err == nil {
		targetID = targetCookie.Value
	}
	return userCookie.Value, roleCookie.Value, targetID
}
