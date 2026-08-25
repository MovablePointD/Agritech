package controller

import "rxtcloud/post-service/model"

// BuildCommentPostTree 将评论列表转换为树形结构
// level 1: 一级评论，显示在列表中
// level 2: 二级评论，显示在一级评论的 children 中，可回复
// level 3: 三级评论，显示在二级评论的 children 中，不可回复，使用 "回复@用户名" 格式
func BuildCommentPostTree(comments []model.CommentPost) []map[string]interface{} {
	// 按 level 分组
	var level1 []model.CommentPost
	var level2 []model.CommentPost
	var level3 []model.CommentPost

	for _, c := range comments {
		switch c.Level {
		case 1:
			level1 = append(level1, c)
		case 2:
			level2 = append(level2, c)
		case 3:
			level3 = append(level3, c)
		}
	}

	// 构建父评论用户名映射（用于显示回复目标）
	parentUserMap := make(map[uint]string)
	for _, c := range comments {
		if c.User != nil {
			parentUserMap[c.ID] = c.User.Username
		}
	}

	// 构建 level 2 的父级映射（key: 父评论ID, value: 子评论列表）
	level2Map := make(map[uint][]model.CommentPost)
	for _, c := range level2 {
		level2Map[c.ParentID] = append(level2Map[c.ParentID], c)
	}

	// 构建 level 3 的父级映射
	level3Map := make(map[uint][]model.CommentPost)
	for _, c := range level3 {
		level3Map[c.ParentID] = append(level3Map[c.ParentID], c)
	}

	// 构建评论项
	buildCommentItem := func(c model.CommentPost, isChild bool) map[string]interface{} {
		item := map[string]interface{}{
			"id":         c.ID,
			"user_id":    c.UserID,
			"content":    c.Content,
			"level":      c.Level,
			"likes":      c.Likes,
			"created_at": c.CreatedAt,
		}

		// 添加用户信息
		if c.User != nil {
			item["user"] = map[string]interface{}{
				"id":         c.User.ID,
				"username":   c.User.Username,
				"nickname":   c.User.Nickname,
				"avatar_url": c.User.AvatarURL,
			}
		}

		// 添加回复目标用户名（用于 "回复 @xxx" 显示）
		if c.ReplyToUserID > 0 {
			item["reply_to_user"] = parentUserMap[c.ReplyToUserID]
		}

		return item
	}

	// 构建子评论列表
	buildChildren := func(parentID uint, currentLevel int) []map[string]interface{} {
		var result []map[string]interface{}

		if currentLevel == 1 {
			// 构建 level 2
			for _, c := range level2Map[parentID] {
				item := buildCommentItem(c, true)
				// 构建 level 3
				var children []map[string]interface{}
				for _, child := range level3Map[c.ID] {
					children = append(children, buildCommentItem(child, true))
				}
				if len(children) > 0 {
					item["children"] = children
				}
				result = append(result, item)
			}
		}

		return result
	}

	// 构建最终结果
	var result []map[string]interface{}
	for _, c := range level1 {
		item := buildCommentItem(c, false)
		children := buildChildren(c.ID, 1)
		if len(children) > 0 {
			item["children"] = children
		}
		result = append(result, item)
	}

	return result
}
