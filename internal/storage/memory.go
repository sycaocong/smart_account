package storage

import (
	"errors"
	"sync"

	"github.com/smartx/account/internal/types"
)

var (
	ErrNotFound = errors.New("not found")
)

// MemoryStorage 内存数据库实现
// 用于开发和测试环境
type MemoryStorage struct {
	mu           sync.RWMutex
	accounts     map[string]*types.Account
	accountsByEmail map[string]*types.Account
	accountsByPhone map[string]*types.Account
	balances     map[string]*types.Balance
	transactions map[string]*types.Transaction
	apiKeys      map[string]*types.APIKey
	apiKeysByKey map[string]*types.APIKey
	riskLimits   map[string]*types.RiskLimit
}

// NewMemoryStorage 创建内存数据库
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		accounts:        make(map[string]*types.Account),
		accountsByEmail: make(map[string]*types.Account),
		accountsByPhone: make(map[string]*types.Account),
		balances:        make(map[string]*types.Balance),
		transactions:    make(map[string]*types.Transaction),
		apiKeys:         make(map[string]*types.APIKey),
		apiKeysByKey:    make(map[string]*types.APIKey),
		riskLimits:      make(map[string]*types.RiskLimit),
	}
}

// 账户操作
func (s *MemoryStorage) CreateAccount(account *types.Account) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.accounts[account.AccountID]; ok {
		return errors.New("account already exists")
	}

	s.accounts[account.AccountID] = account
	if account.Email != "" {
		s.accountsByEmail[account.Email] = account
	}
	if account.Phone != "" {
		s.accountsByPhone[account.Phone] = account
	}

	return nil
}

func (s *MemoryStorage) GetAccount(accountID string) (*types.Account, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	account, ok := s.accounts[accountID]
	if !ok {
		return nil, ErrNotFound
	}
	return account, nil
}

func (s *MemoryStorage) GetAccountByEmail(email string) (*types.Account, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	account, ok := s.accountsByEmail[email]
	if !ok {
		return nil, ErrNotFound
	}
	return account, nil
}

func (s *MemoryStorage) GetAccountByPhone(phone string) (*types.Account, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	account, ok := s.accountsByPhone[phone]
	if !ok {
		return nil, ErrNotFound
	}
	return account, nil
}

func (s *MemoryStorage) UpdateAccount(account *types.Account) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.accounts[account.AccountID]; !ok {
		return ErrNotFound
	}

	if account.Email != "" {
		s.accountsByEmail[account.Email] = account
	}
	if account.Phone != "" {
		s.accountsByPhone[account.Phone] = account
	}

	s.accounts[account.AccountID] = account
	return nil
}

func (s *MemoryStorage) ListSubAccounts(parentID string) ([]*types.Account, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*types.Account
	for _, account := range s.accounts {
		if account.ParentAccountID == parentID {
			result = append(result, account)
		}
	}
	return result, nil
}

// 余额操作
func (s *MemoryStorage) CreateBalance(balance *types.Balance) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := balance.AccountID + ":" + balance.Currency
	if _, ok := s.balances[key]; ok {
		return errors.New("balance already exists")
	}

	s.balances[key] = balance
	return nil
}

func (s *MemoryStorage) GetBalance(accountID, currency string) (*types.Balance, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := accountID + ":" + currency
	balance, ok := s.balances[key]
	if !ok {
		return nil, ErrNotFound
	}
	return balance, nil
}

func (s *MemoryStorage) ListBalances(accountID string) ([]*types.Balance, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*types.Balance
	for _, balance := range s.balances {
		if balance.AccountID == accountID {
			result = append(result, balance)
		}
	}
	return result, nil
}

func (s *MemoryStorage) UpdateBalance(balance *types.Balance) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := balance.AccountID + ":" + balance.Currency
	if _, ok := s.balances[key]; !ok {
		return ErrNotFound
	}

	s.balances[key] = balance
	return nil
}

// 流水操作
func (s *MemoryStorage) CreateTransaction(tx *types.Transaction) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.transactions[tx.TransactionID] = tx
	return nil
}

func (s *MemoryStorage) GetTransaction(txID string) (*types.Transaction, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tx, ok := s.transactions[txID]
	if !ok {
		return nil, ErrNotFound
	}
	return tx, nil
}

func (s *MemoryStorage) ListTransactions(accountID string, txType *types.TransactionType, limit, offset int) ([]*types.Transaction, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*types.Transaction
	for _, tx := range s.transactions {
		if tx.AccountID != accountID {
			continue
		}
		if txType != nil && *txType != tx.Type {
			continue
		}
		result = append(result, tx)
	}

	if offset >= len(result) {
		return []*types.Transaction{}, nil
	}

	end := offset + limit
	if end > len(result) {
		end = len(result)
	}

	return result[offset:end], nil
}

// API Key 操作
func (s *MemoryStorage) CreateAPIKey(apiKey *types.APIKey) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.apiKeys[apiKey.APIKeyID]; ok {
		return errors.New("api key already exists")
	}

	s.apiKeys[apiKey.APIKeyID] = apiKey
	s.apiKeysByKey[apiKey.Key] = apiKey
	return nil
}

func (s *MemoryStorage) GetAPIKey(apiKeyID string) (*types.APIKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	apiKey, ok := s.apiKeys[apiKeyID]
	if !ok {
		return nil, ErrNotFound
	}
	return apiKey, nil
}

func (s *MemoryStorage) GetAPIKeyByKey(key string) (*types.APIKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	apiKey, ok := s.apiKeysByKey[key]
	if !ok {
		return nil, ErrNotFound
	}
	return apiKey, nil
}

func (s *MemoryStorage) ListAPIKeys(accountID string) ([]*types.APIKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*types.APIKey
	for _, apiKey := range s.apiKeys {
		if apiKey.AccountID == accountID {
			result = append(result, apiKey)
		}
	}
	return result, nil
}

func (s *MemoryStorage) UpdateAPIKey(apiKey *types.APIKey) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.apiKeys[apiKey.APIKeyID]; !ok {
		return ErrNotFound
	}

	s.apiKeys[apiKey.APIKeyID] = apiKey
	s.apiKeysByKey[apiKey.Key] = apiKey
	return nil
}

func (s *MemoryStorage) DeleteAPIKey(apiKeyID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	apiKey, ok := s.apiKeys[apiKeyID]
	if !ok {
		return ErrNotFound
	}

	delete(s.apiKeys, apiKeyID)
	delete(s.apiKeysByKey, apiKey.Key)
	return nil
}

// 风控限额操作
func (s *MemoryStorage) CreateRiskLimit(limit *types.RiskLimit) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.riskLimits[limit.LimitID] = limit
	return nil
}

func (s *MemoryStorage) GetRiskLimit(accountID, currency string, limitType types.LimitType) (*types.RiskLimit, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, limit := range s.riskLimits {
		if limit.AccountID == accountID && limit.Currency == currency && limit.LimitType == limitType {
			return limit, nil
		}
	}
	return nil, ErrNotFound
}

func (s *MemoryStorage) UpdateRiskLimit(limit *types.RiskLimit) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.riskLimits[limit.LimitID]; !ok {
		return ErrNotFound
	}

	s.riskLimits[limit.LimitID] = limit
	return nil
}

func (s *MemoryStorage) DeleteRiskLimit(limitID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.riskLimits, limitID)
	return nil
}