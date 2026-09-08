package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestChannelHandler_CreateChannel_BindingError 测试创建频道参数校验
func TestChannelHandler_CreateChannel_BindingError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/v1/channels", (&ChannelHandler{}).CreateChannel)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/channels", strings.NewReader("{"))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	r.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "参数错误")
}

// TestChannelHandler_GetChannelTree_BindingError 测试获取频道树参数校验
func TestChannelHandler_GetChannelTree_BindingError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/v1/channels/:cid/tree", (&ChannelHandler{}).GetChannelTree)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/channels/abc/tree", nil)
	recorder := httptest.NewRecorder()

	r.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "参数错误")
}

// TestChannelHandler_UpdateChannel_BindingError 测试更新频道参数校验
func TestChannelHandler_UpdateChannel_BindingError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.PUT("/api/v1/channels/:cid", (&ChannelHandler{}).UpdateChannel)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/channels/abc", strings.NewReader("{"))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	r.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "参数错误")
}

// TestChannelHandler_DeleteChannel_BindingError 测试删除频道参数校验
func TestChannelHandler_DeleteChannel_BindingError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.DELETE("/api/v1/channels/:cid", (&ChannelHandler{}).DeleteChannel)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/channels/abc", nil)
	recorder := httptest.NewRecorder()

	r.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "参数错误")
}

// TestChannelHandler_ReorderChannels_BindingError 测试排序频道参数校验
func TestChannelHandler_ReorderChannels_BindingError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/v1/channels/:cid/reorder", (&ChannelHandler{}).ReorderChannels)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/channels/abc/reorder", strings.NewReader("{"))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	r.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "参数错误")
}
