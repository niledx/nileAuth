package main

import (
	"encoding/json"
	"net/http"
	"time"
)

type creds struct{ Email, Password string }

func (a *App) HandleRegister(w http.ResponseWriter, r *http.Request) {
	var c creds
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}
	if c.Email == "" || c.Password == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Email and password are required")
		return
	}

	// Get application from context if available
	var appID *int64
	if app, ok := r.Context().Value("application").(*Application); ok && app != nil {
		appID = &app.ID
	}

	hashed, err := hashPassword(c.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to process password")
		return
	}
	user, err := a.DB.CreateUser(c.Email, hashed, appID)
	if err != nil {
		writeError(w, http.StatusConflict, "USER_EXISTS", "User with this email already exists")
		return
	}
	access, _ := createAccessToken(user.ID)
	ref, _ := genToken(32)
	a.DB.CreateRefreshToken(ref, user.ID, time.Now().Add(30*24*time.Hour).Unix(), appID)
	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"user": map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
		},
		"accessToken":  access,
		"refreshToken": ref,
	})
}

func (a *App) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var c creds
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}
	user, err := a.DB.GetUserByEmail(c.Email)
	if err != nil || user == nil {
		writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password")
		return
	}
	if !comparePassword(user.Password, c.Password) {
		writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password")
		return
	}

	// Get application from context if available
	var appID *int64
	if app, ok := r.Context().Value("application").(*Application); ok && app != nil {
		appID = &app.ID
	}

	access, _ := createAccessToken(user.ID)
	ref, _ := genToken(32)
	a.DB.CreateRefreshToken(ref, user.ID, time.Now().Add(30*24*time.Hour).Unix(), appID)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user": map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
		},
		"accessToken":  access,
		"refreshToken": ref,
	})
}

func (a *App) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	var in struct{ RefreshToken string }
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}
	if in.RefreshToken == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Refresh token is required")
		return
	}
	row, _ := a.DB.GetRefreshToken(in.RefreshToken)
	if row == nil {
		writeError(w, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid refresh token")
		return
	}
	if row.Revoked {
		a.DB.RevokeAllRefreshTokensForUser(row.UserID)
		writeError(w, http.StatusUnauthorized, "TOKEN_REUSE_DETECTED", "Token reuse detected - all tokens revoked")
		return
	}
	if row.ExpiresAt < time.Now().Unix() {
		writeError(w, http.StatusUnauthorized, "TOKEN_EXPIRED", "Refresh token has expired")
		return
	}

	// Get application from context if available
	var appID *int64
	if app, ok := r.Context().Value("application").(*Application); ok && app != nil {
		appID = &app.ID
	} else if row.ApplicationID != nil {
		appID = row.ApplicationID
	}

	// rotate
	a.DB.RevokeRefreshToken(in.RefreshToken)
	newRef, _ := genToken(32)
	a.DB.CreateRefreshToken(newRef, row.UserID, time.Now().Add(30*24*time.Hour).Unix(), appID)
	access, _ := createAccessToken(row.UserID)
	writeJSON(w, http.StatusOK, map[string]string{
		"accessToken":  access,
		"refreshToken": newRef,
	})
}

func (a *App) HandleLogout(w http.ResponseWriter, r *http.Request) {
	var in struct{ RefreshToken string }
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}
	if in.RefreshToken == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Refresh token is required")
		return
	}
	err := a.DB.RevokeRefreshToken(in.RefreshToken)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_TOKEN", "Token not found or already revoked")
		return
	}
	writeSuccess(w, http.StatusOK, map[string]bool{"revoked": true})
}
