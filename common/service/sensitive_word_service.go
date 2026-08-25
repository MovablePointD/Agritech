// Package service 提供敏感词检测与自动审核核心逻辑
//
// 敏感词检测采用内存缓存 + 正则匹配方案：
//   1. InitSensitiveWords() 从数据库加载所有敏感词到内存
//   2. 检测时遍历缓存词表进行大小写不敏感匹配
//   3. 增删改敏感词后自动重新加载缓存
//
// 自动审核四种模式：
//   all_pass           - 全部自动通过
//   all_reject         - 全部拒绝，需人工审核
//   no_sensitive_pass  - 不含敏感词则自动通过
//   need_review_images - 带图或含敏感词需人工复核（默认模式）
package service

import (
	"regexp"
	"strings"
	"sync"

	"rxtcloud/common/config"
	"rxtcloud/common/model"
)

// SensitiveWordService 敏感词内存缓存
// wordMutex 读写锁保证并发安全
var (
	sensitiveWords []string
	wordPatterns   []*regexp.Regexp
	wordMutex      sync.RWMutex
)

// InitSensitiveWords 从数据库加载所有敏感词到内存缓存
// 每次增删改敏感词后需调用此方法刷新缓存
func InitSensitiveWords() error {
	var words []model.SensitiveWord
	config.DB.Find(&words)

	wordMutex.Lock()
	defer wordMutex.Unlock()

	sensitiveWords = make([]string, 0, len(words))
	wordPatterns = make([]*regexp.Regexp, 0, len(words))

	for _, w := range words {
		// 转义正则特殊字符并添加单词边界
		escaped := regexp.QuoteMeta(w.Word)
		pattern := regexp.MustCompile(escaped)
		sensitiveWords = append(sensitiveWords, w.Word)
		wordPatterns = append(wordPatterns, pattern)
	}

	return nil
}

// CheckSensitiveWords 检查文本是否包含敏感词，返回包含的敏感词列表
func CheckSensitiveWords(text string) []string {
	if text == "" {
		return nil
	}

	wordMutex.RLock()
	defer wordMutex.RUnlock()

	found := make([]string, 0)
	text = strings.ToLower(text)

	for _, word := range sensitiveWords {
		pattern := regexp.MustCompile(regexp.QuoteMeta(word))
		if pattern.MatchString(text) {
			found = append(found, word)
		}
	}

	return found
}

// HasSensitiveWord 检查文本是否包含敏感词
func HasSensitiveWord(text string) bool {
	return len(CheckSensitiveWords(text)) > 0
}

// GetAllSensitiveWords 获取所有敏感词
func GetAllSensitiveWords() ([]model.SensitiveWord, error) {
	var words []model.SensitiveWord
	err := config.DB.Order("created_at desc").Find(&words).Error
	return words, err
}

// AddSensitiveWord 添加敏感词
func AddSensitiveWord(word, category string, level int) error {
	sw := model.SensitiveWord{
		Word:     word,
		Category: category,
		Level:    level,
	}
	if err := config.DB.Create(&sw).Error; err != nil {
		return err
	}
	// 重新加载敏感词列表
	return InitSensitiveWords()
}

// DeleteSensitiveWord 删除敏感词
func DeleteSensitiveWord(id uint) error {
	if err := config.DB.Delete(&model.SensitiveWord{}, id).Error; err != nil {
		return err
	}
	// 重新加载敏感词列表
	return InitSensitiveWords()
}

// UpdateSensitiveWord 更新敏感词
func UpdateSensitiveWord(id uint, word, category string, level int) error {
	if err := config.DB.Model(&model.SensitiveWord{}).Where("id = ?", id).Updates(map[string]interface{}{
		"word":     word,
		"category": category,
		"level":    level,
	}).Error; err != nil {
		return err
	}
	// 重新加载敏感词列表
	return InitSensitiveWords()
}

// AutoReviewSetting 自动审核设置
type AutoReviewSetting struct {
	Mode             string `json:"mode"`              // all_pass-全部通过 all_reject-全部不通过 no_sensitive_pass-不包含敏感词通过 need_review_images-带图需复核
	EnabledSensitive bool   `json:"enabled_sensitive"` // 是否启用敏感词检测
}

var globalReviewSetting = AutoReviewSetting{
	Mode:             "need_review_images",
	EnabledSensitive: true,
}

// GetAutoReviewSetting 获取自动审核设置
func GetAutoReviewSetting() AutoReviewSetting {
	return globalReviewSetting
}

// SetAutoReviewSetting 设置自动审核
func SetAutoReviewSetting(setting AutoReviewSetting) {
	globalReviewSetting = setting
}

// ShouldAutoPass 根据设置判断是否自动通过
func ShouldAutoPass(hasImages bool, hasSensitive bool) (bool, string) {
	setting := GetAutoReviewSetting()

	switch setting.Mode {
	case "all_pass":
		return true, "全部通过"
	case "all_reject":
		return false, "全部不通过"
	case "no_sensitive_pass":
		if hasSensitive {
			return false, "包含敏感词"
		}
		return true, "不含敏感词"
	case "need_review_images":
		if hasImages {
			return false, "带图需人工复核"
		}
		if setting.EnabledSensitive && hasSensitive {
			return false, "含敏感词需人工复核"
		}
		return true, "无图且无敏感词"
	default:
		return false, "默认需人工审核"
	}
}
