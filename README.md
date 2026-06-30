# Smart Account - 企业级账户管理系统

Smart Account 是一套企业级的账户管理系统，专为交易平台设计，提供多币种余额管理、API Key 认证、资金流追踪和风险限制等核心功能。

## ✨ 核心特性

### 账户管理
- **多币种余额**: 支持多种数字货币和法币的余额管理
- **子账户架构**: 支持主账户和子账户体系
- **账户状态**: 支持正常、冻结、注销等多种账户状态
- **账户操作**: 账户的创建、激活、冻结、解冻、注销

### API Key 认证
- **API Key 生成**: 支持创建多个 API Key
- **权限控制**: 细粒度的权限管理（只读、交易、提现等）
- **IP 白名单**: 支持绑定特定 IP 地址
- **密钥轮换**: 支持 API Key 的定期轮换

### 余额管理
- **实时余额**: 实时追踪账户余额变化
- **可用余额**: 区分可用余额和冻结余额
- **余额对账**: 定期余额对账功能
- **余额变更记录**: 完整的余额变更历史

### 资金流追踪
- **充值记录**: 记录所有充值操作
- **提现记录**: 记录所有提现操作
- **转账记录**: 记录账户间转账
- **交易记录**: 记录交易相关的资金变动

### 风险限制
- **持仓限制**: 单币种最大持仓限制
- **交易限制**: 每日最大交易金额限制
- **提现限制**: 每日最大提现金额限制
- **风险等级**: 支持多等级风险控制

## 🏗️ 架构设计

```
┌─────────────────────────────────────────────────────────────────────┐
│                         API 层                                      │
│                      REST API Endpoints                             │
│  - 账户管理接口 | - API Key 接口 | - 余额接口 | - 交易接口            │
└─────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│                        服务层                                       │
│  ┌───────────────┐ ┌───────────────┐ ┌───────────────┐             │
│  │AccountService │ │ApiKeyService  │ │BalanceService │             │
│  │ (账户服务)     │ │ (API Key服务) │ │  (余额服务)   │             │
│  └───────────────┘ └───────────────┘ └───────────────┘             │
│  ┌───────────────┐                                                  │
│  │TransactionSvc │                                                  │
│  │ (交易服务)     │                                                  │
│  └───────────────┘                                                  │
└─────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│                        存储层                                       │
│                  Memory / Database                                  │
│  - 内存存储 (开发/测试) | - 数据库存储 (生产)                         │
└─────────────────────────────────────────────────────────────────────┘
```

## 📁 项目结构

```
smart_account/
├── cmd/                    # 命令行入口
│   └── server/              # 服务器主程序
│       └── main.go
├── internal/               # 内部模块
│   ├── api/                 # API 处理层
│   │   └── handler.go        # HTTP 处理器
│   ├── service/             # 服务层
│   │   ├── account_service.go    # 账户服务
│   │   ├── api_key_service.go    # API Key 服务
│   │   ├── balance_service.go    # 余额服务
│   │   └── transaction_service.go # 交易服务
│   ├── storage/             # 存储层
│   │   ├── memory.go         # 内存存储实现
│   │   └── storage.go        # 存储接口
│   └── types/               # 类型定义
│       └── types.go          # 通用类型
├── .gitignore              # Git 忽略文件
├── DESIGN_ACCOUNT.md       # 设计文档
├── go.mod                  # Go 依赖管理
└── go.sum                  # Go 依赖校验
```

## 🚀 快速开始

### 环境要求

- **Go**: 1.21+

### 本地开发

```bash
# 进入项目目录
cd smart_account

# 下载依赖
go mod download

# 运行服务器
go run cmd/server/main.go

# 构建可执行文件
go build -o server.exe cmd/server/main.go
```

### 配置说明

系统支持通过环境变量配置：

