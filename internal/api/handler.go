package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/rs/zerolog"

	"github.com/smartx/account/internal/service"
	"github.com/smartx/account/internal/types"
)

type Handler struct {
	accountService     *service.AccountService
	balanceService     *service.BalanceService
	transactionService *service.TransactionService
	apiKeyService      *service.APIKeyService
	logger             zerolog.Logger
	server             *http.Server
}

func NewHandler(accountService *service.AccountService, balanceService *service.BalanceService,
	transactionService *service.TransactionService, apiKeyService *service.APIKeyService,
	logger zerolog.Logger) *Handler {

	h := &Handler{
		accountService:     accountService,
		balanceService:     balanceService,
		transactionService: transactionService,
		apiKeyService:      apiKeyService,
		logger:             logger,
	}

	mux := http.NewServeMux()
	h.setupRoutes(mux)

	h.server = &http.Server{
		Handler: mux,
	}

	return h
}

func (h *Handler) Handler() http.Handler {
	return h.server.Handler
}

func (h *Handler) setupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", h.Health)
	mux.HandleFunc("/api/v1/login", h.handleLogin)
	mux.HandleFunc("/api/v1/accounts", h.handleAccounts)
	mux.HandleFunc("/api/v1/accounts/", h.handleAccountWithPath)
	mux.HandleFunc("/api/v1/transactions/", h.handleTransaction)
	mux.HandleFunc("/api/v1/api-keys/", h.handleAPIKey)
}

func (h *Handler) Start(addr string) error {
	h.server.Addr = addr
	h.logger.Info().Str("addr", addr).Msg("Account service starting")
	return h.server.ListenAndServe()
}

func (h *Handler) handleAccounts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		h.CreateAccount(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handleAccountWithPath(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/accounts/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) == 0 {
		http.Error(w, "Account ID required", http.StatusBadRequest)
		return
	}

	accountID := parts[0]

	if len(parts) == 1 {
		h.handleAccount(w, r, accountID)
		return
	}

	subPath := parts[1]
	switch subPath {
	case "balances":
		h.handleAccountBalances(w, r, accountID)
	case "deposit":
		h.handleDeposit(w, r, accountID)
	case "withdraw":
		h.handleWithdraw(w, r, accountID)
	case "transfer":
		h.handleTransfer(w, r, accountID)
	case "transactions":
		h.handleTransactions(w, r, accountID)
	case "api-keys":
		h.handleAccountAPIKeys(w, r, accountID)
	default:
		if strings.HasPrefix(subPath, "balances/") {
			currency := strings.TrimPrefix(subPath, "balances/")
			h.handleAccountBalance(w, r, accountID, currency)
		} else {
			http.Error(w, "Path not found", http.StatusNotFound)
		}
	}
}

func (h *Handler) handleAccount(w http.ResponseWriter, r *http.Request, accountID string) {
	switch r.Method {
	case "GET":
		h.GetAccount(w, r, accountID)
	case "PUT":
		h.UpdateAccount(w, r, accountID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var req types.CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	account, err := h.accountService.CreateAccount(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(account)
}

func (h *Handler) GetAccount(w http.ResponseWriter, r *http.Request, accountID string) {
	account, err := h.accountService.GetAccount(accountID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(account)
}

func (h *Handler) UpdateAccount(w http.ResponseWriter, r *http.Request, accountID string) {
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	account, err := h.accountService.UpdateAccount(accountID, updates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(account)
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	account, err := h.accountService.Login(req.Email, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(account)
}

func (h *Handler) handleAccountBalances(w http.ResponseWriter, r *http.Request, accountID string) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	balances, err := h.balanceService.GetAllBalances(accountID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"balances": balances,
	})
}

func (h *Handler) handleAccountBalance(w http.ResponseWriter, r *http.Request, accountID, currency string) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	balance, err := h.balanceService.GetBalance(accountID, currency)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(balance)
}

func (h *Handler) handleDeposit(w http.ResponseWriter, r *http.Request, accountID string) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req types.DepositRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tx, err := h.balanceService.Deposit(accountID, req.Currency, req.Amount, req.TxID, req.Memo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tx)
}

func (h *Handler) handleWithdraw(w http.ResponseWriter, r *http.Request, accountID string) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req types.WithdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tx, err := h.balanceService.Withdraw(accountID, req.Currency, req.Amount, req.Address, req.Memo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tx)
}

func (h *Handler) handleTransfer(w http.ResponseWriter, r *http.Request, accountID string) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req types.TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tx, err := h.balanceService.Transfer(accountID, req.ToAccountID, req.Currency, req.Amount, req.Memo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tx)
}

func (h *Handler) handleTransactions(w http.ResponseWriter, r *http.Request, accountID string) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	limit := 100
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil {
			offset = parsed
		}
	}

	txTypeStr := r.URL.Query().Get("type")
	var txType *types.TransactionType
	if txTypeStr != "" {
		if parsed, err := strconv.Atoi(txTypeStr); err == nil {
			t := types.TransactionType(parsed)
			txType = &t
		}
	}

	transactions, err := h.transactionService.ListTransactions(accountID, txType, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"transactions": transactions,
		"limit":        limit,
		"offset":       offset,
	})
}

func (h *Handler) handleTransaction(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/transactions/")
	if len(path) == 0 {
		http.Error(w, "Transaction ID required", http.StatusBadRequest)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tx, err := h.transactionService.GetTransaction(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tx)
}

func (h *Handler) handleAccountAPIKeys(w http.ResponseWriter, r *http.Request, accountID string) {
	switch r.Method {
	case "GET":
		h.handleListAPIKeys(w, r, accountID)
	case "POST":
		h.handleCreateAPIKey(w, r, accountID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handleCreateAPIKey(w http.ResponseWriter, r *http.Request, accountID string) {
	var req types.CreateAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	apiKey, err := h.apiKeyService.CreateAPIKey(accountID, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(apiKey)
}

func (h *Handler) handleListAPIKeys(w http.ResponseWriter, r *http.Request, accountID string) {
	apiKeys, err := h.apiKeyService.ListAPIKeys(accountID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(apiKeys)
}

func (h *Handler) handleAPIKey(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/api-keys/")
	if len(path) == 0 {
		http.Error(w, "API Key ID required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "PUT":
		h.UpdateAPIKey(w, r, path)
	case "DELETE":
		h.DeleteAPIKey(w, r, path)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) UpdateAPIKey(w http.ResponseWriter, r *http.Request, apiKeyID string) {
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	apiKey, err := h.apiKeyService.UpdateAPIKey(apiKeyID, updates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(apiKey)
}

func (h *Handler) DeleteAPIKey(w http.ResponseWriter, r *http.Request, apiKeyID string) {
	if err := h.apiKeyService.DeleteAPIKey(apiKeyID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}