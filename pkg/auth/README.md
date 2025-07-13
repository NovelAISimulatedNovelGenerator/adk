# Auth 认证模块

## 概述

Auth 模块提供了完整的认证和授权功能，支持多种认证方案包括 API Key、HTTP 认证、OAuth2 和 OpenID Connect。该模块基于 OpenAPI 3.0 标准设计，提供了灵活的认证预处理、凭据管理和工具集成能力。

## 核心组件

### 认证方案类型 (AuthSchemeType)
```go
type AuthSchemeType string

const (
    APIKeyScheme        AuthSchemeType = "apiKey"
    HTTPScheme          AuthSchemeType = "http"
    OAuth2Scheme        AuthSchemeType = "oauth2"
    OpenIDConnectScheme AuthSchemeType = "openIdConnect"
)
```

支持四种主要认证方案类型，涵盖现代 API 认证的主要场景。

### OAuth2 授权类型 (OAuthGrantType)
```go
type OAuthGrantType string

const (
    ClientCredentialsGrant OAuthGrantType = "client_credentials"
    AuthorizationCodeGrant OAuthGrantType = "authorization_code"
    ImplicitGrant         OAuthGrantType = "implicit"
    PasswordGrant         OAuthGrantType = "password"
)
```

定义 OAuth2 的四种主要授权流程类型。

## 数据结构

### SecurityScheme (安全方案)
```go
type SecurityScheme struct {
    Type             AuthSchemeType         `json:"type"`
    Description      string                 `json:"description,omitempty"`
    Name             string                 `json:"name,omitempty"`
    In               string                 `json:"in,omitempty"`
    Scheme           string                 `json:"scheme,omitempty"`
    BearerFormat     string                 `json:"bearerFormat,omitempty"`
    Flows            *OAuthFlows            `json:"flows,omitempty"`
    OpenIDConnectURL string                 `json:"openIdConnectUrl,omitempty"`
    ExtraFields      map[string]interface{} `json:"-"`
}
```

**字段说明:**
- **Type**: 认证方案类型
- **Description**: 方案描述
- **Name**: API Key 名称（用于 apiKey 类型）
- **In**: API Key 位置（header/query/cookie）
- **Scheme**: HTTP 认证方案（basic/bearer 等）
- **BearerFormat**: Bearer token 格式说明
- **Flows**: OAuth2 流程配置
- **OpenIDConnectURL**: OpenID Connect 发现端点
- **ExtraFields**: 扩展字段支持

### OAuthFlow (OAuth2 流程)
```go
type OAuthFlow struct {
    AuthorizationURL string            `json:"authorizationUrl,omitempty"`
    TokenURL         string            `json:"tokenUrl,omitempty"`
    RefreshURL       string            `json:"refreshUrl,omitempty"`
    Scopes           map[string]string `json:"scopes,omitempty"`
}
```

### OAuthFlows (OAuth2 流程集合)
```go
type OAuthFlows struct {
    Implicit          *OAuthFlow `json:"implicit,omitempty"`
    Password          *OAuthFlow `json:"password,omitempty"`
    ClientCredentials *OAuthFlow `json:"clientCredentials,omitempty"`
    AuthorizationCode *OAuthFlow `json:"authorizationCode,omitempty"`
}
```

### OpenIDConnectWithConfig (扩展 OpenID Connect 配置)
```go
type OpenIDConnectWithConfig struct {
    Type                              AuthSchemeType         `json:"type"`
    AuthorizationEndpoint             string                 `json:"authorization_endpoint"`
    TokenEndpoint                     string                 `json:"token_endpoint"`
    UserinfoEndpoint                  string                 `json:"userinfo_endpoint,omitempty"`
    RevocationEndpoint                string                 `json:"revocation_endpoint,omitempty"`
    TokenEndpointAuthMethodsSupported []string               `json:"token_endpoint_auth_methods_supported,omitempty"`
    GrantTypesSupported               []string               `json:"grant_types_supported,omitempty"`
    Scopes                            []string               `json:"scopes,omitempty"`
    ExtraFields                       map[string]interface{} `json:"-"`
}
```