```bash
# 服务器端口
export PORT=8080

# 存储类型: memory | database
export STORAGE_TYPE=memory

# 数据库配置 (当 STORAGE_TYPE=database 时)
export DB_HOST=localhost
export DB_PORT=3306
export DB_USER=root
export DB_PASSWORD=password
export DB_NAME=smart_account
```

## 🧪 测试

```bash
# 运行所有测试
go test ./...

# 运行特定模块测试
go test ./internal/service/...

# 生成测试覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 🔧 API 接口

### 账户接口

```bash
# 创建账户
POST /api/v1/accounts
Content-Type: application/json

{
  "user_id": "user-001",
  "currency": "USDT",
  "balance": "10000"
}

# 获取账户
GET /api/v1/accounts/{accountId}

# 获取用户账户列表
GET /api/v1/accounts?user_id=user-001

# 更新账户状态
PUT /api/v1/accounts/{accountId}/status
Content-Type: application/json

{
  "status": "FROZEN"
}
```

### API Key 接口

```bash
# 创建 API Key
POST /api/v1/api-keys
Content-Type: application/json

{
  "user_id": "user-001",
  "name": "trading-key",
  "permissions": ["read", "trade"],
  "ip_whitelist": ["192.168.1.1"]
}

# 获取 API Key 列表
GET /api/v1/api-keys?user_id=user-001

# 更新 API Key
PUT /api/v1/api-keys/{keyId}
Content-Type: application/json

{
  "name": "updated-key",
  "status": "DISABLED"
}

# 删除 API Key
DELETE /api/v1/api-keys/{keyId}
```

### 余额接口

```bash
# 获取余额
GET /api/v1/balances/{accountId}

# 充值
POST /api/v1/balances/{accountId}/deposit
Content-Type: application/json

{
  "amount": "1000",
  "tx_hash": "0x1234..."
}

# 提现
POST /api/v1/balances/{accountId}/withdraw
Content-Type: application/json

{
  "amount": "500",
  "address": "0x5678..."
}

# 转账
POST /api/v1/balances/transfer
Content-Type: application/json

{
  "from_account_id": "acc-001",
  "to_account_id": "acc-002",
  "amount": "100"
}
```

### 交易记录接口

```bash
# 获取交易记录
GET /api/v1/transactions?account_id=acc-001&limit=100&offset=0

# 获取交易记录详情
GET /api/v1/transactions/{txId}
```

## 📊 核心模块

### 账户服务

账户服务负责账户的生命周期管理：

- **账户创建**: 创建新账户并初始化余额
- **账户查询**: 根据用户 ID 或账户 ID 查询账户信息
- **账户状态管理**: 冻结、解冻、注销账户
- **子账户管理**: 创建和管理子账户

### API Key 服务

API Key 服务负责 API 认证相关功能：

- **Key 生成**: 生成加密的 API Key 和 Secret
- **权限管理**: 设置和验证 API Key 的操作权限
- **IP 白名单**: 验证请求来源 IP 是否在白名单内
- **Key 轮换**: 支持 API Key 的定期更新

### 余额服务

余额服务负责余额的管理和操作：

- **余额查询**: 查询账户的可用余额和冻结余额
- **余额更新**: 根据交易结果更新账户余额
- **余额冻结**: 冻结部分余额用于订单交易
- **余额解冻**: 解冻已冻结的余额

### 交易服务

交易服务负责记录所有资金变动：

- **交易记录**: 记录充值、提现、转账等操作
- **交易查询**: 根据条件查询交易历史
- **交易统计**: 统计账户的资金流入流出

## 📖 文档

- [设计文档](DESIGN_ACCOUNT.md) - 系统架构和设计细节

## 📝 贡献指南

欢迎提交 Issue 和 Pull Request！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE.txt](LICENSE.txt)

## 📞 联系方式

如有问题或建议，请通过以下方式联系：

- 提交 [Issue](https://github.com/your-username/smart_account/issues)
- 发送邮件至 support@smartaccount.io