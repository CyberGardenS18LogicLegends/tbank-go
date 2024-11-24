package users

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
)

// UpdateUserNamesRequest defines the structure for the request body
type UpdateUserNamesRequest struct {
	FirstName  string `json:"first_name" example:"John"`
	SecondName string `json:"second_name" example:"Doe"`
}

// UpdateUserNamesHandler updates the first and second name of a user
// @Summary Update User Names
// @Description Updates the first and second name of the logged-in user.
// @Tags Users
// @Accept json
// @Produce json
// @Param body body UpdateUserNamesRequest true "First and Second Name"
// @Security BearerAuth
// @Success 200 {object} map[string]string "Success message"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/users/ [put]
func UpdateUserNamesHandler(db *sql.DB, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the request body
		var data UpdateUserNamesRequest
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			log.Error("invalid JSON body", slog.Any("error", err))
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		if data.FirstName == "" || data.SecondName == "" {
			log.Error("missing fields in input", slog.Any("data", data))
			http.Error(w, "Both first_name and second_name are required", http.StatusBadRequest)
			return
		}

		userUID := r.Context().Value("userUID").(string)

		query := `UPDATE users SET first_name = ?, second_name = ? WHERE uid = ?`
		_, err := db.Exec(query, data.FirstName, data.SecondName, userUID)
		if err != nil {
			log.Error("failed to update user names", slog.Any("error", err))
			http.Error(w, "Failed to update user names", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "User names updated successfully"})
	}
}
