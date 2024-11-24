package auth

import (
	"database/sql"
	"encoding/json"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
	"tbank-go/internal/user-service"
	"tbank-go/internal/utils"
	"time"
)

// AuthRequest defines the request body for both Register and Login endpoints.
// @Description Request body for user authentication operations.
type AuthRequest struct {
	Username string `json:"username" example:"johndoe"`
	Password string `json:"password" example:"password123"`
}

// ChangePasswordRequest defines the request body for the Change Password endpoint.
// @Description Request body for changing the user's password.
type ChangePasswordRequest struct {
	Username    string `json:"username" example:"johndoe"`
	OldPassword string `json:"old_password" example:"oldpassword123"`
	NewPassword string `json:"new_password" example:"newpassword123"`
}

// Register @Summary Register a new user
// @Description Create a new user in the system with a username and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param user body AuthRequest true "User Information"
// @Success 201 {object} map[string]string
// @Failure 400 {string} string "Invalid input"
// @Failure 500 {string} string "Error registering user"
// @Router /register [post]
func Register(db *sql.DB, w http.ResponseWriter, r *http.Request, log *slog.Logger) {
	var requestBody AuthRequest
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		log.Error("invalid input during registration", slog.Any("error", err))
		http.Error(w, "Invalid input", http.StatusBadRequest) // 400 Bad Request
		return
	}

	log.Info("registering user", slog.String("username", requestBody.Username))

	user := user_service.User{
		UID:      uuid.NewString(),
		Username: requestBody.Username,
		Password: requestBody.Password,
	}

	log.Info("generated UID for user", slog.String("uid", user.UID))

	err = user.HashPassword(log)
	if err != nil {
		log.Error("error hashing password during registration", slog.String("username", user.Username), slog.Any("error", err))
		http.Error(w, "Error hashing password", http.StatusInternalServerError) // 500 Internal Server Error
		return
	}

	err = user.Create(db, log)
	if err != nil {
		log.Error("error registering user", slog.String("username", user.Username), slog.Any("error", err))
		http.Error(w, "Error registering user", http.StatusInternalServerError) // 500 Internal Server Error
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201 Created
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User registered successfully",
		"uid":     user.UID,
	})
}

// @Summary Login a user
// @Description Authenticate user and return a JWT
// @Tags Auth
// @Accept json
// @Produce json
// @Param credentials body AuthRequest true "Login Credentials"
// @Success 200 {object} map[string]string
// @Failure 400 {string} string "Invalid input"
// @Failure 404 {string} string "User not found"
// @Failure 401 {string} string "Invalid credentials"
// @Failure 500 {string} string "Error generating token"
// @Router /login [post]
func Login(db *sql.DB, w http.ResponseWriter, r *http.Request, log *slog.Logger, jwtSecret string, jwtLifetime time.Duration) {
	var requestBody AuthRequest
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		log.Error("invalid input during login", slog.Any("error", err))
		http.Error(w, "Invalid input", http.StatusBadRequest) // 400 Bad Request
		return
	}

	log.Info("logging in user", slog.String("username", requestBody.Username))

	user, err := user_service.GetUserByUsername(db, requestBody.Username, log)
	if err != nil {
		log.Error("user not found during login", slog.String("username", requestBody.Username), slog.Any("error", err))
		http.Error(w, "User not found", http.StatusNotFound) // 404 Not Found
		return
	}

	err = utils.CheckPassword(user.Password, requestBody.Password)
	if err != nil {
		log.Error("invalid credentials during login", slog.String("username", requestBody.Username))
		http.Error(w, "Invalid credentials", http.StatusUnauthorized) // 401 Unauthorized
		return
	}

	tokenString, err := utils.GenerateJWT(user.UID, jwtSecret, jwtLifetime)
	if err != nil {
		log.Error("error generating token during login", slog.String("username", requestBody.Username), slog.Any("error", err))
		http.Error(w, "Error generating token", http.StatusInternalServerError) // 500 Internal Server Error
		return
	}

	log.Info("user logged in successfully", slog.String("username", requestBody.Username))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // 200 OK
	json.NewEncoder(w).Encode(map[string]string{
		"token": tokenString,
		"uid":   user.UID,
	})
}

// @Summary Change user password
// @Description Update the user's password after verifying the old password
// @Tags Auth
// @Accept json
// @Produce plain
// @Param changePassword body ChangePasswordRequest true "Change Password Information"
// @Success 200 {string} string "Password updated successfully"
// @Failure 400 {string} string "Invalid input"
// @Failure 404 {string} string "User not found"
// @Failure 401 {string} string "Invalid old password"
// @Failure 500 {string} string "Error updating password"
// @Router /change-password [post]
func ChangePassword(db *sql.DB, w http.ResponseWriter, r *http.Request, log *slog.Logger) {
	var requestBody ChangePasswordRequest
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		log.Error("invalid input during password change", slog.Any("error", err))
		http.Error(w, "Invalid input", http.StatusBadRequest) // 400 Bad Request
		return
	}
	log.Info("changing password for user", slog.String("username", requestBody.Username))

	user, err := user_service.GetUserByUsername(db, requestBody.Username, log)
	if err != nil {
		log.Error("user not found during password change", slog.String("username", requestBody.Username), slog.Any("error", err))
		http.Error(w, "User not found", http.StatusNotFound) // 404 Not Found
		return
	}

	err = utils.CheckPassword(user.Password, requestBody.OldPassword)
	if err != nil {
		log.Error("invalid old password during password change", slog.String("username", requestBody.Username))
		http.Error(w, "Invalid old password", http.StatusUnauthorized) // 401 Unauthorized
		return
	}

	user.Password = requestBody.NewPassword
	err = user.HashPassword(log)
	if err != nil {
		log.Error("error hashing new password during password change", slog.String("username", requestBody.Username), slog.Any("error", err))
		http.Error(w, "Error hashing new password", http.StatusInternalServerError) // 500 Internal Server Error
		return
	}

	err = user_service.UpdatePassword(db, log, &user)
	if err != nil {
		log.Error("error updating password", slog.String("username", user.Username), slog.Any("error", err))
		http.Error(w, "Error updating password", http.StatusInternalServerError) // 500 Internal Server Error
		return
	}

	log.Info("password changed successfully", slog.String("username", requestBody.Username))
	w.WriteHeader(http.StatusOK) // 200 OK
	w.Write([]byte("Password updated successfully"))
}
