package middleware

import (
	"net/http"
	"strings"

	"github.com/yourorg/livemix/config"
)

// OriginValidator Origin验证器
// 用于CORS和WebSocket连接的来源验证，防止CSRF攻击
type OriginValidator struct {
	allowOrigins map[string]bool
}

// NewOriginValidator 创建Origin验证器实例
// 参数 allowOrigins: 允许的来源域名列表
func NewOriginValidator(allowOrigins []string) *OriginValidator {
	originMap := make(map[string]bool, len(allowOrigins))
	for _, origin := range allowOrigins {
		// 存储小写版本以支持大小写不敏感匹配
		originMap[strings.ToLower(origin)] = true
	}
	return &OriginValidator{
		allowOrigins: originMap,
	}
}

// IsAllowed 检查指定的Origin是否在白名单中
// 参数 origin: HTTP请求头中的Origin值
// 返回值: true表示允许，false表示拒绝
func (v *OriginValidator) IsAllowed(origin string) bool {
	if origin == "" {
		// 无Origin头的请求（如同源请求、服务器间调用）默认允许
		// 安全说明：浏览器同源请求不会发送Origin头，这类请求应被允许
		return true
	}
	return v.allowOrigins[strings.ToLower(origin)]
}

// CheckOriginFunc 返回用于WebSocket Upgrader的CheckOrigin函数
// 该函数验证WebSocket升级请求的Origin是否在白名单中
func (v *OriginValidator) CheckOriginFunc() func(r *http.Request) bool {
	return func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		return v.IsAllowed(origin)
	}
}

// globalOriginValidator 全局Origin验证器实例
// 由 InitializeOriginValidator 函数初始化
var globalOriginValidator *OriginValidator

// InitializeOriginValidator 初始化全局Origin验证器
// 必须在配置加载完成后调用
// 参数 cfg: 应用配置，包含CORS白名单配置
func InitializeOriginValidator(cfg *config.Config) {
	globalOriginValidator = NewOriginValidator(cfg.CORS.AllowOrigins)
}

// GetOriginValidator 获取全局Origin验证器
// 返回值: 全局Origin验证器实例
func GetOriginValidator() *OriginValidator {
	return globalOriginValidator
}
