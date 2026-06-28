package types

import "time"

// AccountStatus 账户状态枚举
// [Design: 账户状态枚举](../DESIGN_ACCOUNT.md#21-账户-account)
type AccountStatus int

const (
	AccountActive   AccountStatus = iota // 正常
	AccountFrozen                        // 冻结
	AccountLocked                        // 锁定
	AccountClosed                        // 已关闭
)

// TransactionType 资金流水类型枚举
// [Design: 资金流水-transaction](../DESIGN_ACCOUNT.md#23-资金流水-transaction)
type TransactionType int

const (
	TxDeposit      TransactionType = iota // 充值
	TxWithdraw                            // 提现
	TxTradeBuy                            // 交易买入
	TxTradeSell                           // 交易卖出
	TxFee                                 // 手续费
	TxTransfer                            // 划转
	TxFunding                             // 资金费用
	TxLiquidation                         // 强平
)

// APIKeyStatus API Key 状态枚举
// [Design: API Key](../DESIGN_ACCOUNT.md#24-api-key)
type APIKeyStatus int

const (
	APIKeyActive   APIKeyStatus = iota // 启用
	APIKeyInactive                     // 禁用
)

// LimitType 风控限额类型枚举
// [Design: 风控限额-risklimit](../DESIGN_ACCOUNT.md#25-风控限额-risklimit)
type LimitType int

const (
	LimitDailyDeposit  LimitType = iota // 每日充值限额
	LimitDailyWithdraw                  // 每日提现限额
	LimitDailyTrade                     // 每日交易限额
	LimitSingleOrder                    // 单笔订单限额
)

// Account 账户结构
// [Design: 账户-account](../DESIGN_ACCOUNT.md#21-账户-account)
type Account struct {
	AccountID       string        `json:"account_id"`       // 账户唯一ID
	UserID          string        `json:"user_id"`          // 用户ID
	Email           string        `json:"email"`            // 邮箱
	Phone           string        `json:"phone"`            // 手机号
	PasswordHash    string        `json:"-"`                // 密码哈希(不序列化)
	Status          AccountStatus `json:"status"`           // 账户状态
	Tier            int           `json:"tier"`             // KYC等级(0-3)
	CreatedAt       time.Time     `json:"created_at"`       // 创建时间
	UpdatedAt       time.Time     `json:"updated_at"`       // 更新时间
	LastLoginAt     time.Time     `json:"last_login_at"`    // 最后登录时间
	ReferralCode    string        `json:"referral_code"`    // 推荐码
	ParentAccountID string        `json:"parent_account_id"`// 父账户ID(子账户)
}

// Balance 余额结构
// [Design: 余额-balance](../DESIGN_ACCOUNT.md#22-余额-balance)
type Balance struct {
	BalanceID string    `json:"balance_id"`   // 余额记录ID
	AccountID string    `json:"account_id"`   // 账户ID
	Currency  string    `json:"currency"`     // 币种(如 BTC, USDT, ETH)
	Available float64   `json:"available"`    // 可用余额
	Frozen    float64   `json:"frozen"`       // 冻结余额
	Total     float64   `json:"total"`        // 总余额(Available + Frozen)
	UpdatedAt time.Time `json:"updated_at"`   // 更新时间
}

// Transaction 资金流水结构
// [Design: 资金流水-transaction](../DESIGN_ACCOUNT.md#23-资金流水-transaction)
type Transaction struct {
	TransactionID   string              `json:"transaction_id"`   // 流水ID
	AccountID       string              `json:"account_id"`       // 账户ID
	Currency        string              `json:"currency"`         // 币种
	Type            TransactionType     `json:"type"`             // 流水类型
	Amount          float64             `json:"amount"`           // 变动金额
	BalanceBefore   float64             `json:"balance_before"`   // 变动前余额
	BalanceAfter    float64             `json:"balance_after"`    // 变动后余额
	RelatedOrderID  string              `json:"related_order_id"` // 关联订单ID
	RelatedTxID     string              `json:"related_tx_id"`    // 关联链上交易ID
	Memo            string              `json:"memo"`             // 备注
	CreatedAt       time.Time           `json:"created_at"`       // 创建时间
}

// APIKey API Key 结构
// [Design: API Key](../DESIGN_ACCOUNT.md#24-api-key)
type APIKey struct {
	APIKeyID    string        `json:"api_key_id"`    // API Key ID
	AccountID   string        `json:"account_id"`    // 账户ID
	Key         string        `json:"key"`           // API Key (公钥)
	Secret      string        `json:"-"`             // Secret Key (私钥，不序列化)
	Permissions []string      `json:"permissions"`   // 权限列表
	Status      APIKeyStatus  `json:"status"`        // 状态
	IPWhitelist []string      `json:"ip_whitelist"`  // IP白名单
	CreatedAt   time.Time     `json:"created_at"`    // 创建时间
	UpdatedAt   time.Time     `json:"updated_at"`    // 更新时间
	LastUsedAt  time.Time     `json:"last_used_at"`  // 最后使用时间
}

// RiskLimit 风控限额结构
// [Design: 风控限额-risklimit](../DESIGN_ACCOUNT.md#25-风控限额-risklimit)
type RiskLimit struct {
	LimitID   string    `json:"limit_id"`     // 限额ID
	AccountID string    `json:"account_id"`   // 账户ID
	Currency  string    `json:"currency"`     // 币种
	LimitType LimitType `json:"limit_type"`   // 限额类型
	Amount    float64   `json:"amount"`       // 限额金额
	UsedToday float64   `json:"used_today"`   // 今日已使用
	ResetTime time.Time `json:"reset_time"`   // 重置时间
}

// CreateAccountRequest 创建账户请求
type CreateAccountRequest struct {
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

// DepositRequest 充值请求
type DepositRequest struct {
	Currency  string  `json:"currency"`
	Amount    float64 `json:"amount"`
	TxID      string  `json:"tx_id"`
	Memo      string  `json:"memo"`
}

// WithdrawRequest 提现请求
type WithdrawRequest struct {
	Currency  string  `json:"currency"`
	Amount    float64 `json:"amount"`
	Address   string  `json:"address"`
	Memo      string  `json:"memo"`
}

// TransferRequest 划转请求
type TransferRequest struct {
	ToAccountID string  `json:"to_account_id"`
	Currency    string  `json:"currency"`
	Amount      float64 `json:"amount"`
	Memo        string  `json:"memo"`
}

// CreateAPIKeyRequest 创建 API Key 请求
type CreateAPIKeyRequest struct {
	Permissions []string `json:"permissions"`
	IPWhitelist []string `json:"ip_whitelist"`
}