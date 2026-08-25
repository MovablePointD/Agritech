package main

import (
	"fmt"
	"log"
	"time"

	"rxtcloud/common/model"

	"gorm.io/gorm"
)

func stripDemoPrefix(db *gorm.DB) {
	cols := []struct{ table, col string }{
		{"knowledges", "title"}, {"knowledges", "content"},
		{"products", "title"}, {"products", "content"},
		{"posts", "title"}, {"posts", "content"},
		{"rural_infos", "title"}, {"rural_infos", "content"},
		{"policy_notices", "title"}, {"policy_notices", "content"},
		{"rural_affairs", "title"}, {"rural_affairs", "content"}, {"rural_affairs", "process_content"},
		{"comment_knowledges", "content"},
		{"comment_posts", "content"},
		{"comment_products", "content"},
		{"comment_rural_infos", "content"},
		{"affair_follow_ups", "content"},
	}
	for _, c := range cols {
		sql := fmt.Sprintf("UPDATE %s SET %s = REPLACE(%s, ?, '') WHERE %s LIKE ?", c.table, c.col, c.col, c.col)
		res := db.Exec(sql, demoPrefix, demoPrefix+"%")
		if res.Error != nil {
			log.Printf("清理前缀失败 %s.%s: %v\n", c.table, c.col, res.Error)
		} else if res.RowsAffected > 0 {
			log.Printf("已清理 %s.%s 中 %d 条记录的前缀\n", c.table, c.col, res.RowsAffected)
		}
	}
}

func seedSensitiveWords(db *gorm.DB) {
	// 替换测试用词，补充真实场景敏感词库
	db.Where("word IN ?", []string{"123", "456", "789"}).Delete(&model.SensitiveWord{})

	defs := []model.SensitiveWord{
		{Word: "网络赌博", Level: 3, Category: "gamble"},
		{Word: "六合彩", Level: 3, Category: "gamble"},
		{Word: "刷单返利", Level: 3, Category: "fraud"},
		{Word: "高额回报", Level: 2, Category: "fraud"},
		{Word: "稳赚不赔", Level: 2, Category: "fraud"},
		{Word: "加微信领取", Level: 1, Category: "advertising"},
		{Word: "免费领取", Level: 1, Category: "advertising"},
		{Word: "扫码加群", Level: 1, Category: "advertising"},
		{Word: "假货仿冒", Level: 2, Category: "other"},
		{Word: "违禁农药", Level: 2, Category: "other"},
		{Word: "低俗内容", Level: 2, Category: "porn"},
		{Word: "政治敏感词", Level: 3, Category: "politics"},
	}
	for _, w := range defs {
		var count int64
		db.Model(&model.SensitiveWord{}).Where("word = ?", w.Word).Count(&count)
		if count > 0 {
			continue
		}
		w.CreatedAt = now
		if err := db.Create(&w).Error; err != nil {
			log.Printf("创建敏感词失败 %s: %v\n", w.Word, err)
		} else {
			log.Printf("已创建敏感词: %s (%s)\n", w.Word, w.Category)
		}
	}
}