支持完整的 OpenID Connect 配置信息。

## 核心功能模块

### 1. 认证预处理器 (AuthPreprocessor)
- 处理认证头的注入和预处理
- 支持多种认证方案的统一处理
- 提供认证上下文管理

### 2. 认证处理器 (AuthHandler)
- 处理认证流程的核心逻辑
- 管理认证状态和会话
- 提供认证结果验证

### 3. 凭据管理 (AuthCredential)
- 安全存储和管理认证凭据
- 支持多种凭据类型
- 提供凭据刷新和轮换机制

### 4. 认证工具 (AuthTool)
- 提供认证相关的工具函数
- 支持认证配置的验证和转换
- 集成工具调用的认证支持

## 使用示例

### API Key 认证
```go
package main

import (
    "github.com/nvcnvn/adk-golang/pkg/auth"
)

func main() {
    // 创建 API Key 认证方案
    apiKeyScheme := &auth.SecurityScheme{
        Type: auth.APIKeyScheme,
        Name: "X-API-Key",
        In:   "header",
        Description: "API Key 认证",
    }
    
    // 使用认证方案
    schemeType := auth.GetAuthSchemeType(apiKeyScheme)
    fmt.Printf("认证方案类型: %s\n", schemeType)
}
```

### OAuth2 客户端凭据流程
```go
// 创建 OAuth2 客户端凭据认证方案
oauth2Scheme := &auth.SecurityScheme{
    Type: auth.OAuth2Scheme,
    Description: "OAuth2 客户端凭据流程",
    Flows: &auth.OAuthFlows{
        ClientCredentials: &auth.OAuthFlow{
            TokenURL: "https://auth.example.com/oauth/token",
            Scopes: map[string]string{
                "read":  "读取权限",
                "write": "写入权限",
            },
        },
    },
}

// 确定授权类型
grantType := auth.FromOAuthFlows(oauth2Scheme.Flows)
fmt.Printf("授权类型: %s\n", grantType) // 输出: client_credentials
```

### HTTP Bearer 认证
```go
// 创建 HTTP Bearer 认证方案
bearerScheme := &auth.SecurityScheme{
    Type:         auth.HTTPScheme,
    Scheme:       "bearer",
    BearerFormat: "JWT",
    Description:  "JWT Bearer Token 认证",
}
```

### OpenID Connect 认证
```go
// 创建 OpenID Connect 认证方案
oidcScheme := &auth.SecurityScheme{
    Type:             auth.OpenIDConnectScheme,
    OpenIDConnectURL: "https://accounts.google.com/.well-known/openid_configuration",
    Description:      "Google OpenID Connect 认证",
}

// 扩展配置
oidcConfig := &auth.OpenIDConnectWithConfig{
    Type:                  auth.OpenIDConnectScheme,
    AuthorizationEndpoint: "https://accounts.google.com/o/oauth2/v2/auth",
    TokenEndpoint:         "https://oauth2.googleapis.com/token",
    UserinfoEndpoint:      "https://openidconnect.googleapis.com/v1/userinfo",
    GrantTypesSupported:   []string{"authorization_code", "refresh_token"},
    Scopes:                []string{"openid", "email", "profile"},
}
```

## 高级功能

### 深拷贝支持
```go
// SecurityScheme 深拷贝
originalScheme := &auth.SecurityScheme{
    Type: auth.APIKeyScheme,
    Name: "Authorization",
    In:   "header",
}

copiedScheme := originalScheme.Copy()
// 修改副本不会影响原始对象
copiedScheme.Name = "X-Auth-Token"
```

