package incomes

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

type Income struct {
	Category    string  `json:"category"`
	Amount      float64 `json:"amount"`
	Date        string  `json:"date"`        // Формат: YYYY-MM-DD
	Description string  `json:"description"` // Описание
}

// @Summary Add a new income
// @Description Add a new income record for the authenticated user
// @Tags Incomes
// @Accept json
// @Produce plain
// @Param Authorization header string true "Bearer token"
// @Param income body incomes.Income true "Income details"
// @Success 201 {string} string "Income added successfully"
// @Failure 400 {string} string "Invalid input"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Failed to add income"
// @Router /api/incomes [post]
func AddIncomeHandler(db *sql.DB, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var income Income
		userUID := r.Context().Value("userUID").(string)

		if err := json.NewDecoder(r.Body).Decode(&income); err != nil {
			log.Error("invalid input for income", slog.Any("error", err))
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		// Проверяем дату
		if _, err := time.Parse("2006-01-02", income.Date); err != nil {
			log.Error("invalid date format", slog.Any("error", err))
			http.Error(w, "Invalid date format (YYYY-MM-DD)", http.StatusBadRequest)
			return
		}

		query := `INSERT INTO income (user_uid, category, amount, date, description) VALUES (?, ?, ?, ?, ?)`
		_, err := db.Exec(query, userUID, income.Category, income.Amount, income.Date, income.Description)
		if err != nil {
			log.Error("failed to add income", slog.Any("error", err))
			http.Error(w, "Failed to add income", http.StatusInternalServerError)
			return
		}

		log.Info("income added successfully", slog.String("userUID", userUID))
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Income added successfully"))
	}
}
