package auth

import (
	"database/sql"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"tbank-go/internal/models"
)

type UserService struct {
	DB     *sql.DB
	Logger *slog.Logger
}

func (us *UserService) CreateUser(user *models.User) error {
	user.UID = uuid.NewString()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		us.Logger.Error("failed to hash password", slog.String("username", user.Username), slog.Any("error", err))
		return err
	}
	user.Password = string(hashedPassword)
	user.RegisteredAt = time.Now().Format(time.RFC3339)

	_, err = us.DB.Exec("INSERT INTO users (uid, username, password, registered_at) VALUES (?, ?, ?, ?)", user.UID, user.Username, user.Password, user.RegisteredAt)
	if err != nil {
		us.Logger.Error("failed to create auth", slog.String("username", user.Username), slog.Any("error", err))
	} else {
		us.Logger.Info("auth created successfully", slog.String("uid", user.UID), slog.String("username", user.Username))
	}
	return err
}

func (us *UserService) GetUserByUsername(username string) (models.User, error) {
	var user models.User
	err := us.DB.QueryRow("SELECT uid, username, password, registered_at FROM users WHERE username = ?", username).Scan(&user.UID, &user.Username, &user.Password, &user.RegisteredAt)
	if err != nil {
		us.Logger.Error("failed to get auth by username", slog.String("username", username), slog.Any("error", err))
	} else {
		us.Logger.Info("auth retrieved successfully", slog.String("uid", user.UID), slog.String("username", user.Username))
	}
	return user, err
}

func (us *UserService) UpdatePassword(user *models.User) error {
	_, err := us.DB.Exec("UPDATE users SET password = ? WHERE uid = ?", user.Password, user.UID)
	if err != nil {
		us.Logger.Error("failed to update password", slog.String("uid", user.UID), slog.Any("error", err))
	} else {
		us.Logger.Info("password updated successfully", slog.String("uid", user.UID))
	}
	return err
}

func (us *UserService) GetUIDByUsername(username string) (string, error) {
	var uid string
	err := us.DB.QueryRow("SELECT uid FROM users WHERE username = ?", username).Scan(&uid)
	if err != nil {
		us.Logger.Error("failed to get uid by username", slog.String("username", username), slog.Any("error", err))
	} else {
		us.Logger.Info("uid retrieved successfully", slog.String("uid", uid), slog.String("username", username))
	}
	return uid, err
}
