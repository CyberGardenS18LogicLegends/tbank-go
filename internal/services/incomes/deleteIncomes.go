package incomes

import (
	"database/sql"
	"github.com/go-chi/chi/v5"
	"log/slog"
	"net/http"
)

// DeleteIncomeHandler deletes an income record by its ID and adjusts the user's income balance
// @Summary Delete Income by ID
// @Description Deletes a specific income record by its unique ID and updates the user's income balance.
// @Tags Incomes
// @Accept json
// @Produce json
// @Param id path int true "Income ID"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success message"
// @Failure 400 {object} map[string]interface{} "Invalid ID"
// @Failure 404 {object} map[string]interface{} "Income not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/income/{id} [delete]
func DeleteIncomeHandler(db *sql.DB, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		incomeID := chi.URLParam(r, "id")
		if incomeID == "" {
			log.Error("missing income ID parameter")
			http.Error(w, "Income ID is required", http.StatusBadRequest)
			return
		}

		userUID := r.Context().Value("userUID").(string)

		var ownerUID string
		var incomeAmount float64
		query := `SELECT user_uid, amount FROM income WHERE id = ?`
		err := db.QueryRow(query, incomeID).Scan(&ownerUID, &incomeAmount)
		if err == sql.ErrNoRows {
			log.Warn("income not found", slog.String("incomeID", incomeID))
			http.Error(w, "Income not found", http.StatusNotFound)
			return
		} else if err != nil {
			log.Error("failed to fetch income details", slog.Any("error", err))
			http.Error(w, "Failed to fetch income details", http.StatusInternalServerError)
			return
		}

		if ownerUID != userUID {
			log.Warn("unauthorized attempt to delete income", slog.String("userUID", userUID), slog.String("ownerUID", ownerUID))
			http.Error(w, "Unauthorized to delete this income", http.StatusForbidden)
			return
		}

		tx, err := db.Begin()
		if err != nil {
			log.Error("failed to start transaction", slog.Any("error", err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		deleteQuery := `DELETE FROM income WHERE id = ?`
		result, err := tx.Exec(deleteQuery, incomeID)
		if err != nil {
			tx.Rollback()
			log.Error("failed to delete income", slog.String("incomeID", incomeID), slog.Any("error", err))
			http.Error(w, "Failed to delete income", http.StatusInternalServerError)
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil || rowsAffected == 0 {
			tx.Rollback()
			log.Warn("income not found during deletion", slog.String("incomeID", incomeID))
			http.Error(w, "Income not found", http.StatusNotFound)
			return
		}

		updateBalanceQuery := `UPDATE users SET incomes_balance = incomes_balance - ? WHERE uid = ?`
		_, err = tx.Exec(updateBalanceQuery, incomeAmount, userUID)
		if err != nil {
			tx.Rollback()
			log.Error("failed to update income balance", slog.Any("error", err))
			http.Error(w, "Failed to update income balance", http.StatusInternalServerError)
			return
		}

		if err := tx.Commit(); err != nil {
			log.Error("failed to commit transaction", slog.Any("error", err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		log.Info("income deleted successfully", slog.String("incomeID", incomeID), slog.String("userUID", userUID))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "Income deleted successfully"}`))
	}
}
