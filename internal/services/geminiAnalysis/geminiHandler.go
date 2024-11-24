package geminiAnalysis

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// Expense struct to hold the expense data
type Expense struct {
	Category    string  `json:"category"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
}

// FinancialAdviceResponse struct to hold the response from Gemini
type FinancialAdviceResponse struct {
	Advice string `json:"advice"`
}

// GenerateFinancialAdviceHandler handles the request for financial advice based on user's expenses
// @Summary Generate Financial Advice
// @Description Provides financial advice based on the user's expenses
// @Tags Expenses
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} FinancialAdviceResponse "Financial advice provided successfully"
// @Failure 400 {string} string "Invalid input"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Failed to generate financial advice"
// @Router /api/financial-advice [get]
func GenerateFinancialAdviceHandler(db *sql.DB, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userUID := r.Context().Value("userUID").(string)
		log.Debug("Fetching user expenses", slog.String("userUID", userUID))

		expenses, err := getUserExpenses(db, userUID)
		if err != nil {
			log.Error("Failed to fetch user expenses", slog.Any("error", err), slog.String("userUID", userUID))
			http.Error(w, "Failed to fetch user expenses", http.StatusInternalServerError)
			return
		}

		log.Debug("User expenses fetched", slog.Int("numExpenses", len(expenses)))

		prompt := constructPrompt(expenses)

		client, err := genai.NewClient(context.Background(), option.WithAPIKey("AIzaSyCXAxfBfbQ4M5l_I8HSP1N6FdIms8pg0Z0"))
		if err != nil {
			log.Error("Failed to initialize Gemini client", slog.Any("error", err))
			http.Error(w, "Failed to initialize Gemini client", http.StatusInternalServerError)
			return
		}
		defer client.Close()

		model := client.GenerativeModel("gemini-1.5-flash")
		resp, err := model.GenerateContent(context.Background(), genai.Text(prompt))
		if err != nil {
			log.Error("Failed to get response from Gemini", slog.Any("error", err))
			http.Error(w, "Failed to get response from Gemini", http.StatusInternalServerError)
			return
		}

		log.Debug("Received response from Gemini")

		advice := extractAdviceFromResponse(resp)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(FinancialAdviceResponse{Advice: advice})

		log.Info("Financial advice generated successfully", slog.String("advice", advice))
	}
}

// getUserExpenses retrieves the expenses for the user from the database
func getUserExpenses(db *sql.DB, userUID string) ([]Expense, error) {
	var expenses []Expense

	query := `SELECT category, amount, description FROM expenses WHERE user_uid = ?`
	rows, err := db.Query(query, userUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var expense Expense
		if err := rows.Scan(&expense.Category, &expense.Amount, &expense.Description); err != nil {
			return nil, err
		}
		expenses = append(expenses, expense)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return expenses, nil
}

// constructPrompt builds the prompt for Gemini based on user's expenses
func constructPrompt(expenses []Expense) string {
	expensesText := "Ваши текущие расходы:\n"
	for _, expense := range expenses {
		expensesText += fmt.Sprintf("%s: %.2f руб. (%s)\n", expense.Category, expense.Amount, expense.Description)
	}

	prompt := fmt.Sprintf(`
%s
На основе этих данных, пожалуйста, дайте рекомендации, где можно сократить расходы и какие категории наиболее неэффективны. Все цены описаны в рублях. Продолжение диалога не планируется.
`, expensesText)

	return prompt
}

// extractAdviceFromResponse extracts financial advice from Gemini's response
func extractAdviceFromResponse(resp *genai.GenerateContentResponse) string {
	var advice string
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				advice += fmt.Sprintf("%s", part)
			}
		}
	}
	return advice
}