func seedExtendedPolicies(db *gorm.DB, users map[string]uint, images []string) {
	adminID := users["zhangwei"]
	var admin model.User
	if err := db.Where("username = ?", "admin").First(&admin).Error; err == nil {
		adminID = admin.ID
	}

	defs := []struct {
		title       string
		content     string
		category    string
		address     string
		publishDept string
		publishDate string
		status      int
		isTop       bool
	}{
		{"农机购置与应用补贴实施细则", "对购买列入补贴目录的农机具，按单机补贴上限执行。购机者须在规定时间内完成申请，经审核公示后发放补贴。", "subsidy", "全县", "县农业农村局", "2026-02-20", 1, false},
		{"农村宅基地审批管理补充规定", "严格控制宅基地面积，一户一宅。新建住宅须符合村庄规划，禁止占用永久基本农田。", "land", "全县", "县自然资源局", "2026-01-20", 1, false},
		{"秸秆禁烧与综合利用奖励办法", "全年禁止露天焚烧秸秆。对开展秸秆还田、打捆离田的村集体给予奖励，违者依法处罚。", "environmental", "河东镇", "县生态环境局", "2026-03-10", 1, true},
		{"数字乡村建设试点扶持政策", "支持建设村级电商服务站、智慧农业示范点，对达标项目给予一次性建设补助。", "technology", "河西镇", "县工信局", "2026-04-15", 1, false},
		{"农村养老服务补贴发放公告", "对80岁以上农村老人按月发放高龄津贴，对特困老人提供居家养老服务补贴。", "healthcare", "全县", "县民政局", "2026-05-01", 1, false},
		{"乡村小规模学校优化布局方案", "在保障学生就近入学前提下，统筹配置师资力量，改善农村学校办学条件。", "education", "全县", "县教育局", "2026-03-20", 1, false},
		{"乡村振兴示范村创建工作方案", "围绕产业、人才、文化、生态、组织五个方面，每年遴选10个示范村给予重点扶持。", "comprehensive", "全县", "县乡村振兴局", "2026-06-15", 1, true},
		{"农村饮水安全巩固提升工程通知", "对水质不达标、供水不稳定的村庄纳入改造计划，确保农村居民喝上放心水。", "other", "河东镇", "县水利局", "2026-07-01", 1, false},
		{"耕地地力保护补贴发放公示（草案）", "补贴对象为拥有耕地承包权的种地农民，补贴标准待上级批复后执行。", "subsidy", "全县", "县农业农村局", "2026-08-01", -1, false},
		{"农村公益性岗位管理办法", "开发保洁、护林、河道巡查等公益性岗位，优先安置脱贫户和就业困难人员。", "comprehensive", "河西镇", "县人社局", "2026-02-28", 0, false},
	}

	for _, d := range defs {
		var count int64
		db.Model(&model.PolicyNotice{}).Where("title = ?", d.title).Count(&count)
		if count > 0 {
			continue
		}
		notice := model.PolicyNotice{
			UserID:      adminID,
			Title:       d.title,
			Content:     d.content,
			Category:    d.category,
			Address:     d.address,
			PublishDept: d.publishDept,
			PublishDate: d.publishDate,
			Images:      imgJSON(images[8]),
			Views:       50,
			IsTop:       d.isTop,
			Status:      d.status,
		}
		if err := db.Create(&notice).Error; err != nil {
			log.Printf("创建政策公告失败 %s: %v\n", d.title, err)
		} else {
			log.Printf("已创建政策公告: %s (status=%d)\n", d.title, d.status)
		}
	}
}

