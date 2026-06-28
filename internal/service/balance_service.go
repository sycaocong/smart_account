package service

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/smartx/account/internal/types"
	"github.com/smartx/account/internal/storage"
)

var (
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrBalanceNotFound     = errors.New("balance not found")
)

// BalanceService 余额服务
// [Design: BalanceService 余额服务](../DESIGN_ACCOUNT.md#32-balanceservice-余额服务)
type BalanceService struct {
	db             storage.Database
	txService      *TransactionService
	balanceLocks   sync.Map
}

// NewBalanceService 创建余额服务
func NewBalanceService(db storage.Database, txService *TransactionService) *BalanceService {
	return &BalanceService{
		db:          db,
		txService:   txService,
	}
}

// InitializeBalance 初始化余额
func (s *BalanceService) InitializeBalance(accountID, currency string) error {
	balance := &types.Balance{
		BalanceID: uuid.New().String(),
		AccountID: accountID,
		Currency:  currency,
		Available: 0,
		Frozen:    0,
		Total:     0,
		UpdatedAt: time.Now(),
	}
	return s.db.CreateBalance(balance)
}

// GetBalance 获取余额
// [Design: BalanceService 余额服务](../DESIGN_ACCOUNT.md#32-balanceservice-余额服务)
func (s *BalanceService) GetBalance(accountID, currency string) (*types.Balance, error) {
	balance, err := s.db.GetBalance(accountID, currency)
	if err != nil {
		return nil, ErrBalanceNotFound
	}
	return balance, nil
}

// GetAllBalances 获取所有余额
// [Design: BalanceService 余额服务](../DESIGN_ACCOUNT.md#32-balanceservice-余额服务)
func (s *BalanceService) GetAllBalances(accountID string) ([]*types.Balance, error) {
	return s.db.ListBalances(accountID)
}

// acquireLock 获取余额锁
func (s *BalanceService) acquireLock(accountID, currency string) func() {
	key := accountID + ":" + currency
	mu, _ := s.balanceLocks.LoadOrStore(key, &sync.Mutex{})
	mutex := mu.(*sync.Mutex)
	mutex.Lock()
	return func() { mutex.Unlock() }
}

// Frozen 冻结余额
// [Design: BalanceService 余额服务](../DESIGN_ACCOUNT.md#32-balanceservice-余额服务)
func (s *BalanceService) Frozen(accountID, currency string, amount float64) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	unlock := s.acquireLock(accountID, currency)
	defer unlock()

	balance, err := s.GetBalance(accountID, currency)
	if err != nil {
		return err
	}

	if balance.Available < amount {
		return ErrInsufficientBalance
	}

	balance.Available -= amount
	balance.Frozen += amount
	balance.Total = balance.Available + balance.Frozen
	balance.UpdatedAt = time.Now()

	return s.db.UpdateBalance(balance)
}

// Unfrozen 解冻余额
// [Design: BalanceService 余额服务](../DESIGN_ACCOUNT.md#32-balanceservice-余额服务)
func (s *BalanceService) Unfrozen(accountID, currency string, amount float64) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	unlock := s.acquireLock(accountID, currency)
	defer unlock()

	balance, err := s.GetBalance(accountID, currency)
	if err != nil {
		return err
	}

	if balance.Frozen < amount {
		return errors.New("frozen balance insufficient")
	}

	balance.Frozen -= amount
	balance.Available += amount
	balance.Total = balance.Available + balance.Frozen
	balance.UpdatedAt = time.Now()

	return s.db.UpdateBalance(balance)
}

// Deposit 充值
// [Design: BalanceService 余额服务](../DESIGN_ACCOUNT.md#32-balanceservice-余额服务)
func (s *BalanceService) Deposit(accountID, currency string, amount float64, txID, memo string) (*types.Transaction, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be positive")
	}

	unlock := s.acquireLock(accountID, currency)
	defer unlock()

	balance, err := s.GetBalance(accountID, currency)
	if err != nil {
		return nil, err
	}

	balanceBefore := balance.Total

	balance.Available += amount
	balance.Total = balance.Available + balance.Frozen
	balance.UpdatedAt = time.Now()

	if err := s.db.UpdateBalance(balance); err != nil {
		return nil, err
	}

	tx := &types.Transaction{
		TransactionID:  uuid.New().String(),
		AccountID:      accountID,
		Currency:       currency,
		Type:           types.TxDeposit,
		Amount:         amount,
		BalanceBefore:  balanceBefore,
		BalanceAfter:   balance.Total,
		RelatedTxID:    txID,
		Memo:           memo,
		CreatedAt:      time.Now(),
	}

	return s.txService.CreateTransaction(tx)
}

