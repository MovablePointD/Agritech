package controller

import (
	"net/http"
	"strconv"
	"time"

	"rxtcloud/common/model"
	"rxtcloud/common/service"
	"rxtcloud/common/utils"

	"github.com/gin-gonic/gin"
)

// 注册
func Register(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Nickname string `json:"nickname"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	if req.Username == "" || req.Password == "" {
		c.JSON(400, gin.H{"error": "用户名和密码不能为空"})
		return
	}

	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(500, gin.H{"error": "密码加密失败"})
		return
	}

	user := model.User{
		Username: req.Username,
		Password: hash,
		Nickname: req.Nickname,
		Role:     "normal",
		Status:   1,
	}
	// 默认昵称等于用户�?
	if user.Nickname == "" {
		user.Nickname = user.Username
	}

	if err := service.CreateUser(user); err != nil {
		c.JSON(500, gin.H{"error": "注册失败"})
		return
	}

	c.JSON(200, gin.H{"msg": "注册成功"})
}

// 登录
func Login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	if req.Username == "" || req.Password == "" {
		c.JSON(400, gin.H{"error": "用户名和密码不能为空"})
		return
	}

	user, err := service.GetUserByUsername(req.Username)
	if err != nil {
		c.JSON(401, gin.H{"error": "用户不存在"})
		return
	}

	if user.Status == 0 {
		// 检查封禁是否已过期
		if user.Banned && user.BannedUntil != nil && time.Now().After(*user.BannedUntil) {
			// 封禁已过期，自动解封
			_ = service.UnbanUser(user.ID)
			user.Status = 1
			user.Banned = false
		} else {
			c.JSON(403, gin.H{"error": "账号已被禁用"})
			return
		}
	}

	if !utils.CheckPassword(user.Password, req.Password) {
		c.JSON(401, gin.H{"error": "密码错误"})
		return
	}

	token, err := utils.GenerateToken(user.ID, user.Role)
	if err != nil {
		c.JSON(500, gin.H{"error": "生成令牌失败"})
		return
	}

	c.JSON(200, gin.H{
		"token": token,
	})
}

// 查询用户
func GetUserList(c *gin.Context) {
	users, _ := service.GetUserList()
	c.JSON(200, users)
}

// 获取当前用户信息
func GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{"error": "未登录"})
		return
	}

	user, err := service.GetUserByID(uint(userID.(uint)))
	if err != nil {
		c.JSON(404, gin.H{"error": "用户不存在"})
		return
	}

	// 不返回密码
	user.Password = ""
	// 如果昵称为空，默认使用用户名
	if user.Nickname == "" {
		user.Nickname = user.Username
	}
	c.JSON(200, user)
}

// 更新用户
func UpdateUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{"error": "未登录"})
		return
	}

	var input struct {
		Nickname  string `json:"nickname"`
		Phone     string `json:"phone"`
		Email     string `json:"email"`
		Signature string `json:"signature"`
		AvatarURL string `json:"avatar_url"`
	}
	c.ShouldBindJSON(&input)

	// 只更新允许的字段，不更新username
	err := service.UpdateUserProfile(uint(userID.(uint)), input.Nickname, input.Phone, input.Email, input.Signature, input.AvatarURL)
	if err != nil {
		c.JSON(500, gin.H{"error": "更新失败"})
		return
	}

	c.JSON(200, gin.H{"msg": "更新成功"})
}

// 搜索用户（用于私信等场景�?
func SearchUsers(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
		return
	}

	keyword := c.DefaultQuery("keyword", "")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	users, total, err := service.SearchUsers(keyword, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "搜索失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"list":      users,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetUserInfoAPI 供其他微服务调用的内部用户信息查询接�?
// GET /api/user/info/:id
// 相比前端接口，此接口不要�?JWT 认证（通过 Nacos 内部网络调用�?
func GetUserInfoAPI(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户ID格式错误"})
		return
	}

	user, err := service.GetUserByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         user.ID,
		"username":   user.Username,
		"nickname":   user.Nickname,
		"avatar_url": user.AvatarURL,
		"role":       user.Role,
		"phone":      user.Phone,
		"email":      user.Email,
	})
}

// 重置密码（忘记密码功能）
func ResetPassword(c *gin.Context) {
	var req struct {
		Username    string `json:"username"`
		NewPassword string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	if req.Username == "" || req.NewPassword == "" {
		c.JSON(400, gin.H{"error": "用户名和新密码不能为空"})
		return
	}

	if len(req.NewPassword) < 6 {
		c.JSON(400, gin.H{"error": "密码长度至少6个字符"})
		return
	}

	// 检查用户是否存�?
	_, err := service.GetUserByUsername(req.Username)
	if err != nil {
		c.JSON(404, gin.H{"error": "用户不存在"})
		return
	}

	// 加密新密�?
	hash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(500, gin.H{"error": "密码加密失败"})
		return
	}

	if err := service.ResetPassword(req.Username, hash); err != nil {
		c.JSON(500, gin.H{"error": "重置密码失败"})
		return
	}

	c.JSON(200, gin.H{"msg": "密码重置成功，请重新登录"})
}

// 删除用户
func DeleteUser(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	service.DeleteUser(uint(id))

	c.JSON(200, gin.H{"msg": "删除成功"})
}

// 封禁用户
func BanUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	var req struct {
		BannedUntil *string `json:"banned_until"` // 封禁截止时间，格�?"2006-01-02 15:04:05"，为空表示永久封�?
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	var until *time.Time
	if req.BannedUntil != nil && *req.BannedUntil != "" {
		t, err := time.Parse("2006-01-02 15:04:05", *req.BannedUntil)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "时间格式错误，请使用 YYYY-MM-DD HH:mm:ss"})
			return
		}
		until = &t
	}

	if err := service.BanUser(uint(id), until); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "封禁失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "封禁成功"})
}

// 解封用户
func UnbanUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	if err := service.UnbanUser(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "解封失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "解封成功"})
}

// 用户自助注销
func SoftDeleteUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	if err := service.SoftDeleteUser(uint(userID.(uint))); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "注销失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "账号已注销"})
}

// 获取已注销用户列表
func GetDeletedUsers(c *gin.Context) {
	keyword := c.DefaultQuery("keyword", "")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	users, total, err := service.GetDeletedUsers(page, pageSize, keyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"list":      users,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// 更新已注销用户信息
func UpdateDeletedUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 允许更新的字段白名单
	allowedFields := map[string]bool{
		"nickname":  true,
		"phone":     true,
		"email":     true,
		"signature": true,
		"banned":    true,
		"status":    true,
		"role":      true,
	}
	filtered := make(map[string]interface{})
	for k, v := range updates {
		if allowedFields[k] {
			filtered[k] = v
		}
	}

	if len(filtered) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "没有可更新的字段"})
		return
	}

	if err := service.UpdateDeletedUser(uint(id), filtered); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "更新成功"})
}

// 管理端获取全部用户列�?
func GetAllUsersForAdmin(c *gin.Context) {
	keyword := c.DefaultQuery("keyword", "")
	role := c.DefaultQuery("role", "")
	bannedStr := c.DefaultQuery("banned", "")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	var banned *bool
	if bannedStr == "true" {
		b := true
		banned = &b
	} else if bannedStr == "false" {
		b := false
		banned = &b
	}

	users, total, err := service.GetAllUsersForAdmin(page, pageSize, keyword, role, banned)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"list":      users,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}