func seedExtendedAffairs(db *gorm.DB, users map[string]uint, images []string) {
	processTime := now.Add(-8 * 24 * time.Hour)
	completeTime := now.Add(-1 * 24 * time.Hour)
	firstResponse := now.Add(-6 * 24 * time.Hour)
	followUpTime := now.Add(-3 * 24 * time.Hour)

	type affairDef struct {
		submitter    string
		handler      string
		title        string
		content      string
		typ          string
		address      string
		images       string
		status       int
		rejectReason string
		process      string
		procImages   string
		withFollowUp bool
		withAppeal   bool
	}

	defs := []affairDef{
		{"yanghm", "", "村篮球场地面开裂", "篮球场塑胶地面多处开裂起皮，雨天积水打滑，学生打球存在安全隐患。", "infrastructure", "河东镇永安村小学旁", imgJSON(images[7]), 3, "经现场核查，该项目已纳入镇文体设施年度维修计划，建议下学期统一施工，本次暂不单独受理。", "", "", false, false},
		{"wuxl", "liucl", "养猪场异味扰民", "村东养猪场气味刺鼻，尤其夏季影响周边20余户村民正常生活，希望协调整改。", "env", "河东镇新丰村东侧", imgJSON(images[16]), 6, "", "已约谈养殖场负责人，要求加装除臭设备并控制存栏量，本周内完成整改验收。", imgJSON(images[5]), true, false},
		{"zhaojj", "chenxh", "垃圾分类点选址争议", "新建垃圾分类点距民居过近，村民反映异味和噪音影响休息，要求重新选址。", "other", "河西镇和平村5组", imgJSON(images[5]), 7, "", "经协调，分类点已临时停用，正在重新选址并征求村民意见。", imgJSON(images[5]), false, true},
		{"yanghm", "liucl", "村口水泵设备故障", "灌溉主水泵无法启动，春耕在即，200亩农田灌溉告急。", "hardware", "河东镇永安村泵站", imgJSON(images[18]), 4, "", "已联系维修人员更换电机电容，设备恢复运行，后续将安排季度检修。", imgJSON(images[18]), false, false},
		{"wuxl", "chenxh", "老旧桥梁护栏缺失", "通村小桥一侧护栏断裂缺失，学生上学必经之路，存在跌落风险。", "safety", "河西镇向阳村小桥", imgJSON(images[9]), 5, "", "已安装临时防护栏并完成永久护栏修复，经镇安监办验收合格。", imgJSON(images[9]), false, false},
		{"zhaojj", "", "农田排水沟被建筑垃圾堵塞", "施工队将建筑垃圾倾倒入排水沟，雨季将导致农田内涝。", "env", "河东镇新丰村南侧", imgJSON(images[10]), 1, "", "", "", false, false},
		{"yanghm", "chenxh", "村口监控摄像头失灵", "治安监控3个点位画面丢失，影响夜间治安防控。", "hardware", "河西镇向阳村路口", imgJSON(images[18]), 4, "", "已更换故障摄像头并升级存储设备，监控恢复正常。", imgJSON(images[18]), false, false},
		{"wuxl", "", "废弃农药瓶随意丢弃", "田间地头发现多处废弃农药包装，存在环境污染和误触风险。", "env", "河西镇和平村田间", imgJSON(images[5]), 2, "", "", "", false, false},
		{"zhaojj", "liucl", "村道会车困难需拓宽", "进村道路过窄，农产品运输车会车困难，多次发生剐蹭。", "infrastructure", "河东镇永安村进村村道", imgJSON(images[6]), 5, "", "已完成会车岛增设和警示标线划设，道路通行能力明显改善。", imgJSON(images[6]), false, false},
		{"yanghm", "", "祠堂屋顶瓦片脱落", "百年祠堂屋顶多处瓦片松动脱落，台风季节存在伤人风险。", "safety", "河东镇永安村祠堂", imgJSON(images[4]), 1, "", "", "", false, false},
	}

	for _, d := range defs {
		var count int64
		db.Model(&model.RuralAffair{}).Where("title = ?", d.title).Count(&count)
		if count > 0 {
			continue
		}

		affair := model.RuralAffair{
			Title:        d.title,
			Content:      d.content,
			Type:         d.typ,
			Address:      d.address,
			Images:       d.images,
			UserID:       users[d.submitter],
			Status:       d.status,
			RejectReason: d.rejectReason,
		}

		if d.status >= 2 && d.status != 3 {
			affair.AuditTime = &auditAt
			affair.AuditName = "admin"
		}
		if d.status == 3 {
			affair.AuditTime = &auditAt
			affair.AuditName = "admin"
		}

		if d.handler != "" && d.status >= 4 {
			var handler model.User
			db.Where("username = ?", d.handler).First(&handler)
			affair.HandlerID = handler.ID
			affair.HandlerName = handler.Nickname
			affair.ProcessContent = d.process
			affair.ProcessImages = d.procImages
			affair.ProcessTime = &processTime
			affair.FirstResponseAt = &firstResponse
		}
		if d.status == 5 {
			affair.CompletedTime = &completeTime
		}

		if err := db.Create(&affair).Error; err != nil {
			log.Printf("创建农村事务失败 %s: %v\n", d.title, err)
			continue
		}
		log.Printf("已创建农村事务: %s (type=%s, status=%d)\n", d.title, d.typ, d.status)

		if d.withFollowUp {
			var submitter model.User
			db.Where("username = ?", d.submitter).First(&submitter)
			db.Create(&model.AffairFollowUp{
				AffairID:  affair.ID,
				Type:      "question",
				UserID:    submitter.ID,
				UserName:  submitter.Nickname,
				Content:   "整改后异味明显减轻，但傍晚仍有轻微气味，能否持续跟踪？",
				Images:    imgJSON(images[5]),
				CreatedAt: followUpTime,
			})
			var handler model.User
			db.Where("username = ?", d.handler).First(&handler)
			db.Create(&model.AffairFollowUp{
				AffairID:  affair.ID,
				Type:      "answer",
				UserID:    handler.ID,
				UserName:  handler.Nickname,
				Content:   "已安排专人每日巡查，若再反弹将责令停业整顿，请继续监督。",
				CreatedAt: now.Add(-2 * 24 * time.Hour),
			})
		}

		if d.withAppeal {
			var submitter model.User
			db.Where("username = ?", d.submitter).First(&submitter)
			db.Create(&model.AffairAppeal{
				AffairID:      affair.ID,
				ApplicantType: "user",
				ApplicantID:   submitter.ID,
				ApplicantName: submitter.Nickname,
				Reason:        "对重新选址进度不满，希望明确完成时限并公示方案。",
				Images:        imgJSON(images[5]),
				Status:        "pending",
				CreatedAt:     now.Add(-1 * 24 * time.Hour),
			})
		}
	}
}

