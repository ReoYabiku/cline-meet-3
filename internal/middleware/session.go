package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type contextKey string

const sessionIDKey contextKey = "session_id"

const SessionCookieName = "signaling_session"

// SessionMiddleware handles session management via cookies
func SessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if session cookie exists
		cookie, err := r.Cookie(SessionCookieName)
		if err != nil || cookie.Value == "" {
			// Create new session
			sessionID := uuid.New().String()
			http.SetCookie(w, &http.Cookie{
				Name:     SessionCookieName,
				Value:    sessionID,
				Path:     "/",
				MaxAge:   86400, // 24 hours
				HttpOnly: true,
				Secure:   false, // Set to true in production with HTTPS
				SameSite: http.SameSiteLaxMode,
			})
			
			// Add session ID to request context
			r = r.WithContext(setSessionID(r.Context(), sessionID))
		} else {
			// Use existing session
			r = r.WithContext(setSessionID(r.Context(), cookie.Value))
		}

		next.ServeHTTP(w, r)
	})
}

// GetSessionID retrieves session ID from request context
func GetSessionID(r *http.Request) string {
	if sessionID := getSessionIDFromContext(r.Context()); sessionID != "" {
		return sessionID
	}
	
	// Fallback to cookie
	if cookie, err := r.Cookie(SessionCookieName); err == nil {
		return cookie.Value
	}
	
	return ""
}

// setSessionID adds session ID to context
func setSessionID(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, sessionIDKey, sessionID)
}

// getSessionIDFromContext retrieves session ID from context
func getSessionIDFromContext(ctx context.Context) string {
	if sessionID, ok := ctx.Value(sessionIDKey).(string); ok {
		return sessionID
	}
	return ""
}
