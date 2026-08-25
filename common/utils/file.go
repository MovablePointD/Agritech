package utils

import (
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// 上传文件到本地
func UploadFile(file io.Reader, filename string, dir string) (string, error) {
	// 创建上传目录
	uploadDir := filepath.Join("uploads", dir)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", err
	}

	// 生成新文件名，使用纯数字时间戳避免任何格式问题
	ext := filepath.Ext(filename)
	newFilename := strconv.FormatInt(time.Now().UnixNano(), 10) + ext
	filePath := filepath.Join(uploadDir, newFilename)

	// 创建文件
	f, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// 复制内容
	_, err = io.Copy(f, file)
	if err != nil {
		return "", err
	}

	// 返回访问URL
	url := "/uploads/" + dir + "/" + newFilename
	return url, nil
}

// 删除文件
func DeleteFile(url string) error {
	if url == "" {
		return nil
	}
	// 移除开头的 /
	path := strings.TrimPrefix(url, "/")
	return os.Remove(path)
}

// 批量上传文件
func UploadFiles(files []struct {
	File     io.Reader
	Filename string
}, dir string) ([]string, error) {
	var urls []string
	for _, f := range files {
		url, err := UploadFile(f.File, f.Filename, dir)
		if err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}
	return urls, nil
}
