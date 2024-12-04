package server

import (
	"encoding/json"
	"meramoney/backend/infrastructure/domains"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type TransactionRequest struct {
	CreatedAt   *string `json:"created_at"`
	CategoryID  int     `json:"category_id"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
	Type        string  `json:"type"` // "income" or "expense"
}

// CreateTransaction creates a new transaction
func (s *Server) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("id").(int)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	var transactionRequest TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&transactionRequest); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	var transaction domains.Transaction
	transaction.UserID = userID
	transaction.CategoryID = transactionRequest.CategoryID
	transaction.Amount = transactionRequest.Amount
	transaction.Description = transactionRequest.Description
	transaction.Type = transactionRequest.Type

	// Parse and set the created_at field
	if transactionRequest.CreatedAt != nil {
		createdAt, err := time.Parse("2006-01-02", *transactionRequest.CreatedAt)
		if err != nil {
			http.Error(w, "Invalid date format", http.StatusBadRequest)
			return
		}
		transaction.CreatedAt = &createdAt
	}

	if err := s.DB.Create(&transaction).Error; err != nil {
		http.Error(w, "Failed to create transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(transaction)
}

// GetTransaction retrieves a transaction by ID
func (s *Server) GetTransaction(w http.ResponseWriter, r *http.Request) {

	userID, ok := r.Context().Value("id").(int)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid transaction ID", http.StatusBadRequest)
		return
	}

	var transaction domains.Transaction
	if err := s.DB.Where("user_id = ?", userID).First(&transaction, id).Error; err != nil {
		http.Error(w, "Transaction not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(transaction)
}

// GetAllTransactions retrieves all transactions
func (s *Server) GetAllTransactions(w http.ResponseWriter, r *http.Request) {

	userID, ok := r.Context().Value("id").(int)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	// Parse query parameters for pagination
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	// Parse query parameter for type
	transactionType := r.URL.Query().Get("type")

	// Parse query parameters for start and end time
	start := r.URL.Query().Get("start")
	end := r.URL.Query().Get("end")

	categoryID, err := strconv.Atoi(r.URL.Query().Get("search"))
	if err != nil {
		categoryID = 0
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	// Retrieve transactions with pagination and optional filters
	var transactions []domains.Transaction
	query := s.DB.Limit(pageSize).Offset(offset)
	if transactionType != "" {
		query = query.Where("type = ?", transactionType)
	}
	if start != "" && end != "" {
		query = query.Where("created_at BETWEEN ? AND ?", start, end)
	}

	if categoryID != 0 {
		query = query.Where("category_id = ?", categoryID)
	}

	if err := query.Where("user_id = ?", userID).Find(&transactions).Error; err != nil {
		http.Error(w, "Failed to retrieve transactions", http.StatusInternalServerError)
		return
	}

	// Get total count of transactions
	var total int64
	countQuery := s.DB.Model(&domains.Transaction{}).Where("user_id = ?", userID)
	if transactionType != "" {
		countQuery = countQuery.Where("type = ?", transactionType)
	}
	if start != "" && end != "" {
		countQuery = countQuery.Where("created_at BETWEEN ? AND ?", start, end)
	}
	if categoryID != 0 {
		countQuery = countQuery.Where("category_id = ?", categoryID)
	}
	if err := countQuery.Count(&total).Error; err != nil {
		http.Error(w, "Failed to count transactions", http.StatusInternalServerError)
		return
	}

	// Calculate paging information
	pageCount := (total + int64(pageSize) - 1) / int64(pageSize)
	canNext := page < int(pageCount)
	canPre := page > 1

	// Prepare the response
	response := map[string]interface{}{
		"paging": map[string]interface{}{
			"total":      total,
			"page":       page,
			"page_count": pageCount,
			"can_next":   canNext,
			"can_pre":    canPre,
		},
		"transaction_data": transactions,
	}

	// Encode the response to JSON and send it
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// UpdateTransaction updates a transaction by ID
func (s *Server) UpdateTransaction(w http.ResponseWriter, r *http.Request) {

	userID, ok := r.Context().Value("id").(int)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid transaction ID", http.StatusBadRequest)
		return
	}

	var transaction domains.Transaction
	if err := s.DB.Where("user_id = ? AND id = ?", userID, id).First(&transaction).Error; err != nil {
		http.Error(w, "Transaction not found", http.StatusNotFound)
		return
	}

	var transactionRequest TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&transactionRequest); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	transaction.CategoryID = transactionRequest.CategoryID
	transaction.Amount = transactionRequest.Amount
	transaction.Description = transactionRequest.Description
	transaction.Type = transactionRequest.Type

	if err := s.DB.Save(&transaction).Error; err != nil {
		http.Error(w, "Failed to update transaction", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(transaction)
}

// DeleteTransaction deletes a transaction by ID
func (s *Server) DeleteTransaction(w http.ResponseWriter, r *http.Request) {

	userID, ok := r.Context().Value("id").(int)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid transaction ID", http.StatusBadRequest)
		return
	}

	if err := s.DB.Where("user_id = ?", userID).Delete(&domains.Transaction{}, id).Error; err != nil {
		http.Error(w, "Failed to delete transaction", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode("Transaction deleted")
}

// GetBalance calculates the total balance from income and expense within a time range
func (s *Server) GetBalance(w http.ResponseWriter, r *http.Request) {

	userID, ok := r.Context().Value("id").(int)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	start := r.URL.Query().Get("start")
	end := r.URL.Query().Get("end")

	var totalIncome, totalExpense float64

	// Calculate total income
	if err := s.DB.Model(&domains.Transaction{}).
		Where("user_id = ? AND type = ? AND created_at BETWEEN ? AND ?", userID, "income", start, end).
		Select("SUM(amount)").Scan(&totalIncome).Error; err != nil {
		http.Error(w, "Failed to calculate total income", http.StatusInternalServerError)
		return
	}

	// Calculate total expense
	if err := s.DB.Model(&domains.Transaction{}).
		Where("user_id = ? AND type = ? AND created_at BETWEEN ? AND ?", userID, "expense", start, end).
		Select("SUM(amount)").Scan(&totalExpense).Error; err != nil {
		http.Error(w, "Failed to calculate total expense", http.StatusInternalServerError)
		return
	}

	// Calculate balance
	balance := totalIncome - totalExpense

	json.NewEncoder(w).Encode(map[string]float64{"balance": balance})
}

// GetTotalIncome calculates the total income within a time range
func (s *Server) GetTotalIncome(w http.ResponseWriter, r *http.Request) {

	userID, ok := r.Context().Value("id").(int)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	start := r.URL.Query().Get("start")
	end := r.URL.Query().Get("end")

	var totalIncome float64

	if err := s.DB.Model(&domains.Transaction{}).
		Where("user_id = ? AND type = ? AND created_at BETWEEN ? AND ?", userID, "income", start, end).
		Select("SUM(amount)").Scan(&totalIncome).Error; err != nil {
		http.Error(w, "Failed to calculate total income", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]float64{"total_income": totalIncome})
}

// GetTotalExpense calculates the total expense within a time range
func (s *Server) GetTotalExpense(w http.ResponseWriter, r *http.Request) {

	userID, ok := r.Context().Value("id").(int)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	start := r.URL.Query().Get("start")
	end := r.URL.Query().Get("end")

	var totalExpense float64

	if err := s.DB.Model(&domains.Transaction{}).
		Where("user_id = ? AND type = ? AND created_at BETWEEN ? AND ?", userID, "expense", start, end).
		Select("SUM(amount)").Scan(&totalExpense).Error; err != nil {
		http.Error(w, "Failed to calculate total expense", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]float64{"total_expense": totalExpense})
}