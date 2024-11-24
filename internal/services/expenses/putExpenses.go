package expenses

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
)

type UpdateExpenseRequest struct {
	Category    string  `json:"category"`
	Amount      float64 `json:"amount"`
	Date        string  `json:"date"`
	Description string  `json:"description,omitempty"`
}

// AddExpenseHandler adds an expense and adjusts the user's balance
// @Summary Add Expense
// @Description Adds a new expense record and adjusts the user's expense balance.
// @Tags Expenses
// @Accept json
// @Produce plain
// @Param Authorization header string true "Bearer token"
// @Param expense body expenses.UpdateExpenseRequest true "New expense details"
// @Success 200 {string} string "Expense added successfully"
// @Failure 400 {string} string "Invalid input"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Failed to add expense"
// @Router /api/expense [post]
func AddExpenseHandler(db *sql.DB, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req UpdateExpenseRequest
		userUID := r.Context().Value("userUID").(string)
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		var expenseID string
		query := `INSERT INTO expenses (user_uid, category, amount, date, description) 
		          VALUES (?, ?, ?, ?, ?) RETURNING id`
		err := db.QueryRow(query, userUID, req.Category, req.Amount, req.Date, req.Description).
			Scan(&expenseID)
		if err != nil {
			log.Error("failed to insert expense", slog.Any("error", err))
			http.Error(w, "Failed to insert expense", http.StatusInternalServerError)
			return
		}

		tx, err := db.Begin()
		if err != nil {
			log.Error("failed to start transaction", slog.Any("error", err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		adjustBalanceQuery := `UPDATE users SET expenses_balance = expenses_balance + ? WHERE uid = ?`
		_, err = tx.Exec(adjustBalanceQuery, req.Amount, userUID)
		if err != nil {
			tx.Rollback()
			log.Error("failed to adjust expense balance", slog.Any("error", err))
			http.Error(w, "Failed to adjust expense balance", http.StatusInternalServerError)
			return
		}

		if err := tx.Commit(); err != nil {
			log.Error("failed to commit transaction", slog.Any("error", err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		log.Info("expense added successfully", slog.String("expenseID", expenseID), slog.String("userUID", userUID), slog.Float64("amount", req.Amount))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Expense added successfully"))
	}
}
