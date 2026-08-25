// controller/comment_product_tree.go
package controller

import "rxtcloud/product-service/model"

// BuildCommentProductTree 将商品评论列表转换为树形结构
func BuildCommentProductTree(comments []model.CommentProduct) []map[string]interface{} {
	// 按 level 分组
	var level1 []model.CommentProduct
	var level2 []model.CommentProduct
	var level3 []model.CommentProduct

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

	// 构建父评论用户名映射
	parentUserMap := make(map[uint]string)
	for _, c := range comments {
		if c.User != nil {
			parentUserMap[c.ID] = c.User.Username
		}
	}

	// 构建 level 2 的映射
	level2Map := make(map[uint][]model.CommentProduct)
	for _, c := range level2 {
		level2Map[c.ParentID] = append(level2Map[c.ParentID], c)
	}

	// 构建 level 3 的映射
	level3Map := make(map[uint][]model.CommentProduct)
	for _, c := range level3 {
		level3Map[c.ParentID] = append(level3Map[c.ParentID], c)
	}

	// 构建评论项
	buildCommentItem := func(c model.CommentProduct) map[string]interface{} {
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

		// 添加回复目标用户名
		if c.ReplyToUserID > 0 {
			item["reply_to_user"] = parentUserMap[c.ReplyToUserID]
		}

		return item
	}

	// 构建最终结果
	var result []map[string]interface{}
	for _, c := range level1 {
		item := buildCommentItem(c)

		// 构建 level 2 和 level 3
		var children []map[string]interface{}
		for _, c2 := range level2Map[c.ID] {
			childItem := buildCommentItem(c2)
			var grandchildren []map[string]interface{}
			for _, c3 := range level3Map[c2.ID] {
				grandchildren = append(grandchildren, buildCommentItem(c3))
			}
			if len(grandchildren) > 0 {
				childItem["children"] = grandchildren
			}
			children = append(children, childItem)
		}

		if len(children) > 0 {
			item["children"] = children
		}
		result = append(result, item)
	}

	return result
}
