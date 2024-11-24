package user_service

import (
	"database/sql"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"time"
)

type User struct {
	UID          string `json:"uid"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	RegisteredAt string `json:"registered_at"`
}

func (u *User) HashPassword(log *slog.Logger) error {
	if u.UID == "" {
		u.UID = uuid.NewString()
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to hash password", slog.String("username", u.Username), slog.Any("error", err))
		return err

	}
	u.Password = string(hashedPassword)
	u.RegisteredAt = time.Now().Format(time.RFC3339)
	log.Info("password hashed successfully", slog.String("username", u.Username))
	return nil
}

func (u *User) Create(db *sql.DB, log *slog.Logger) error {
	_, err := db.Exec(
		`INSERT INTO users (uid, username, password, registered_at, first_name, second_name, incomes_balance, expenses_balance) 
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		u.UID, u.Username, u.Password, u.RegisteredAt, "", "", 0, 0,
	)

	if err != nil {
		log.Error("failed to create user", slog.String("username", u.Username), slog.Any("error", err))
		return err
	}
	log.Info("user added successfully", slog.String("username", u.Username))
	return nil
}

func GetUserByUsername(db *sql.DB, username string, log *slog.Logger) (User, error) {
	var user User
	err := db.QueryRow(
		"SELECT uid, username, password, registered_at FROM users WHERE username = ?", username,
	).Scan(&user.UID, &user.Username, &user.Password, &user.RegisteredAt)
	if err != nil {
		log.Error("failed to get user by username", slog.String("username", username), slog.Any("error", err))
		return user, err
	}
	log.Info("user retrieved successfully", slog.String("username", username))
	return user, err
}

func UpdatePassword(db *sql.DB, log *slog.Logger, user *User) error {
	_, err := db.Exec(
		"UPDATE users SET password = ? WHERE username = ?",
		user.Password, user.Username,
	)
	if err != nil {
		log.Error("failed to update password", slog.String("username", user.Username), slog.Any("error", err))
		return err
	}

	log.Info("password updated successfully", slog.String("username", user.Username))
	return nil
}