func seedModerationSamples(db *gorm.DB, users map[string]uint, images []string) {
	// 含敏感词、待审核的内容样本（模拟审核队列）
	samples := []struct {
		table   string
		userKey string
		title   string
		content string
		status  int
		extra   map[string]interface{}
	}{
		{
			table: "posts", userKey: "wuxl",
			title:   "警惕刷单返利骗局",
			content: "近期有人以刷单返利、高额回报诱导村民转账，请勿轻信，谨防诈骗。",
			status:  -1,
			extra:   map[string]interface{}{"type": "share", "images": imgJSON(images[8])},
		},
		{
			table: "knowledges", userKey: "wangzj",
			title:   "识别假冒农资产品",
			content: "购买种子化肥时注意查验生产日期和防伪标识，切勿购买假货仿冒产品，可向执法部门举报。",
			status:  -1,
			extra:   map[string]interface{}{"image_url": images[12]},
		},
		{
			table: "products", userKey: "zhangwei",
			title:   "特价有机肥促销",
			content: "限时优惠，加微信领取专属折扣，数量有限先到先得。",
			status:  -1,
			extra: map[string]interface{}{
				"price": 49.9, "type": "fertilizer", "image_url": images[19],
				"publisher": users["zhangwei"], "address": "江苏省宿迁市", "stock": 50,
			},
		},
	}

	for _, s := range samples {
		var count int64
		db.Table(s.table).Where("title = ?", s.title).Count(&count)
		if count > 0 {
			continue
		}
		row := map[string]interface{}{
			"title":      s.title,
			"content":    s.content,
			"status":     s.status,
			"created_at": now,
			"updated_at": now,
		}
		if s.table == "posts" {
			row["user_id"] = users[s.userKey]
		} else if s.table == "knowledges" {
			row["user_id"] = users[s.userKey]
		}
		for k, v := range s.extra {
			row[k] = v
		}
		if err := db.Table(s.table).Create(row).Error; err != nil {
			log.Printf("创建待审核样本失败 %s: %v\n", s.title, err)
		} else {
			log.Printf("已创建待审核样本: %s (含敏感词)\n", s.title)
		}
	}
}

