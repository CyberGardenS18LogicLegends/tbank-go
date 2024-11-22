package services

import (
	"database/sql"
	"encoding/json"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
	"tbank-go/internal/user-service"
	"tbank-go/internal/utils"
)

func Register(db *sql.DB, w http.ResponseWriter, r *http.Request, log *slog.Logger) {
	var user user_service.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		log.Error("invalid input during registration", slog.Any("error", err))
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	log.Info("registering user", slog.String("username", user.Username))

	user.UID = uuid.NewString()
	log.Info("generated UID for user", slog.String("uid", user.UID))

	err = user.HashPassword(log)
	if err != nil {
		log.Error("error hashing password during registration", slog.String("username", user.Username), slog.Any("error", err))
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	err = user.Create(db, log)
	if err != nil {
		log.Error("error registering user", slog.String("username", user.Username), slog.Any("error", err))
		http.Error(w, "Error registering user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User registered successfully",
		"uid":     user.UID,
	})
}

func Login(db *sql.DB, w http.ResponseWriter, r *http.Request, log *slog.Logger, jwtSecret string) {
	var loginUser user_service.User
	err := json.NewDecoder(r.Body).Decode(&loginUser)
	if err != nil {
		log.Error("invalid input during login", slog.Any("error", err))
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	log.Info("logging in user", slog.String("username", loginUser.Username))

	user, err := user_service.GetUserByUsername(db, loginUser.Username, log)
	if err != nil {
		log.Error("user not found during login", slog.String("username", loginUser.Username), slog.Any("error", err))
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	err = utils.CheckPassword(user.Password, loginUser.Password)
	if err != nil {
		log.Error("invalid credentials during login", slog.String("username", loginUser.Username))
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	tokenString, err := utils.GenerateJWT(user.Username, jwtSecret)
	if err != nil {
		log.Error("error generating token during login", slog.String("username", loginUser.Username), slog.Any("error", err))
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	log.Info("user logged in successfully", slog.String("username", loginUser.Username))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token": tokenString,
		"uid":   user.UID,
	})
}

func ChangePassword(db *sql.DB, w http.ResponseWriter, r *http.Request, log *slog.Logger) {
	var requestBody struct {
		Username    string `json:"username"`
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		log.Error("invalid input during password change", slog.Any("error", err))
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	log.Info("changing password for user", slog.String("username", requestBody.Username))

	// Get user from database
	user, err := user_service.GetUserByUsername(db, requestBody.Username, log)
	if err != nil {
		log.Error("user not found during password change", slog.String("username", requestBody.Username), slog.Any("error", err))
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	// Verify old password
	err = utils.CheckPassword(user.Password, requestBody.OldPassword)
	if err != nil {
		log.Error("invalid old password during password change", slog.String("username", requestBody.Username))
		http.Error(w, "Invalid old password", http.StatusUnauthorized)
		return
	}

	// Hash new password
	user.Password = requestBody.NewPassword
	err = user.HashPassword(log)
	if err != nil {
		log.Error("error hashing new password during password change", slog.String("username", requestBody.Username), slog.Any("error", err))
		http.Error(w, "Error hashing new password", http.StatusInternalServerError)
		return
	}

	// Update password in database
	err = user_service.UpdatePassword(db, log, &user)
	if err != nil {
		log.Error("error updating password", slog.String("username", user.Username), slog.Any("error", err))
		http.Error(w, "Error updating password", http.StatusInternalServerError)
		return
	}

	log.Info("password changed successfully", slog.String("username", requestBody.Username))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Password updated successfully"))
}
