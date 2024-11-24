package incomes

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

// GetIncomesHandler retrieves incomes for a user within the specified date range
// @Summary Get Incomes within a date range
// @Description Retrieves all income records for a user within the specified date range.
// @Tags Incomes
// @Accept json
// @Produce json
// @Param from query string true "Start date (YYYY-MM-DD)"
// @Param to query string true "End date (YYYY-MM-DD)"
// @Security BearerAuth
// @Success 200 {array} map[string]interface{} "Incomes list"
// @Failure 400 {object} map[string]interface{} "Invalid date format or missing parameters"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/income [get]
func GetIncomesHandler(db *sql.DB, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startDate := r.URL.Query().Get("from")
		endDate := r.URL.Query().Get("to")

		if startDate == "" || endDate == "" {
			log.Error("missing date parameter", slog.String("start_date", startDate), slog.String("end_date", endDate))
			http.Error(w, "Both start_date and end_date are required", http.StatusBadRequest)
			return
		}

		if _, err := time.Parse("2006-01-02", startDate); err != nil {
			log.Error("invalid start_date format", slog.Any("error", err))
			http.Error(w, "Invalid start_date format (YYYY-MM-DD)", http.StatusBadRequest)
			return
		}
		if _, err := time.Parse("2006-01-02", endDate); err != nil {
			log.Error("invalid end_date format", slog.Any("error", err))
			http.Error(w, "Invalid end_date format (YYYY-MM-DD)", http.StatusBadRequest)
			return
		}

		userUID := r.Context().Value("userUID").(string)

		query := `
			SELECT id, category, amount, date, description 
			FROM income 
			WHERE user_uid = ? AND date BETWEEN ? AND ?
		`

		rows, err := db.Query(query, userUID, startDate, endDate)
		if err != nil {
			log.Error("failed to fetch incomes", slog.Any("error", err))
			http.Error(w, "Failed to fetch incomes", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var incomes []map[string]interface{}
		for rows.Next() {
			var id int
			var category, date, description string
			var amount float64
			if err := rows.Scan(&id, &category, &amount, &date, &description); err != nil {
				log.Error("failed to scan row", slog.Any("error", err))
				http.Error(w, "Failed to scan income data", http.StatusInternalServerError)
				return
			}
			incomes = append(incomes, map[string]interface{}{
				"id":          id,
				"category":    category,
				"amount":      amount,
				"date":        date,
				"description": description,
			})
		}

		if err := rows.Err(); err != nil {
			log.Error("failed to iterate over rows", slog.Any("error", err))
			http.Error(w, "Failed to iterate over income data", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(incomes)
	}
}