// commentThread 三级评论线程定义
type commentThread struct {
	table      string
	targetKey  string
	targetID   uint
	l1User     string
	l1Content  string
	l2User     string
	l2Content  string
	l3User     string
	l3Content  string
	extraCols  map[string]interface{} // e.g. knowledge_id, post_id
}

func seedMultiLevelComments(db *gorm.DB, users map[string]uint) {
	// 选取各模块代表性条目构建三级评论
	var knowledgeID, postID, productID, ruralInfoID, policyID uint
	db.Table("knowledges").Where("title = ?", "水稻育秧关键技术要点").Pluck("id", &knowledgeID)
	db.Table("posts").Where("title = ?", "春耕开始了，田野一片生机").Pluck("id", &postID)
	db.Table("products").Where("title = ?", "东北优质长粒香大米 5kg").Pluck("id", &productID)
	db.Model(&model.RuralInfo{}).Where("title = ?", "永安村概况").Pluck("id", &ruralInfoID)
	db.Model(&model.PolicyNotice{}).Where("title = ?", "2026年粮食种植补贴政策").Pluck("id", &policyID)

	threads := []commentThread{
		{
			table: "comment_knowledges", targetKey: "knowledge",
			l1User: "yanghm", l1Content: "王老师的育秧经验很接地气，正好春耕用得上。",
			l2User: "wangzj", l2Content: "感谢认可，育秧期注意控水控温，有问题随时留言。",
			l3User: "wuxl", l3Content: "我也准备按这个方法试试，后续跟大家分享效果。",
			extraCols: map[string]interface{}{"knowledge_id": knowledgeID},
		},
		{
			table: "comment_posts", targetKey: "post",
			l1User: "zhaojj", l1Content: "春耕照片真好看，我们这边也开始育苗了。",
			l2User: "yanghm", l2Content: "一起加油！今年打算试试有机种植。",
			l3User: "lina", l3Content: "有机种植可以看看我发布的蔬菜配送日记，有土壤改良心得。",
			extraCols: map[string]interface{}{"post_id": postID},
		},
		{
			table: "comment_products", targetKey: "product",
			l1User: "wuxl", l1Content: "大米口感确实好，熬粥特别香。",
			l2User: "zhangwei", l2Content: "谢谢支持，今年新米九月份还会上新。",
			l3User: "zhaojj", l3Content: "能发一下储存方法吗？怕受潮。",
			extraCols: map[string]interface{}{"product_id": productID},
		},
	}

	for _, t := range threads {
		if t.extraCols[t.targetKey+"_id"] == nil || t.extraCols[t.targetKey+"_id"] == uint(0) {
			// fix key naming - knowledge_id not knowledge_id from targetKey
		}
		targetID := uint(0)
		for _, v := range t.extraCols {
			if id, ok := v.(uint); ok && id > 0 {
				targetID = id
			}
		}
		if targetID == 0 {
			continue
		}

		marker := t.l1Content[:12]
		var count int64
		db.Table(t.table).Where("content LIKE ?", marker+"%").Count(&count)
		if count > 0 {
			continue
		}

		createThread(db, t.table, users, t.extraCols, t.l1User, t.l1Content, t.l2User, t.l2Content, t.l3User, t.l3Content)
	}

	// 农村信息与政策三级评论
	if ruralInfoID > 0 {
		var count int64
		db.Model(&model.CommentRuralInfo{}).Where("content = ?", "永安村这几年变化真大，想回去看看。").Count(&count)
		if count == 0 {
			l1 := model.CommentRuralInfo{UserID: users["zhaojj"], TargetType: "rural_info", TargetID: ruralInfoID, Content: "永安村这几年变化真大，想回去看看。", ParentID: 0, Level: 1, Likes: 5}
			db.Create(&l1)
			l2 := model.CommentRuralInfo{UserID: users["zhangwei"], TargetType: "rural_info", TargetID: ruralInfoID, Content: "欢迎回乡！村里民宿和采摘园都办起来了。", ParentID: l1.ID, Level: 2, ReplyToUserID: users["zhaojj"], Likes: 3}
			db.Create(&l2)
			db.Create(&model.CommentRuralInfo{UserID: users["zhaojj"], TargetType: "rural_info", TargetID: ruralInfoID, Content: "太好了，五一假期安排起来！", ParentID: l2.ID, Level: 3, ReplyToUserID: users["zhangwei"], Likes: 1})
		}
	}

	if policyID > 0 {
		var count int64
		db.Model(&model.CommentRuralInfo{}).Where("content = ?", "这个补贴政策什么时候开始申报？").Count(&count)
		if count == 0 {
			l1 := model.CommentRuralInfo{UserID: users["yanghm"], TargetType: "policy_notice", TargetID: policyID, Content: "这个补贴政策什么时候开始申报？", ParentID: 0, Level: 1, Likes: 8}
			db.Create(&l1)
			l2 := model.CommentRuralInfo{UserID: users["wuxl"], TargetType: "policy_notice", TargetID: policyID, Content: "同问，需要准备哪些材料？", ParentID: l1.ID, Level: 2, ReplyToUserID: users["yanghm"], Likes: 4}
			db.Create(&l2)
			db.Create(&model.CommentRuralInfo{UserID: users["zhangwei"], TargetType: "policy_notice", TargetID: policyID, Content: "带上土地承包合同和身份证到镇农业服务中心即可。", ParentID: l2.ID, Level: 3, ReplyToUserID: users["wuxl"], Likes: 6})
		}
	}

	// 为更多条目补充三级评论
	extraKnowledgeTitles := []string{"生猪养殖圈舍消毒规范", "旋耕机日常维护保养"}
	for _, title := range extraKnowledgeTitles {
		var kid uint
		db.Table("knowledges").Where("title = ?", title).Pluck("id", &kid)
		if kid == 0 {
			continue
		}
		marker := "专家解读：" + title[:6]
		var count int64
		db.Table("comment_knowledges").Where("knowledge_id = ? AND content LIKE ?", kid, marker+"%").Count(&count)
		if count > 0 {
			continue
		}
		createThread(db, "comment_knowledges", users,
			map[string]interface{}{"knowledge_id": kid},
			"wuxl", marker, "zhaomy", "说得对，实际操作中还要注意记录消毒日期。", "yanghm", "记下了，回去就落实。")
	}

	log.Println("已创建三级评论线程")
}

