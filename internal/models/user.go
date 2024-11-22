package models

type User struct {
	UID          string `json:"uid"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	RegisteredAt string `json:"registered_at"`
}
