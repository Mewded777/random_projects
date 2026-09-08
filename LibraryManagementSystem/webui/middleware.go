package webui

import (
	"library/domain"
	"net/http"
)

// RequireRole checks if the logged-in user matches any of the allowed roles
func RequireRole(allowedRoles ...domain.Role) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			_, roleStr, _ := GetUserSession(r)

			// If no session exists, redirect straight to login
			if roleStr == "" {
				http.Redirect(w, r, "/login?error=please_log_in", http.StatusSeeOther)
				return
			}

			userRole := domain.Role(roleStr)
			for _, allowed := range allowedRoles {
				if userRole == allowed {
					next(w, r)
					return
				}
			}

			// If authenticated but role doesn't have permissions, deny access
			http.Error(w, "Access Denied: Insufficient Permissions", http.StatusForbidden)
		}
	}
}
