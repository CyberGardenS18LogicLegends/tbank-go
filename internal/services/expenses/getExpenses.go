package expenses

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

type Expense struct {
	ID          int     `json:"id"`
	Category    string  `json:"category"`
	Amount      float64 `json:"amount"`
	Date        string  `json:"date"`        // Format: YYYY-MM-DD
	Description string  `json:"description"` // Description of the expense
}

// GetExpensesHandler @Summary Get expenses for a user in a given date range
// @Description Fetches all expenses for the authenticated user within the specified date range (from YYYY-MM-DD to YYYY-MM-DD)
// @Tags Expenses
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param from query string true "Start date (YYYY-MM-DD)"
// @Param to query string true "End date (YYYY-MM-DD)"
// @Success 200 {array} expenses.Expense "List of expenses"
// @Failure 400 {string} string "Invalid date format"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Failed to fetch expenses"
// @Router /api/expense [get]
func GetExpensesHandler(db *sql.DB, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userUID := r.Context().Value("userUID").(string)

		from := r.URL.Query().Get("from")
		to := r.URL.Query().Get("to")

		if _, err := time.Parse("2006-01-02", from); err != nil {
			log.Error("invalid from date format", slog.Any("error", err))
			http.Error(w, "Invalid from date format (YYYY-MM-DD)", http.StatusBadRequest)
			return
		}

		if _, err := time.Parse("2006-01-02", to); err != nil {
			log.Error("invalid to date format", slog.Any("error", err))
			http.Error(w, "Invalid to date format (YYYY-MM-DD)", http.StatusBadRequest)
			return
		}

		query := `
		SELECT id, category, amount, date, description
		FROM expenses
		WHERE user_uid = ? AND date BETWEEN ? AND ?;`

		rows, err := db.Query(query, userUID, from, to)
		if err != nil {
			log.Error("failed to fetch expenses", slog.Any("error", err))
			http.Error(w, "Failed to fetch expenses", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var expenses []Expense
		for rows.Next() {
			var expense Expense
			if err := rows.Scan(&expense.ID, &expense.Category, &expense.Amount, &expense.Date, &expense.Description); err != nil {
				log.Error("failed to scan expense", slog.Any("error", err))
				http.Error(w, "Error processing expenses", http.StatusInternalServerError)
				return
			}
			expenses = append(expenses, expense)
		}

		// Check for any errors while scanning
		if err := rows.Err(); err != nil {
			log.Error("error iterating over expenses", slog.Any("error", err))
			http.Error(w, "Error processing expenses", http.StatusInternalServerError)
			return
		}

		// Return the expenses in JSON format
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(expenses); err != nil {
			log.Error("failed to encode expenses", slog.Any("error", err))
			http.Error(w, "Failed to encode expenses", http.StatusInternalServerError)
			return
		}

		log.Info("expenses fetched successfully", slog.String("userUID", userUID), slog.String("from", from), slog.String("to", to))
	}
}
