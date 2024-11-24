package users

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
)

// GetUserInfoHandler fetches all user information
// @Summary Get User Info
// @Description Fetches all information about the logged-in user.
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "User info"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/users/ [get]
func GetUserInfoHandler(db *sql.DB, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userUID := r.Context().Value("userUID").(string)

		query := `SELECT id, uid, username, first_name, second_name, incomes_balance, expenses_balance FROM users WHERE uid = ?`
		row := db.QueryRow(query, userUID)

		var user struct {
			ID              int    `json:"id"`
			UID             string `json:"uid"`
			Username        string `json:"username"`
			FirstName       string `json:"first_name,omitempty"`
			SecondName      string `json:"second_name,omitempty"`
			IncomesBalance  int    `json:"incomes_balance"`
			ExpensesBalance int    `json:"expenses_balance"`
		}

		if err := row.Scan(&user.ID, &user.UID, &user.Username, &user.FirstName, &user.SecondName, &user.IncomesBalance, &user.ExpensesBalance); err != nil {
			log.Error("failed to fetch user info", slog.Any("error", err))
			http.Error(w, "Failed to fetch user info", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(user)
	}
}
