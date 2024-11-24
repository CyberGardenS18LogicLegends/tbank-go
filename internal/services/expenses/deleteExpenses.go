package expenses

import (
	"database/sql"
	"github.com/go-chi/chi/v5"
	"log/slog"
	"net/http"
)

// DeleteExpenseHandler deletes an expense by its ID and adjusts the user's expense balance
// @Summary Delete Expense
// @Description Deletes a specific expense record and updates the user's expense balance.
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

		var ownerUID string
		var expenseAmount float64
		query := `SELECT user_uid, amount FROM expenses WHERE id = ?`
		err := db.QueryRow(query, expenseID).Scan(&ownerUID, &expenseAmount)
		if err == sql.ErrNoRows {
			log.Warn("expense not found", slog.String("expenseID", expenseID))
			http.Error(w, "Expense not found", http.StatusNotFound)
			return
		} else if err != nil {
			log.Error("failed to fetch expense details", slog.Any("error", err))
			http.Error(w, "Failed to fetch expense details", http.StatusInternalServerError)
			return
		}

		if ownerUID != userUID {
			log.Warn("unauthorized attempt to delete expense", slog.String("userUID", userUID), slog.String("ownerUID", ownerUID))
			http.Error(w, "Unauthorized to delete this expense", http.StatusForbidden)
			return
		}

		tx, err := db.Begin()
		if err != nil {
			log.Error("failed to start transaction", slog.Any("error", err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		deleteQuery := `DELETE FROM expenses WHERE id = ?`
		result, err := tx.Exec(deleteQuery, expenseID)
		if err != nil {
			tx.Rollback()
			log.Error("failed to delete expense", slog.String("expenseID", expenseID), slog.Any("error", err))
			http.Error(w, "Failed to delete expense", http.StatusInternalServerError)
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil || rowsAffected == 0 {
			tx.Rollback()
			log.Warn("expense not found during deletion", slog.String("expenseID", expenseID))
			http.Error(w, "Expense not found", http.StatusNotFound)
			return
		}

		updateBalanceQuery := `UPDATE users SET expenses_balance = expenses_balance - ? WHERE uid = ?`
		_, err = tx.Exec(updateBalanceQuery, expenseAmount, userUID)
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

		log.Info("expense deleted successfully", slog.String("expenseID", expenseID), slog.String("userUID", userUID))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "Expense deleted successfully"}`))
	}
}
