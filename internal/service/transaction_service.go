package service

import (
	"github.com/smartx/account/internal/types"
	"github.com/smartx/account/internal/storage"
)

// TransactionService 流水服务
// [Design: TransactionService 流水服务](../DESIGN_ACCOUNT.md#33-transactionservice-流水服务)
type TransactionService struct {
	db storage.Database
}

// NewTransactionService 创建流水服务
func NewTransactionService(db storage.Database) *TransactionService {
	return &TransactionService{
		db: db,
	}
}

// GetTransaction 获取流水
// [Design: TransactionService 流水服务](../DESIGN_ACCOUNT.md#33-transactionservice-流水服务)
func (s *TransactionService) GetTransaction(txID string) (*types.Transaction, error) {
	return s.db.GetTransaction(txID)
}

// ListTransactions 列出流水
// [Design: TransactionService 流水服务](../DESIGN_ACCOUNT.md#33-transactionservice-流水服务)
func (s *TransactionService) ListTransactions(accountID string, txType *types.TransactionType, limit, offset int) ([]*types.Transaction, error) {
	return s.db.ListTransactions(accountID, txType, limit, offset)
}

// CreateTransaction 创建流水
// [Design: TransactionService 流水服务](../DESIGN_ACCOUNT.md#33-transactionservice-流水服务)
func (s *TransactionService) CreateTransaction(tx *types.Transaction) (*types.Transaction, error) {
	if err := s.db.CreateTransaction(tx); err != nil {
		return nil, err
	}
	return tx, nil
}