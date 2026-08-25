package service

import (
	"strings"
	"sync"

	"rxtcloud/post-service/config"
)

// 敏感词缓存（与 common/service 保持同步，共享同一 MySQL）
var (
	sensitiveWords []string
	wordMutex      sync.RWMutex
)

// InitSensitiveWords 初始化敏感词列表（从共享 DB 加载）
func InitSensitiveWords() error {
	type SensitiveWord struct {
		Word string
	}
	var words []SensitiveWord
	config.DB.Model(&SensitiveWord{}).Find(&words)

	wordMutex.Lock()
	defer wordMutex.Unlock()

	sensitiveWords = make([]string, 0, len(words))
	for _, w := range words {
		sensitiveWords = append(sensitiveWords, w.Word)
	}
	return nil
}

// HasSensitiveWord 检查文本是否包含敏感词
func HasSensitiveWord(text string) bool {
	if text == "" {
		return false
	}
	wordMutex.RLock()
	defer wordMutex.RUnlock()
	text = strings.ToLower(text)
	for _, word := range sensitiveWords {
		if strings.Contains(text, strings.ToLower(word)) {
			return true
		}
	}
	return false
}

// ShouldAutoPass 根据设置判断是否自动通过审核
func ShouldAutoPass(hasImages bool, hasSensitive bool) (bool, string) {
	if hasSensitive {
		return false, "包含敏感词需人工复核"
	}
	if hasImages {
		return false, "带图需人工复核"
	}
	return true, "自动通过"
}
