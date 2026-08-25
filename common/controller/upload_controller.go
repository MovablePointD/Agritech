package controller

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"rxtcloud/common/utils"

	"github.com/gin-gonic/gin"
)

// 单文件上传
func UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择文件"})
		return
	}

	// 限制文件大小 10MB
	if file.Size > 10*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件大小不能超过10MB"})
		return
	}

	// 打开文件
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "打开文件失败"})
		return
	}
	defer src.Close()

	// 获取目录类型
	dir := c.DefaultPostForm("dir", "misc")

	// 上传
	url, err := utils.UploadFile(src, file.Filename, dir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "上传失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"url":      url,
		"filename": file.Filename,
		"size":     file.Size,
	})
}

// 批量上传
func UploadFiles(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "获取文件失败"})
		return
	}

	dir := c.DefaultPostForm("dir", "misc")
	files := form.File["files"]

	var results []map[string]string
	for _, f := range files {
		if f.Size > 10*1024*1024 {
			continue
		}

		src, err := f.Open()
		if err != nil {
			continue
		}

		url, err := utils.UploadFile(src, f.Filename, dir)
		src.Close()
		if err != nil {
			continue
		}

		results = append(results, map[string]string{
			"url":      url,
			"filename": f.Filename,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"list": results,
	})
}

// 上传用户头像
func UploadAvatar(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择文件"})
		return
	}

	if file.Size > 5*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "头像大小不能超过5MB"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "打开文件失败"})
		return
	}
	defer src.Close()

	url, err := utils.UploadFile(src, file.Filename, "avatar")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "上传失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": url})
}

// 上传知识图片
func UploadKnowledgeImage(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择文件"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "打开文件失败"})
		return
	}
	defer src.Close()

	url, err := utils.UploadFile(src, file.Filename, "knowledge")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "上传失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": url})
}

// 上传商品图片
func UploadProductImage(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择文件"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "打开文件失败"})
		return
	}
	defer src.Close()

	url, err := utils.UploadFile(src, file.Filename, "product")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "上传失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": url})
}

// 上传动态图片
func UploadPostImage(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择文件"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "打开文件失败"})
		return
	}
	defer src.Close()

	url, err := utils.UploadFile(src, file.Filename, "post")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "上传失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": url})
}

// 上传视频
func UploadVideo(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择视频文件"})
		return
	}

	// 检查文件类型
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := []string{".mp4", ".avi", ".mov", ".wmv", ".flv", ".mkv", ".webm"}
	isAllowed := false
	for _, ext := range allowedExts {
		if ext == strings.ToLower(filepath.Ext(file.Filename)) {
			isAllowed = true
			break
		}
	}
	if !isAllowed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的视频格式，仅支持: mp4, avi, mov, wmv, flv, mkv, webm"})
		return
	}

	// 限制视频大小 100MB
	if file.Size > 100*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "视频大小不能超过100MB"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "打开文件失败"})
		return
	}
	defer src.Close()

	// 使用video目录
	ext = filepath.Ext(file.Filename)
	newFilename := strconv.FormatInt(time.Now().UnixNano(), 10) + ext
	uploadDir := filepath.Join("uploads", "video")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建目录失败"})
		return
	}
	filePath := filepath.Join(uploadDir, newFilename)

	f, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建文件失败"})
		return
	}
	defer f.Close()

	_, err = io.Copy(f, src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败"})
		return
	}

	url := "/uploads/video/" + newFilename
	c.JSON(http.StatusOK, gin.H{
		"url":      url,
		"filename": file.Filename,
		"size":     file.Size,
	})
}

// 删除文件
func DeleteFile(c *gin.Context) {
	url := c.Query("url")
	if url == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请提供文件路径"})
		return
	}

	if err := utils.DeleteFile(url); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "删除成功"})
}

// 获取文件列表（指定目录）
func ListFiles(c *gin.Context) {
	dir := c.DefaultQuery("dir", "")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	basePath := filepath.Join("uploads", dir)
	var files []map[string]interface{}

	if info, err := os.Stat(basePath); err == nil && info.IsDir() {
		filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				relPath, _ := filepath.Rel("uploads", path)
				files = append(files, map[string]interface{}{
					"name": info.Name(),
					"url":  "/uploads/" + strings.Replace(relPath, "\\", "/", -1),
					"size": info.Size(),
				})
			}
			return nil
		})
	}

	total := len(files)
	start := (page - 1) * pageSize
	end := start + pageSize
	if end > total {
		end = total
	}
	if start > total {
		files = []map[string]interface{}{}
	} else {
		files = files[start:end]
	}

	c.JSON(http.StatusOK, gin.H{
		"list":      files,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}
