package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/smartx/account/internal/types"
	"github.com/smartx/account/internal/storage"
)

var (
	ErrAPIKeyNotFound = errors.New("api key not found")
	ErrAPIKeyInvalid  = errors.New("api key invalid")
	ErrSignatureInvalid = errors.New("signature invalid")
)

// APIKeyService API Key 服务
// [Design: APIKeyService API Key 服务](../DESIGN_ACCOUNT.md#34-apikeyservice-api-key-服务)
type APIKeyService struct {
	db storage.Database
}

// NewAPIKeyService 创建 API Key 服务
func NewAPIKeyService(db storage.Database) *APIKeyService {
	return &APIKeyService{
		db: db,
	}
}

// generateKeyPair 生成 API Key 密钥对
func generateKeyPair() (string, string) {
	key := uuid.New().String()
	secret := uuid.New().String() + uuid.New().String()
	return key, secret
}

// CreateAPIKey 创建 API Key
// [Design: APIKeyService API Key 服务](../DESIGN_ACCOUNT.md#34-apikeyservice-api-key-服务)
func (s *APIKeyService) CreateAPIKey(accountID string, req *types.CreateAPIKeyRequest) (*types.APIKey, error) {
	key, secret := generateKeyPair()

	apiKey := &types.APIKey{
		APIKeyID:    uuid.New().String(),
		AccountID:   accountID,
		Key:         key,
		Secret:      secret,
		Permissions: req.Permissions,
		Status:      types.APIKeyActive,
		IPWhitelist: req.IPWhitelist,
		CreatedAt:   time.Now(),
	}

	if err := s.db.CreateAPIKey(apiKey); err != nil {
		return nil, err
	}

	return apiKey, nil
}

// GetAPIKey 获取 API Key
// [Design: APIKeyService API Key 服务](../DESIGN_ACCOUNT.md#34-apikeyservice-api-key-服务)
func (s *APIKeyService) GetAPIKey(apiKeyID string) (*types.APIKey, error) {
	return s.db.GetAPIKey(apiKeyID)
}

// GetAPIKeyByKey 通过公钥获取 API Key
func (s *APIKeyService) GetAPIKeyByKey(key string) (*types.APIKey, error) {
	return s.db.GetAPIKeyByKey(key)
}

// ListAPIKeys 列出 API Key
// [Design: APIKeyService API Key 服务](../DESIGN_ACCOUNT.md#34-apikeyservice-api-key-服务)
func (s *APIKeyService) ListAPIKeys(accountID string) ([]*types.APIKey, error) {
	return s.db.ListAPIKeys(accountID)
}

// UpdateAPIKey 更新 API Key
// [Design: APIKeyService API Key 服务](../DESIGN_ACCOUNT.md#34-apikeyservice-api-key-服务)
func (s *APIKeyService) UpdateAPIKey(apiKeyID string, updates map[string]interface{}) (*types.APIKey, error) {
	apiKey, err := s.GetAPIKey(apiKeyID)
	if err != nil {
		return nil, err
	}

	if permissions, ok := updates["permissions"].([]string); ok {
		apiKey.Permissions = permissions
	}

	if ipWhitelist, ok := updates["ip_whitelist"].([]string); ok {
		apiKey.IPWhitelist = ipWhitelist
	}

	if status, ok := updates["status"].(types.APIKeyStatus); ok {
		apiKey.Status = status
	}

	apiKey.UpdatedAt = time.Now()

	if err := s.db.UpdateAPIKey(apiKey); err != nil {
		return nil, err
	}

	return apiKey, nil
}

// DeleteAPIKey 删除 API Key
// [Design: APIKeyService API Key 服务](../DESIGN_ACCOUNT.md#34-apikeyservice-api-key-服务)
func (s *APIKeyService) DeleteAPIKey(apiKeyID string) error {
	return s.db.DeleteAPIKey(apiKeyID)
}

// ValidateAPIKey 验证 API Key
// [Design: APIKeyService API Key 服务](../DESIGN_ACCOUNT.md#34-apikeyservice-api-key-服务)
func (s *APIKeyService) ValidateAPIKey(key, signature, timestamp string) (*types.APIKey, error) {
	apiKey, err := s.GetAPIKeyByKey(key)
	if err != nil {
		return nil, ErrAPIKeyNotFound
	}

	if apiKey.Status != types.APIKeyActive {
		return nil, ErrAPIKeyInvalid
	}

	timestampInt, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return nil, ErrSignatureInvalid
	}

	now := time.Now().Unix()
	if now-timestampInt > 5 {
		return nil, errors.New("signature expired")
	}

	expectedSignature := hmac.New(sha256.New, []byte(apiKey.Secret))
	expectedSignature.Write([]byte(timestamp))
	expectedSignatureStr := hex.EncodeToString(expectedSignature.Sum(nil))

	if !hmac.Equal([]byte(expectedSignatureStr), []byte(signature)) {
		return nil, ErrSignatureInvalid
	}

	apiKey.LastUsedAt = time.Now()
	s.db.UpdateAPIKey(apiKey)

	return apiKey, nil
}