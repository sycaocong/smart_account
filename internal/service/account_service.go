package service

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/smartx/account/internal/types"
	"github.com/smartx/account/internal/storage"
)

var (
	ErrAccountNotFound   = errors.New("account not found")
	ErrAccountAlreadyExist = errors.New("account already exists")
	ErrAccountFrozen     = errors.New("account is frozen")
	ErrInvalidPassword   = errors.New("invalid password")
)

// AccountService 账户服务
// [Design: AccountService 账户服务](../DESIGN_ACCOUNT.md#31-accountservice-账户服务)
type AccountService struct {
	db        storage.Database
	balanceSvc *BalanceService
}

// NewAccountService 创建账户服务
func NewAccountService(db storage.Database, balanceSvc *BalanceService) *AccountService {
	return &AccountService{
		db:        db,
		balanceSvc: balanceSvc,
	}
}

// CreateAccount 创建账户
// [Design: AccountService 账户服务](../DESIGN_ACCOUNT.md#31-accountservice-账户服务)
func (s *AccountService) CreateAccount(req *types.CreateAccountRequest) (*types.Account, error) {
	if req.Email == "" && req.Phone == "" {
		return nil, errors.New("email or phone is required")
	}

	if req.Password == "" {
		return nil, errors.New("password is required")
	}

	if req.Email != "" {
		_, err := s.db.GetAccountByEmail(req.Email)
		if err == nil {
			return nil, ErrAccountAlreadyExist
		}
	}

	if req.Phone != "" {
		_, err := s.db.GetAccountByPhone(req.Phone)
		if err == nil {
			return nil, ErrAccountAlreadyExist
		}
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	account := &types.Account{
		AccountID:  uuid.New().String(),
		UserID:     uuid.New().String(),
		Email:      req.Email,
		Phone:      req.Phone,
		PasswordHash: string(passwordHash),
		Status:     types.AccountActive,
		Tier:       0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.db.CreateAccount(account); err != nil {
		return nil, err
	}

	defaultCurrencies := []string{"BTC", "ETH", "USDT"}
	for _, currency := range defaultCurrencies {
		if err := s.balanceSvc.InitializeBalance(account.AccountID, currency); err != nil {
			return nil, err
		}
	}

	return account, nil
}

// GetAccount 获取账户
// [Design: AccountService 账户服务](../DESIGN_ACCOUNT.md#31-accountservice-账户服务)
func (s *AccountService) GetAccount(accountID string) (*types.Account, error) {
	account, err := s.db.GetAccount(accountID)
	if err != nil {
		return nil, ErrAccountNotFound
	}
	return account, nil
}

// GetAccountByEmail 通过邮箱获取账户
func (s *AccountService) GetAccountByEmail(email string) (*types.Account, error) {
	return s.db.GetAccountByEmail(email)
}

// UpdateAccount 更新账户
// [Design: AccountService 账户服务](../DESIGN_ACCOUNT.md#31-accountservice-账户服务)
func (s *AccountService) UpdateAccount(accountID string, updates map[string]interface{}) (*types.Account, error) {
	account, err := s.GetAccount(accountID)
	if err != nil {
		return nil, err
	}

	if email, ok := updates["email"].(string); ok && email != "" {
		existing, _ := s.db.GetAccountByEmail(email)
		if existing != nil && existing.AccountID != accountID {
			return nil, ErrAccountAlreadyExist
		}
		account.Email = email
	}

	if phone, ok := updates["phone"].(string); ok && phone != "" {
		existing, _ := s.db.GetAccountByPhone(phone)
		if existing != nil && existing.AccountID != accountID {
			return nil, ErrAccountAlreadyExist
		}
		account.Phone = phone
	}

	if tier, ok := updates["tier"].(int); ok {
		account.Tier = tier
	}

	account.UpdatedAt = time.Now()

	if err := s.db.UpdateAccount(account); err != nil {
		return nil, err
	}

	return account, nil
}

// Login 登录
func (s *AccountService) Login(email, password string) (*types.Account, error) {
	account, err := s.db.GetAccountByEmail(email)
	if err != nil {
		return nil, ErrAccountNotFound
	}

	if account.Status == types.AccountLocked || account.Status == types.AccountClosed {
		return nil, ErrAccountFrozen
	}

	if err := bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidPassword
	}

	account.LastLoginAt = time.Now()
	s.db.UpdateAccount(account)

	return account, nil
}

// FrozenAccount 冻结账户
// [Design: AccountService 账户服务](../DESIGN_ACCOUNT.md#31-accountservice-账户服务)
func (s *AccountService) FrozenAccount(accountID string, reason string) error {
	account, err := s.GetAccount(accountID)
	if err != nil {
		return err
	}

	account.Status = types.AccountFrozen
	account.UpdatedAt = time.Now()

	return s.db.UpdateAccount(account)
}

// UnlockAccount 解冻账户
// [Design: AccountService 账户服务](../DESIGN_ACCOUNT.md#31-accountservice-账户服务)
func (s *AccountService) UnlockAccount(accountID string) error {
	account, err := s.GetAccount(accountID)
	if err != nil {
		return err
	}

	account.Status = types.AccountActive
	account.UpdatedAt = time.Now()

	return s.db.UpdateAccount(account)
}

// CreateSubAccount 创建子账户
// [Design: AccountService 账户服务](../DESIGN_ACCOUNT.md#31-accountservice-账户服务)
func (s *AccountService) CreateSubAccount(parentID, email string) (*types.Account, error) {
	parent, err := s.GetAccount(parentID)
	if err != nil {
		return nil, err
	}

	if parent.Status != types.AccountActive {
		return nil, ErrAccountFrozen
	}

	if email != "" {
		_, err := s.db.GetAccountByEmail(email)
		if err == nil {
			return nil, ErrAccountAlreadyExist
		}
	}

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte(uuid.New().String()), bcrypt.DefaultCost)

	account := &types.Account{
		AccountID:        uuid.New().String(),
		UserID:           uuid.New().String(),
		Email:            email,
		PasswordHash:     string(passwordHash),
		Status:           types.AccountActive,
		Tier:             parent.Tier,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		ParentAccountID:  parentID,
	}

	if err := s.db.CreateAccount(account); err != nil {
		return nil, err
	}

	defaultCurrencies := []string{"BTC", "ETH", "USDT"}
	for _, currency := range defaultCurrencies {
		s.balanceSvc.InitializeBalance(account.AccountID, currency)
	}

	return account, nil
}

// ListSubAccounts 列出子账户
// [Design: AccountService 账户服务](../DESIGN_ACCOUNT.md#31-accountservice-账户服务)
func (s *AccountService) ListSubAccounts(parentID string) ([]*types.Account, error) {
	_, err := s.GetAccount(parentID)
	if err != nil {
		return nil, err
	}

	return s.db.ListSubAccounts(parentID)
}