func createThread(db *gorm.DB, table string, users map[string]uint, extra map[string]interface{},
	l1User, l1Content, l2User, l2Content, l3User, l3Content string) {

	l1 := map[string]interface{}{
		"user_id":    users[l1User],
		"content":    l1Content,
		"parent_id":  0,
		"level":      1,
		"likes":      3,
		"created_at": now,
	}
	for k, v := range extra {
		l1[k] = v
	}
	db.Table(table).Create(l1)

	var l1ID uint
	db.Table(table).Where("content = ? AND level = 1", l1Content).Pluck("id", &l1ID)

	l2 := map[string]interface{}{
		"user_id":          users[l2User],
		"content":          l2Content,
		"parent_id":        l1ID,
		"level":            2,
		"reply_to_user_id": users[l1User],
		"likes":            2,
		"created_at":       now,
	}
	for k, v := range extra {
		l2[k] = v
	}
	db.Table(table).Create(l2)

	var l2ID uint
	db.Table(table).Where("content = ? AND level = 2", l2Content).Pluck("id", &l2ID)

	l3 := map[string]interface{}{
		"user_id":          users[l3User],
		"content":          l3Content,
		"parent_id":        l2ID,
		"level":            3,
		"reply_to_user_id": users[l2User],
		"likes":            1,
		"created_at":       now,
	}
	for k, v := range extra {
		l3[k] = v
	}
	db.Table(table).Create(l3)
}
