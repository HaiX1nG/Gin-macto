package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestVoiceHandler_JoinVoice_BindingError 测试加入语音频道参数校验
func TestVoiceHandler_JoinVoice_BindingError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/v1/channels/:cid/voice/join", (&VoiceHandler{}).JoinVoice)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/channels/abc/voice/join", nil)
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	r.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "参数错误")
}

// TestVoiceHandler_LeaveVoice_BindingError 测试离开语音频道参数校验
func TestVoiceHandler_LeaveVoice_BindingError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/v1/channels/:cid/voice/leave", (&VoiceHandler{}).LeaveVoice)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/channels/abc/voice/leave", nil)
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	r.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "参数错误")
}

// TestVoiceHandler_GetVoiceParticipants_BindingError 测试获取参与者参数校验
func TestVoiceHandler_GetVoiceParticipants_BindingError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/v1/channels/:cid/voice/participants", (&VoiceHandler{}).GetVoiceParticipants)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/channels/abc/voice/participants", nil)
	recorder := httptest.NewRecorder()

	r.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "参数错误")
}

// TestVoiceHandler_SetMute_BindingError 测试静音设置参数校验
func TestVoiceHandler_SetMute_BindingError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/v1/channels/:cid/voice/mute", (&VoiceHandler{}).SetMute)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/channels/abc/voice/mute", strings.NewReader("{"))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	r.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "参数错误")
}

// TestVoiceHandler_StartScreenShare_BindingError 测试开始屏幕共享参数校验
func TestVoiceHandler_StartScreenShare_BindingError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/v1/channels/:cid/screenshare/start", (&VoiceHandler{}).StartScreenShare)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/channels/abc/screenshare/start", nil)
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	r.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "参数错误")
}

// TestVoiceHandler_StopScreenShare_BindingError 测试停止屏幕共享参数校验
func TestVoiceHandler_StopScreenShare_BindingError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/v1/channels/:cid/screenshare/stop", (&VoiceHandler{}).StopScreenShare)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/channels/abc/screenshare/stop", nil)
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	r.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "参数错误")
}

// TestVoiceHandler_GetActiveScreenShare_BindingError 测试获取屏幕共享参数校验
func TestVoiceHandler_GetActiveScreenShare_BindingError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/v1/channels/:cid/screenshare", (&VoiceHandler{}).GetActiveScreenShare)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/channels/abc/screenshare", nil)
	recorder := httptest.NewRecorder()

	r.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "参数错误")
}