// Withdraw 提现
// [Design: BalanceService 余额服务](../DESIGN_ACCOUNT.md#32-balanceservice-余额服务)
func (s *BalanceService) Withdraw(accountID, currency string, amount float64, address, memo string) (*types.Transaction, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be positive")
	}

	if address == "" {
		return nil, errors.New("address is required")
	}

	unlock := s.acquireLock(accountID, currency)
	defer unlock()

	balance, err := s.GetBalance(accountID, currency)
	if err != nil {
		return nil, err
	}

	if balance.Available < amount {
		return nil, ErrInsufficientBalance
	}

	balanceBefore := balance.Total

	balance.Available -= amount
	balance.Total = balance.Available + balance.Frozen
	balance.UpdatedAt = time.Now()

	if err := s.db.UpdateBalance(balance); err != nil {
		return nil, err
	}

	tx := &types.Transaction{
		TransactionID:  uuid.New().String(),
		AccountID:      accountID,
		Currency:       currency,
		Type:           types.TxWithdraw,
		Amount:         -amount,
		BalanceBefore:  balanceBefore,
		BalanceAfter:   balance.Total,
		Memo:           memo,
		CreatedAt:      time.Now(),
	}

	return s.txService.CreateTransaction(tx)
}

// Transfer 划转
// [Design: BalanceService 余额服务](../DESIGN_ACCOUNT.md#32-balanceservice-余额服务)
func (s *BalanceService) Transfer(fromID, toID, currency string, amount float64, memo string) (*types.Transaction, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be positive")
	}

	if fromID == toID {
		return nil, errors.New("cannot transfer to self")
	}

	fromLock := s.acquireLock(fromID, currency)
	toLock := s.acquireLock(toID, currency)
	defer func() {
		fromLock()
		toLock()
	}()

	fromBalance, err := s.GetBalance(fromID, currency)
	if err != nil {
		return nil, err
	}

	toBalance, err := s.GetBalance(toID, currency)
	if err != nil {
		return nil, err
	}

	if fromBalance.Available < amount {
		return nil, ErrInsufficientBalance
	}

	fromBalanceBefore := fromBalance.Total

	fromBalance.Available -= amount
	fromBalance.Total = fromBalance.Available + fromBalance.Frozen
	fromBalance.UpdatedAt = time.Now()

	toBalance.Available += amount
	toBalance.Total = toBalance.Available + toBalance.Frozen
	toBalance.UpdatedAt = time.Now()

	if err := s.db.UpdateBalance(fromBalance); err != nil {
		return nil, err
	}

	if err := s.db.UpdateBalance(toBalance); err != nil {
		return nil, err
	}

	tx := &types.Transaction{
		TransactionID:  uuid.New().String(),
		AccountID:      fromID,
		Currency:       currency,
		Type:           types.TxTransfer,
		Amount:         -amount,
		BalanceBefore:  fromBalanceBefore,
		BalanceAfter:   fromBalance.Total,
		Memo:           memo,
		CreatedAt:      time.Now(),
	}

	return s.txService.CreateTransaction(tx)
}

// DeductFee 扣除手续费
func (s *BalanceService) DeductFee(accountID, currency string, amount float64) (*types.Transaction, error) {
	if amount <= 0 {
		return nil, errors.New("fee amount must be positive")
	}

	unlock := s.acquireLock(accountID, currency)
	defer unlock()

	balance, err := s.GetBalance(accountID, currency)
	if err != nil {
		return nil, err
	}

	if balance.Available < amount {
		return nil, ErrInsufficientBalance
	}

	balanceBefore := balance.Total

	balance.Available -= amount
	balance.Total = balance.Available + balance.Frozen
	balance.UpdatedAt = time.Now()

	if err := s.db.UpdateBalance(balance); err != nil {
		return nil, err
	}

	tx := &types.Transaction{
		TransactionID:  uuid.New().String(),
		AccountID:      accountID,
		Currency:       currency,
		Type:           types.TxFee,
		Amount:         -amount,
		BalanceBefore:  balanceBefore,
		BalanceAfter:   balance.Total,
		Memo:           "trading fee",
		CreatedAt:      time.Now(),
	}

	return s.txService.CreateTransaction(tx)
}