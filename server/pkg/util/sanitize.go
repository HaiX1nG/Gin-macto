package util

import (
	"html"
	"strings"
)

// EscapeHTML HTML转义，防止XSS攻击
func EscapeHTML(s string) string {
	return html.EscapeString(s)
}

// UnescapeHTML HTML反转义
func UnescapeHTML(s string) string {
	return html.UnescapeString(s)
}

// TrimAndEscape 去除首尾空格并转义HTML
func TrimAndEscape(s string) string {
	return EscapeHTML(strings.TrimSpace(s))
}

// MaskEmail 邮箱脱敏，隐藏部分字符
func MaskEmail(email string) string {
	if email == "" {
		return ""
	}

	atIndex := strings.Index(email, "@")
	if atIndex <= 2 {
		return email
	}

	// 保留前2个字符和@后面的域名
	masked := email[:2] + "***" + email[atIndex:]
	return masked
}

// MaskPhone 手机号脱敏，隐藏中间4位
func MaskPhone(phone string) string {
	if len(phone) < 7 {
		return phone
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}