### 自定义序列化
```go
// 支持自定义 JSON 序列化/反序列化
schemeJSON := `{
    "type": "apiKey",
    "name": "X-API-Key", 
    "in": "header",
    "customField": "customValue"
}`

var scheme auth.SecurityScheme
err := json.Unmarshal([]byte(schemeJSON), &scheme)
if err == nil {
    // ExtraFields 包含自定义字段
    fmt.Printf("自定义字段: %v\n", scheme.ExtraFields["customField"])
}
```

## 认证流程集成

### 1. 预处理阶段
```go
// 在请求发送前进行认证预处理
func preprocessAuth(req *http.Request, scheme *auth.SecurityScheme) error {
    switch scheme.Type {
    case auth.APIKeyScheme:
        if scheme.In == "header" {
            req.Header.Set(scheme.Name, getAPIKey())
        } else if scheme.In == "query" {
            q := req.URL.Query()
            q.Set(scheme.Name, getAPIKey())
            req.URL.RawQuery = q.Encode()
        }
    case auth.HTTPScheme:
        if scheme.Scheme == "bearer" {
            token := getBearerToken()
            req.Header.Set("Authorization", "Bearer "+token)
        }
    }
    return nil
}
```

### 2. 认证处理
```go
// 处理认证响应和状态管理
func handleAuthResponse(resp *http.Response, scheme *auth.SecurityScheme) error {
    if resp.StatusCode == 401 {
        // 认证失败，可能需要刷新 token
        if scheme.Type == auth.OAuth2Scheme {
            return refreshToken(scheme)
        }
        return fmt.Errorf("认证失败")
    }
    return nil
}
```

## 配置管理

### 认证配置结构
```go
type AuthConfig struct {
    DefaultScheme string                           `yaml:"default_scheme"`
    Schemes       map[string]*auth.SecurityScheme `yaml:"schemes"`
    Credentials   map[string]interface{}           `yaml:"credentials"`
}

// 从配置文件加载认证方案
func LoadAuthConfig(configPath string) (*AuthConfig, error) {
    data, err := os.ReadFile(configPath)
    if err != nil {
        return nil, err
    }
    
    var config AuthConfig
    err = yaml.Unmarshal(data, &config)
    return &config, err
}
```

### 配置文件示例
```yaml
# auth.yaml
default_scheme: "api_key"
schemes:
  api_key:
    type: "apiKey"
    name: "X-API-Key"
    in: "header"
    description: "API Key 认证"
  
  oauth2_client:
    type: "oauth2"
    description: "OAuth2 客户端凭据"
    flows:
      clientCredentials:
        tokenUrl: "https://auth.example.com/oauth/token"
        scopes:
          read: "读取权限"
          write: "写入权限"

credentials:
  api_key: "${API_KEY}"
  oauth2_client_id: "${OAUTH2_CLIENT_ID}"
  oauth2_client_secret: "${OAUTH2_CLIENT_SECRET}"
```

## 安全最佳实践

1. **凭据存储**: 使用环境变量或安全存储服务管理敏感凭据
2. **Token 刷新**: 实现自动 token 刷新机制
3. **请求重试**: 认证失败时的智能重试策略
4. **日志安全**: 避免在日志中记录敏感认证信息
5. **HTTPS**: 所有认证操作必须使用 HTTPS
6. **作用域控制**: 遵循最小权限原则设置 OAuth2 scopes

## 扩展性

模块设计支持灵活扩展：
- 添加新的认证方案类型
- 自定义认证处理逻辑
- 集成第三方认证服务
- 支持多租户认证场景

## 依赖

- Go 标准库: `encoding/json`
- 无外部依赖，保持轻量级设计

## 测试

该模块包含完整的测试覆盖：
- 单元测试覆盖所有认证方案
- 序列化/反序列化测试
- 边界条件和错误处理测试
- 并发安全性测试

Auth 模块为 ADK-Golang 框架提供了企业级的认证和授权能力，是构建安全 API 应用的重要基础设施。
