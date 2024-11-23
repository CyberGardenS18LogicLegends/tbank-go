package expenses

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
)

type AddExpenseRequest struct {
	Category    string  `json:"category"`
	Amount      float64 `json:"amount"`
	Date        string  `json:"date"`
	Description string  `json:"description,omitempty"`
}

// @Summary Add a new expense
// @Description Add a new expense for the authenticated user
// @Tags Expenses
// @Accept json
// @Produce plain
// @Param Authorization header string true "Bearer token"
// @Param expense body expenses.AddExpenseRequest true "Expense details"
// @Success 201 {string} string "Expense added successfully"
// @Failure 400 {string} string "Invalid input"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Failed to add expense"
// @Router /api/expenses [post]
func AddExpenseHandler(db *sql.DB, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req AddExpenseRequest
		userUID := r.Context().Value("userUID").(string)
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		query := `
		INSERT INTO expenses (user_uid, category, amount, date, description)
		VALUES (?, ?, ?, ?, ?);`

		_, err := db.Exec(query, userUID, req.Category, req.Amount, req.Date, req.Description)
		if err != nil {
			log.Error("failed to add expense", slog.Any("error", err))
			http.Error(w, "Failed to add expense", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Expense added successfully"))
		log.Info("Expense added successfully", slog.String("userUID", userUID), slog.String("category", req.Category), slog.Float64("amount", req.Amount))
	}
}
