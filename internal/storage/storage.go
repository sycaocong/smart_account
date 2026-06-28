package storage

import (
	"github.com/smartx/account/internal/types"
)

// Database 数据库接口
// [Design: 数据存储层](../DESIGN_ACCOUNT.md#13-系统架构)
type Database interface {
	// 账户操作
	CreateAccount(account *types.Account) error
	GetAccount(accountID string) (*types.Account, error)
	GetAccountByEmail(email string) (*types.Account, error)
	GetAccountByPhone(phone string) (*types.Account, error)
	UpdateAccount(account *types.Account) error
	ListSubAccounts(parentID string) ([]*types.Account, error)

	// 余额操作
	CreateBalance(balance *types.Balance) error
	GetBalance(accountID, currency string) (*types.Balance, error)
	ListBalances(accountID string) ([]*types.Balance, error)
	UpdateBalance(balance *types.Balance) error

	// 流水操作
	CreateTransaction(tx *types.Transaction) error
	GetTransaction(txID string) (*types.Transaction, error)
	ListTransactions(accountID string, txType *types.TransactionType, limit, offset int) ([]*types.Transaction, error)

	// API Key 操作
	CreateAPIKey(apiKey *types.APIKey) error
	GetAPIKey(apiKeyID string) (*types.APIKey, error)
	GetAPIKeyByKey(key string) (*types.APIKey, error)
	ListAPIKeys(accountID string) ([]*types.APIKey, error)
	UpdateAPIKey(apiKey *types.APIKey) error
	DeleteAPIKey(apiKeyID string) error

	// 风控限额操作
	CreateRiskLimit(limit *types.RiskLimit) error
	GetRiskLimit(accountID, currency string, limitType types.LimitType) (*types.RiskLimit, error)
	UpdateRiskLimit(limit *types.RiskLimit) error
	DeleteRiskLimit(limitID string) error
}