package controller

import (
	"net/http"

	"rxtcloud/common/service"

	"github.com/gin-gonic/gin"
)

// GetSensitiveWords 获取敏感词列表
func GetSensitiveWords(c *gin.Context) {
	words, err := service.GetAllSensitiveWords()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"list": words})
}

// AddSensitiveWord 添加敏感词
func AddSensitiveWord(c *gin.Context) {
	var input struct {
		Word     string `json:"word" binding:"required"`
		Category string `json:"category"`
		Level    int    `json:"level"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if input.Level == 0 {
		input.Level = 1
	}
	if input.Category == "" {
		input.Category = "other"
	}

	if err := service.AddSensitiveWord(input.Word, input.Category, input.Level); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "添加失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "添加成功"})
}

// DeleteSensitiveWord 删除敏感词
func DeleteSensitiveWord(c *gin.Context) {
	var input struct {
		ID uint `json:"id"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if err := service.DeleteSensitiveWord(input.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "删除成功"})
}

// UpdateSensitiveWord 更新敏感词
func UpdateSensitiveWord(c *gin.Context) {
	var input struct {
		ID       uint   `json:"id" binding:"required"`
		Word     string `json:"word" binding:"required"`
		Category string `json:"category"`
		Level    int    `json:"level"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if input.Level == 0 {
		input.Level = 1
	}
	if input.Category == "" {
		input.Category = "other"
	}

	if err := service.UpdateSensitiveWord(input.ID, input.Word, input.Category, input.Level); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "更新成功"})
}

// GetAutoReviewSetting 获取自动审核设置
func GetAutoReviewSetting(c *gin.Context) {
	setting := service.GetAutoReviewSetting()
	c.JSON(http.StatusOK, setting)
}

// SetAutoReviewSetting 设置自动审核
func SetAutoReviewSetting(c *gin.Context) {
	var input struct {
		Mode             string `json:"mode"`
		EnabledSensitive bool   `json:"enabled_sensitive"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 验证 mode
	validModes := map[string]bool{
		"all_pass":           true,
		"all_reject":         true,
		"no_sensitive_pass":  true,
		"need_review_images": true,
	}
	if !validModes[input.Mode] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的审核模式"})
		return
	}

	service.SetAutoReviewSetting(service.AutoReviewSetting{
		Mode:             input.Mode,
		EnabledSensitive: input.EnabledSensitive,
	})

	c.JSON(http.StatusOK, gin.H{"msg": "设置成功"})
}

// CheckContent 通用内容检测接口
func CheckContent(c *gin.Context) {
	var input struct {
		Text string `json:"text" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	found := service.CheckSensitiveWords(input.Text)
	c.JSON(http.StatusOK, gin.H{
		"has_sensitive": len(found) > 0,
		"words":         found,
	})
}
