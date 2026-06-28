# SmartX 账户系统设计文档

## 1. 系统概述

### 1.1 项目简介

SmartX 账户系统是一个企业级交易所账户管理系统，参考主流交易所（如 Binance、OKX）的账户架构设计，支持多币种余额管理、API Key 认证、资金流水追踪、风控限额等核心功能。

### 1.2 核心特性

| 特性 | 描述 |
|------|------|
| **主账户/子账户体系** | 支持机构用户创建子账户，统一管理资金 |
| **多币种余额** | 支持多种加密货币和法币余额管理 |
| **余额冻结/解冻** | 交易下单时冻结余额，成交后解冻 |
| **资金流水** | 完整记录所有资金变动，支持审计追溯 |
| **API Key 管理** | 支持创建多组 API Key，配置权限范围 |
| **风控限额** | 支持单笔、单日交易限额设置 |
| **账户状态管理** | 支持账户冻结、解冻、风控限制等状态 |

### 1.3 系统架构

```
┌─────────────────────────────────────────────────────────────────────┐
│                        API Gateway                                  │
│  REST API / WebSocket / gRPC                                        │
└──────────────────────┬──────────────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────────────────┐
│                    账户服务层                                        │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                 │
│  │ Account     │  │ Balance     │  │ API Key     │                 │
│  │ Service     │  │ Service     │  │ Service     │                 │
│  └─────────────┘  └─────────────┘  └─────────────┘                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                 │
│  │ Transaction │  │ Risk        │  │ SubAccount  │                 │
│  │ Service     │  │ Service     │  │ Service     │                 │
│  └─────────────┘  └─────────────┘  └─────────────┘                 │
└──────────────────────┬──────────────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────────────────┐
│                    数据存储层                                        │
│  PostgreSQL / Redis / Kafka                                         │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 2. 核心数据结构

### 2.1 账户 (Account)

```go
type Account struct {
    AccountID       string        `json:"account_id"`       // 账户唯一ID
    UserID          string        `json:"user_id"`          // 用户ID
    Email           string        `json:"email"`            // 邮箱
    Phone           string        `json:"phone"`            // 手机号
    Status          AccountStatus `json:"status"`           // 账户状态
    Tier            int           `json:"tier"`             // KYC等级(0-3)
    CreatedAt       time.Time     `json:"created_at"`       // 创建时间
    UpdatedAt       time.Time     `json:"updated_at"`       // 更新时间
    LastLoginAt     time.Time     `json:"last_login_at"`    // 最后登录时间
    ReferralCode    string        `json:"referral_code"`    // 推荐码
    ParentAccountID string        `json:"parent_account_id"`// 父账户ID(子账户)
}
```

**账户状态枚举:**

| 状态 | 值 | 描述 |
|------|-----|------|
| `AccountActive` | 0 | 正常 |
| `AccountFrozen` | 1 | 冻结 |
| `AccountLocked` | 2 | 锁定 |
| `AccountClosed` | 3 | 已关闭 |

**实现代码**: [types.go#L16](internal/types/types.go#L16)

### 2.2 余额 (Balance)

```go
type Balance struct {
    BalanceID   string    `json:"balance_id"`   // 余额记录ID
    AccountID   string    `json:"account_id"`   // 账户ID
    Currency    string    `json:"currency"`     // 币种(如 BTC, USDT, ETH)
    Available   float64   `json:"available"`    // 可用余额
    Frozen      float64   `json:"frozen"`       // 冻结余额
    Total       float64   `json:"total"`        // 总余额(Available + Frozen)
    UpdatedAt   time.Time `json:"updated_at"`   // 更新时间
}
```

**实现代码**: [types.go#L64](internal/types/types.go#L64)

### 2.3 资金流水 (Transaction)

```go
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
```

**流水类型枚举:**

| 类型 | 值 | 描述 |
|------|-----|------|
| `TxDeposit` | 0 | 充值 |
| `TxWithdraw` | 1 | 提现 |
| `TxTradeBuy` | 2 | 交易买入 |
| `TxTradeSell` | 3 | 交易卖出 |
| `TxFee` | 4 | 手续费 |
| `TxTransfer` | 5 | 划转 |
| `TxFunding` | 6 | 资金费用 |
| `TxLiquidation` | 7 | 强平 |

**实现代码**: [types.go#L81](internal/types/types.go#L81)

### 2.4 API Key

```go
type APIKey struct {
    APIKeyID    string        `json:"api_key_id"`    // API Key ID
    AccountID   string        `json:"account_id"`    // 账户ID
    Key         string        `json:"key"`           // API Key (公钥)
    Secret      string        `json:"secret"`        // Secret Key (私钥)
    Permissions []string      `json:"permissions"`   // 权限列表
    Status      APIKeyStatus  `json:"status"`        // 状态
    IPWhitelist []string      `json:"ip_whitelist"`  // IP白名单
    CreatedAt   time.Time     `json:"created_at"`    // 创建时间
    LastUsedAt  time.Time     `json:"last_used_at"`  // 最后使用时间
}
```

**API Key 状态枚举:**

| 状态 | 值 | 描述 |
|------|-----|------|
| `APIKeyActive` | 0 | 启用 |
| `APIKeyInactive` | 1 | 禁用 |

**实现代码**: [types.go#L110](internal/types/types.go#L110)

### 2.5 风控限额 (RiskLimit)

```go
type RiskLimit struct {
    LimitID     string    `json:"limit_id"`     // 限额ID
    AccountID   string    `json:"account_id"`   // 账户ID
    Currency    string    `json:"currency"`     // 币种
    LimitType   LimitType `json:"limit_type"`   // 限额类型
    Amount      float64   `json:"amount"`       // 限额金额
    UsedToday   float64   `json:"used_today"`   // 今日已使用
    ResetTime   time.Time `json:"reset_time"`   // 重置时间
}
```

**限额类型枚举:**

| 类型 | 值 | 描述 |
|------|-----|------|
| `LimitDailyDeposit` | 0 | 每日充值限额 |
| `LimitDailyWithdraw` | 1 | 每日提现限额 |
| `LimitDailyTrade` | 2 | 每日交易限额 |
| `LimitSingleOrder` | 3 | 单笔订单限额 |

**实现代码**: [types.go#L134](internal/types/types.go#L134)

---

## 3. 核心服务设计

### 3.1 AccountService 账户服务

**核心方法:**

| 方法 | 功能 | 参数 | 返回 | 代码位置 |
|------|------|------|------|----------|
| `CreateAccount()` | 创建账户 | email, phone, password | Account, error | [account_service.go#L35](./internal/service/account_service.go#L35) |
| `GetAccount()` | 获取账户 | accountID | Account, error | [account_service.go#L78](./internal/service/account_service.go#L78) |
| `UpdateAccount()` | 更新账户 | accountID, updates | Account, error | [account_service.go#L95](./internal/service/account_service.go#L95) |
| `FrozenAccount()` | 冻结账户 | accountID, reason | error | [account_service.go#L120](./internal/service/account_service.go#L120) |
| `UnlockAccount()` | 解冻账户 | accountID | error | [account_service.go#L145](./internal/service/account_service.go#L145) |
| `CreateSubAccount()` | 创建子账户 | parentID, email | Account, error | [account_service.go#L165](./internal/service/account_service.go#L165) |
| `ListSubAccounts()` | 列出子账户 | parentID | []Account, error | [account_service.go#L195](./internal/service/account_service.go#L195) |

### 3.2 BalanceService 余额服务

**核心方法:**

| 方法 | 功能 | 参数 | 返回 | 代码位置 |
|------|------|------|------|----------|
| `GetBalance()` | 获取余额 | accountID, currency | Balance, error | [balance_service.go#L32](./internal/service/balance_service.go#L32) |
| `GetAllBalances()` | 获取所有余额 | accountID | []Balance, error | [balance_service.go#L55](./internal/service/balance_service.go#L55) |
| `Frozen()` | 冻结余额 | accountID, currency, amount | error | [balance_service.go#L78](./internal/service/balance_service.go#L78) |
| `Unfrozen()` | 解冻余额 | accountID, currency, amount | error | [balance_service.go#L110](./internal/service/balance_service.go#L110) |
| `Deposit()` | 充值 | accountID, currency, amount, txID | Transaction, error | [balance_service.go#L145](./internal/service/balance_service.go#L145) |
| `Withdraw()` | 提现 | accountID, currency, amount, address | Transaction, error | [balance_service.go#L180](./internal/service/balance_service.go#L180) |
| `Transfer()` | 划转 | fromID, toID, currency, amount | Transaction, error | [balance_service.go#L220](./internal/service/balance_service.go#L220) |

### 3.3 TransactionService 流水服务

**核心方法:**

| 方法 | 功能 | 参数 | 返回 | 代码位置 |
|------|------|------|------|----------|
| `GetTransaction()` | 获取流水 | txID | Transaction, error | [transaction_service.go#L28](./internal/service/transaction_service.go#L28) |
| `ListTransactions()` | 列出流水 | accountID, type, limit, offset | []Transaction, error | [transaction_service.go#L45](./internal/service/transaction_service.go#L45) |
| `CreateTransaction()` | 创建流水 | tx | Transaction, error | [transaction_service.go#L75](./internal/service/transaction_service.go#L75) |

### 3.4 APIKeyService API Key 服务

**核心方法:**

| 方法 | 功能 | 参数 | 返回 | 代码位置 |
|------|------|------|------|----------|
| `CreateAPIKey()` | 创建 API Key | accountID, permissions | APIKey, error | [api_key_service.go#L35](./internal/service/api_key_service.go#L35) |
| `GetAPIKey()` | 获取 API Key | apiKeyID | APIKey, error | [api_key_service.go#L85](./internal/service/api_key_service.go#L85) |
| `ListAPIKeys()` | 列出 API Key | accountID | []APIKey, error | [api_key_service.go#L105](./internal/service/api_key_service.go#L105) |
| `UpdateAPIKey()` | 更新 API Key | apiKeyID, updates | APIKey, error | [api_key_service.go#L125](./internal/service/api_key_service.go#L125) |
| `DeleteAPIKey()` | 删除 API Key | apiKeyID | error | [api_key_service.go#L155](./internal/service/api_key_service.go#L155) |
| `ValidateAPIKey()` | 验证 API Key | key, signature, timestamp | bool, error | [api_key_service.go#L175](./internal/service/api_key_service.go#L175) |

---

## 4. API 接口设计

### 4.1 账户管理接口

| 方法 | 路径 | 功能 | 代码位置 |
|------|------|------|----------|
| POST | /api/v1/accounts | 创建账户 | [handler.go#L25](./internal/api/handler.go#L25) |
| GET | /api/v1/accounts/{accountID} | 获取账户 | [handler.go#L68](./internal/api/handler.go#L68) |
| PUT | /api/v1/accounts/{accountID} | 更新账户 | [handler.go#L95](./internal/api/handler.go#L95) |
| POST | /api/v1/accounts/{accountID}/freeze | 冻结账户 | [handler.go#L125](./internal/api/handler.go#L125) |
| POST | /api/v1/accounts/{accountID}/unlock | 解冻账户 | [handler.go#L155](./internal/api/handler.go#L155) |
| POST | /api/v1/accounts/{accountID}/subaccounts | 创建子账户 | [handler.go#L185](./internal/api/handler.go#L185) |
| GET | /api/v1/accounts/{accountID}/subaccounts | 列出子账户 | [handler.go#L215](./internal/api/handler.go#L215) |

### 4.2 余额管理接口

| 方法 | 路径 | 功能 | 代码位置 |
|------|------|------|----------|
| GET | /api/v1/accounts/{accountID}/balances | 获取所有余额 | [handler.go#L245](./internal/api/handler.go#L245) |
| GET | /api/v1/accounts/{accountID}/balances/{currency} | 获取指定余额 | [handler.go#L275](./internal/api/handler.go#L275) |
| POST | /api/v1/accounts/{accountID}/deposit | 充值 | [handler.go#L305](./internal/api/handler.go#L305) |
| POST | /api/v1/accounts/{accountID}/withdraw | 提现 | [handler.go#L335](./internal/api/handler.go#L335) |
| POST | /api/v1/accounts/{accountID}/transfer | 划转 | [handler.go#L375](./internal/api/handler.go#L375) |

### 4.3 流水查询接口

| 方法 | 路径 | 功能 | 代码位置 |
|------|------|----------|
| GET | /api/v1/accounts/{accountID}/transactions | 列出流水 | [handler.go#L415](./internal/api/handler.go#L415) |
| GET | /api/v1/transactions/{txID} | 获取流水详情 | [handler.go#L445](./internal/api/handler.go#L445) |

### 4.4 API Key 管理接口

| 方法 | 路径 | 功能 | 代码位置 |
|------|------|------|----------|
| POST | /api/v1/accounts/{accountID}/api-keys | 创建 API Key | [handler.go#L475](./internal/api/handler.go#L475) |
| GET | /api/v1/accounts/{accountID}/api-keys | 列出 API Key | [handler.go#L505](./internal/api/handler.go#L505) |
| PUT | /api/v1/api-keys/{apiKeyID} | 更新 API Key | [handler.go#L535](./internal/api/handler.go#L535) |
| DELETE | /api/v1/api-keys/{apiKeyID} | 删除 API Key | [handler.go#L565](./internal/api/handler.go#L565) |

---

## 5. 数据库设计

### 5.1 accounts 表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| account_id | VARCHAR(36) | PRIMARY KEY | 账户ID |
| user_id | VARCHAR(36) | UNIQUE | 用户ID |
| email | VARCHAR(255) | UNIQUE | 邮箱 |
| phone | VARCHAR(20) | UNIQUE | 手机号 |
| password_hash | VARCHAR(255) | NOT NULL | 密码哈希 |
| status | INT | DEFAULT 0 | 账户状态 |
| tier | INT | DEFAULT 0 | KYC等级 |
| referral_code | VARCHAR(20) | UNIQUE | 推荐码 |
| parent_account_id | VARCHAR(36) | FOREIGN KEY | 父账户ID |
| created_at | TIMESTAMP | DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMP | DEFAULT NOW() | 更新时间 |
| last_login_at | TIMESTAMP | | 最后登录时间 |

### 5.2 balances 表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| balance_id | VARCHAR(36) | PRIMARY KEY | 余额ID |
| account_id | VARCHAR(36) | FOREIGN KEY | 账户ID |
| currency | VARCHAR(20) | NOT NULL | 币种 |
| available | DECIMAL(32,18) | DEFAULT 0 | 可用余额 |
| frozen | DECIMAL(32,18) | DEFAULT 0 | 冻结余额 |
| updated_at | TIMESTAMP | DEFAULT NOW() | 更新时间 |

**联合索引:** (account_id, currency)

### 5.3 transactions 表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| transaction_id | VARCHAR(36) | PRIMARY KEY | 流水ID |
| account_id | VARCHAR(36) | FOREIGN KEY | 账户ID |
| currency | VARCHAR(20) | NOT NULL | 币种 |
| type | INT | NOT NULL | 流水类型 |
| amount | DECIMAL(32,18) | NOT NULL | 变动金额 |
| balance_before | DECIMAL(32,18) | NOT NULL | 变动前余额 |
| balance_after | DECIMAL(32,18) | NOT NULL | 变动后余额 |
| related_order_id | VARCHAR(36) | | 关联订单ID |
| related_tx_id | VARCHAR(64) | | 关联链上交易ID |
| memo | VARCHAR(500) | | 备注 |
| created_at | TIMESTAMP | DEFAULT NOW() | 创建时间 |

**索引:** (account_id, created_at), (type, created_at)

### 5.4 api_keys 表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| api_key_id | VARCHAR(36) | PRIMARY KEY | API Key ID |
| account_id | VARCHAR(36) | FOREIGN KEY | 账户ID |
| key | VARCHAR(64) | UNIQUE | API Key |
| secret | VARCHAR(128) | NOT NULL | Secret Key |
| permissions | JSON | | 权限列表 |
| status | INT | DEFAULT 0 | 状态 |
| ip_whitelist | JSON | | IP白名单 |
| created_at | TIMESTAMP | DEFAULT NOW() | 创建时间 |
| last_used_at | TIMESTAMP | | 最后使用时间 |

**索引:** (account_id), (key)

### 5.5 risk_limits 表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| limit_id | VARCHAR(36) | PRIMARY KEY | 限额ID |
| account_id | VARCHAR(36) | FOREIGN KEY | 账户ID |
| currency | VARCHAR(20) | NOT NULL | 币种 |
| limit_type | INT | NOT NULL | 限额类型 |
| amount | DECIMAL(32,18) | NOT NULL | 限额金额 |
| used_today | DECIMAL(32,18) | DEFAULT 0 | 今日已使用 |
| reset_time | TIMESTAMP | | 重置时间 |

**联合索引:** (account_id, currency, limit_type)

---

## 6. 安全设计

### 6.1 密码安全

- 使用 bcrypt 算法进行密码哈希
- 密码长度至少 8 位，包含大小写字母和数字
- 登录失败 5 次后锁定账户 30 分钟

### 6.2 API Key 安全

- API Key 使用 HMAC SHA256 签名验证
- 签名有效期 5 秒，防止重放攻击
- 支持 IP 白名单限制
- 权限最小化原则，按需分配权限

### 6.3 数据传输

- 所有 API 接口使用 HTTPS
- 敏感数据传输时进行加密
- WebSocket 使用 WSS 协议

### 6.4 风控安全

- 大额提现需要二次验证（短信/邮箱）
- 账户异常操作触发风控告警
- 支持冷热钱包分离

---

## 7. 项目结构

```
e:\codex\smart_account\
├── cmd/
│   └── server/
│       └── main.go           # 应用入口
├── internal/
│   ├── api/
│   │   └── handler.go        # HTTP API 处理器
│   ├── service/
│   │   ├── account_service.go       # 账户服务
│   │   ├── balance_service.go       # 余额服务
│   │   ├── transaction_service.go   # 流水服务
│   │   ├── api_key_service.go       # API Key 服务
│   │   └── risk_service.go          # 风控服务
│   ├── types/
│   │   └── types.go          # 核心数据结构
│   ├── storage/
│   │   ├── database.go       # 数据库存储
│   │   └── redis.go          # Redis 缓存
│   └── middleware/
│       └── auth.go           # 认证中间件
├── pkg/
│   ├── crypto/               # 加密工具
│   ├── logger/               # 日志
│   └── util/                 # 工具函数
├── DESIGN_ACCOUNT.md         # 设计文档
├── go.mod
└── go.sum
```

---

## 8. 配置说明

```toml
[server]
host = "0.0.0.0"
port = 8081

[database]
dsn = "postgres://user:password@localhost:5432/smart_account"
max_conn = 100

[redis]
addr = "localhost:6379"
db = 0

[jwt]
secret = "your-secret-key"
expire_hours = 24

[rate_limit]
login_max_attempts = 5
login_lock_minutes = 30
```

---

## 9. 部署指南

```bash
# 创建数据库表
go run cmd/migrate/main.go

# 启动服务
go run cmd/server/main.go

# 编译
go build -o account-server cmd/server/main.go
```

---

## 10. 扩展说明

### 10.1 与交易引擎集成

账户系统通过 gRPC 或 HTTP API 与交易引擎交互：

1. **下单时**：交易引擎调用 `BalanceService.Frozen()` 冻结余额
2. **成交时**：交易引擎调用 `BalanceService.Unfrozen()` 解冻并扣减余额
3. **结算时**：交易引擎调用 `BalanceService.Transfer()` 划转资金

### 10.2 与 Smart Scanner 集成

Smart Scanner 监控链上充值交易，回调账户系统完成充值入账：

1. Scanner 扫描到充值交易
2. 发送回调到账户系统
3. 账户系统验证交易并增加余额