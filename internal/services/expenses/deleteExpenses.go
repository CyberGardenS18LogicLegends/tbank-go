package expenses

import (
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// DeleteExpenseHandler deletes an expense by its ID
// @Summary Delete Expense
// @Description Deletes a specific expense record by ID.
// @Tags Expenses
// @Accept json
// @Produce json
// @Param id path string true "Expense ID"
// @Security BearerAuth
// @Success 200 {object} map[string]string "Success message"
// @Failure 400 {object} map[string]string "Invalid ID parameter"
// @Failure 404 {object} map[string]string "Expense not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/expense/{id} [delete]
func DeleteExpenseHandler(db *sql.DB, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		expenseID := chi.URLParam(r, "id")
		if expenseID == "" {
			log.Error("missing expense ID parameter")
			http.Error(w, "Expense ID is required", http.StatusBadRequest)
			return
		}

		userUID := r.Context().Value("userUID").(string)

		// Check ownership of the expense
		var ownerUID string
		query := `SELECT user_uid FROM expenses WHERE id = ?`
		err := db.QueryRow(query, expenseID).Scan(&ownerUID)
		if err == sql.ErrNoRows {
			log.Warn("expense not found", slog.String("expenseID", expenseID))
			http.Error(w, "Expense not found", http.StatusNotFound)
			return
		} else if err != nil {
			log.Error("failed to fetch expense owner", slog.Any("error", err))
			http.Error(w, "Failed to fetch expense owner", http.StatusInternalServerError)
			return
		}

		if ownerUID != userUID {
			log.Warn("unauthorized attempt to delete expense", slog.String("userUID", userUID), slog.String("ownerUID", ownerUID))
			http.Error(w, "Unauthorized to delete this expense", http.StatusForbidden)
			return
		}

		deleteQuery := `DELETE FROM expenses WHERE id = ?`
		result, err := db.Exec(deleteQuery, expenseID)
		if err != nil {
			log.Error("failed to delete expense", slog.String("expenseID", expenseID), slog.Any("error", err))
			http.Error(w, "Failed to delete expense", http.StatusInternalServerError)
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			log.Error("failed to get rows affected", slog.Any("error", err))
			http.Error(w, "Failed to determine deletion status", http.StatusInternalServerError)
			return
		}

		if rowsAffected == 0 {
			log.Warn("expense not found during deletion", slog.String("expenseID", expenseID))
			http.Error(w, "Expense not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "Expense deleted successfully"}`))
	}
}
