package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"rxtcloud/common/model"
	"rxtcloud/common/utils"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	dsn          = "root:root@tcp(127.0.0.1:3306)/gorxt_db?charset=utf8mb4&parseTime=True&loc=Local"
	seedPassword = "123456"
	demoPrefix   = "[演示数据]"
)

var (
	now      = time.Now()
	verifyAt = now.Add(-30 * 24 * time.Hour)
	auditAt  = now.Add(-20 * 24 * time.Hour)
)

func main() {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	stripDemoPrefix(db)

	hash, err := utils.HashPassword(seedPassword)
	if err != nil {
		log.Fatal("密码哈希失败:", err)
	}

	images := downloadSeedImages()
	log.Printf("已准备 %d 张演示图片\n", len(images))

	users := seedUsers(db, hash, images)
	experts := seedExperts(db, users, images)
	processors := seedProcessors(db, users, images)
	seedKnowledges(db, experts, images)
	seedProducts(db, users, images)
	seedPosts(db, users, images)
	ruralInfoIDs := seedRuralInfos(db, users, images)
	policyIDs := seedPolicyNotices(db, users, images)
	seedRuralAffairs(db, users, processors, images)
	seedComments(db, users, ruralInfoIDs, policyIDs)

	seedSensitiveWords(db)
	seedExtendedPolicies(db, users, images)
	seedExtendedAffairs(db, users, images)
	seedModerationSamples(db, users, images)
	seedMultiLevelComments(db, users)

	log.Println("========================================")
	log.Println("种子数据导入完成！")
	log.Println("演示账号密码均为: 123456")
	log.Println("专家账号: wangzj / zhaomy / sunjy")
	log.Println("事务处理: liucl / chenxh")
	log.Println("普通用户: yanghm / wuxl / zhaojj")
	log.Println("农户账号: zhangwei / lina")
	log.Println("========================================")
}

func downloadSeedImages() []string {
	dir := filepath.Join("uploads", "misc")
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatal("创建上传目录失败:", err)
	}

	seeds := []string{
		"farm-rice", "farm-corn", "farm-vegetable", "farm-fruit",
		"rural-village", "rural-road", "rural-market", "rural-culture",
		"policy-doc", "affair-bridge", "affair-water", "product-seed",
		"knowledge-tech", "post-harvest", "avatar-man", "avatar-woman",
		"livestock", "machinery", "greenhouse", "fertilizer",
	}

	var urls []string
	client := &http.Client{Timeout: 30 * time.Second}
	for _, seed := range seeds {
		filename := "seed_" + seed + ".jpg"
		filePath := filepath.Join(dir, filename)
		url := "/uploads/misc/" + filename

		if _, err := os.Stat(filePath); err == nil {
			urls = append(urls, url)
			continue
		}

		imgURL := fmt.Sprintf("https://picsum.photos/seed/%s/800/600.jpg", seed)
		resp, err := client.Get(imgURL)
		if err != nil {
			log.Printf("下载图片失败 %s: %v，使用占位路径\n", seed, err)
			urls = append(urls, url)
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			log.Printf("下载图片失败 %s: HTTP %d\n", seed, resp.StatusCode)
			urls = append(urls, url)
			continue
		}
		if err := os.WriteFile(filePath, body, 0644); err != nil {
			log.Printf("保存图片失败 %s: %v\n", seed, err)
		} else {
			log.Printf("已下载图片: %s\n", url)
		}
		urls = append(urls, url)
	}
	return urls
}

func imgJSON(urls ...string) string {
	if len(urls) == 0 {
		return "[]"
	}
	b, _ := json.Marshal(urls)
	return string(b)
}

type seedUser struct {
	Username  string
	Nickname  string
	Role      string
	Phone     string
	Email     string
	Signature string
	Avatar    string
}

func seedUsers(db *gorm.DB, hash string, images []string) map[string]uint {
	defs := []seedUser{
		{"zhangwei", "张伟", "farmer", "13800001001", "zhangwei@demo.com", "扎根田野，用心耕种", images[14]},
		{"lina", "李娜", "farmer", "13800001002", "lina@demo.com", "绿色种植，健康生活", images[15]},
		{"wangzj", "王建国", "expert", "13800002001", "wangzj@demo.com", "水稻种植技术推广", images[14]},
		{"zhaomy", "赵明远", "expert", "13800002002", "zhaomy@demo.com", "畜禽养殖与疫病防控", images[15]},
		{"sunjy", "孙建业", "expert", "13800002003", "sunjy@demo.com", "农机装备与田间管理", images[14]},
		{"liucl", "刘春兰", "processor", "13800003001", "liucl@demo.com", "负责河东片区农村事务", images[15]},
		{"chenxh", "陈晓红", "processor", "13800003002", "chenxh@demo.com", "负责河西片区民生事务", images[14]},
		{"yanghm", "杨红梅", "normal", "13800004001", "yanghm@demo.com", "关注乡村发展", images[15]},
		{"wuxl", "吴小龙", "normal", "13800004002", "wuxl@demo.com", "热爱分享田园生活", images[14]},
		{"zhaojj", "赵俊杰", "normal", "13800004003", "zhaojj@demo.com", "返乡创业青年", images[15]},
	}

	result := make(map[string]uint)
	for _, d := range defs {
		var existing model.User
		if err := db.Where("username = ?", d.Username).First(&existing).Error; err == nil {
			result[d.Username] = existing.ID
			log.Printf("用户已存在，跳过: %s (id=%d)\n", d.Username, existing.ID)
			continue
		}
		u := model.User{
			Username:  d.Username,
			Password:  hash,
			Nickname:  d.Nickname,
			Phone:     d.Phone,
			Email:     d.Email,
			AvatarURL: d.Avatar,
			Signature: d.Signature,
			Role:      d.Role,
			Status:    1,
		}
		if err := db.Create(&u).Error; err != nil {
			log.Fatalf("创建用户 %s 失败: %v", d.Username, err)
		}
		result[d.Username] = u.ID
		log.Printf("已创建用户: %s (id=%d)\n", d.Username, u.ID)
	}
	return result
}

func seedExperts(db *gorm.DB, users map[string]uint, images []string) map[string]uint {
	defs := []struct {
		username   string
		realName   string
		profession string
		title      string
		company    string
		intro      string
		certNo     string
	}{
		{"wangzj", "王建国", "种植", "高级农艺师", "县农业技术推广中心", "从事水稻、小麦栽培技术研究推广20年，擅长病虫害绿色防控。", "AGRI-2021-00128"},
		{"zhaomy", "赵明远", "养殖", "畜牧兽医师", "市畜牧兽医站", "专注生猪、家禽规模化养殖与疫病防控，服务农户超500户。", "VET-2020-00856"},
		{"sunjy", "孙建业", "农机", "农机推广工程师", "省农机推广总站", "熟悉各类农机具选型、维护与田间配套，推动智慧农业落地。", "MECH-2022-00341"},
	}

	result := make(map[string]uint)
	for i, d := range defs {
		uid := users[d.username]
		var existing model.Expert
		if err := db.Where("user_id = ?", uid).First(&existing).Error; err == nil {
			if existing.Status != 1 {
				db.Model(&existing).Updates(map[string]interface{}{"status": 1, "verify_at": verifyAt})
			}
			result[d.username] = existing.ID
			log.Printf("专家已存在，跳过: %s\n", d.username)
			continue
		}
		e := model.Expert{
			UserID:     uid,
			RealName:   d.realName,
			Phone:      "13800002" + fmt.Sprintf("%03d", i+1),
			Profession: d.profession,
			Title:      d.title,
			Company:    d.company,
			Intro:      d.intro,
			CertNo:     d.certNo,
			CertImage:  images[8],
			Status:     1,
			VerifyAt:   &verifyAt,
		}
		if err := db.Create(&e).Error; err != nil {
			log.Fatalf("创建专家 %s 失败: %v", d.username, err)
		}
		db.Model(&model.User{}).Where("id = ?", uid).Update("role", "expert")
		result[d.username] = e.ID
		log.Printf("已创建专家: %s (id=%d)\n", d.realName, e.ID)
	}
	return result
}

func seedProcessors(db *gorm.DB, users map[string]uint, images []string) map[string]uint {
	defs := []struct {
		username string
		realName string
		phone    string
		address  string
		intro    string
	}{
		{"liucl", "刘春兰", "13800003001", "河东镇、永安村、新丰村", "负责道路养护、环境整治、公共设施报修等事务协调处理。"},
		{"chenxh", "陈晓红", "13800003002", "河西镇、和平村、向阳村", "负责水利设施、农田灌溉、安全隐患排查等农村事务。"},
	}

	result := make(map[string]uint)
	for _, d := range defs {
		uid := users[d.username]
		var existing model.AffairProcessor
		if err := db.Where("user_id = ?", uid).First(&existing).Error; err == nil {
			if existing.Status != 1 {
				db.Model(&existing).Updates(map[string]interface{}{"status": 1, "audit_time": auditAt, "audit_name": "admin"})
			}
			result[d.username] = existing.ID
			log.Printf("事务处理人员已存在，跳过: %s\n", d.username)
			continue
		}
		p := model.AffairProcessor{
			UserID:     uid,
			RealName:   d.realName,
			Phone:      d.phone,
			Address:    d.address,
			Intro:      d.intro,
			CertImages: imgJSON(images[8], images[9]),
			Status:     1,
			AuditTime:  &auditAt,
			AuditName:  "admin",
		}
		if err := db.Create(&p).Error; err != nil {
			log.Fatalf("创建事务处理人员 %s 失败: %v", d.username, err)
		}
		db.Model(&model.User{}).Where("id = ?", uid).Update("role", "processor")
		result[d.username] = p.ID
		log.Printf("已创建事务处理人员: %s (id=%d)\n", d.realName, p.ID)
	}
	return result
}

func seedKnowledges(db *gorm.DB, experts map[string]uint, images []string) {
	type kDef struct {
		expertUser string
		title      string
		content    string
		image      string
	}
	defs := []kDef{
		{"wangzj",  "水稻育秧关键技术要点", "水稻育秧应控制秧龄在25-30天，保持秧田水层3-5厘米。移栽前3天排水炼苗，提高抗逆性。注意防治稻瘟病、恶苗病，可选用咪鲜胺浸种。", images[0]},
		{"wangzj",  "小麦赤霉病综合防治方案", "小麦赤霉病防治以预防为主，在抽穗扬花期使用戊唑醇、氰烯菌酯等药剂。收获后及时翻耕，减少病源。避免在阴雨天气收获湿麦。", images[1]},
		{"zhaomy",  "生猪养殖圈舍消毒规范", "进猪前用氢氧化钠溶液对圈舍全面消毒，空栏期不少于7天。日常消毒每周2次，重点关注食槽、饮水器和粪沟。做好人员进出消毒管理。", images[16]},
		{"zhaomy",  "家禽免疫程序参考指南", "肉鸡7日龄新城疫首免，14日龄二免；蛋鸡开产前加强禽流感免疫。免疫前后3天避免应激，确保疫苗冷链运输。建立免疫档案备查。", images[17]},
		{"sunjy",  "旋耕机日常维护保养", "每季作业后清除刀具泥土，检查轴承润滑。刀片磨损超过原厚度1/3应更换。存放时抬起支撑，避免轮胎长期承压变形。", images[18]},
		{"sunjy",  "无人机植保作业注意事项", "作业前检查电池电量和喷头雾化效果，风速超过3级暂停作业。航线规划避开电线杆和树木，药液添加应使用专用助剂。", images[18]},
		{"wangzj",  "有机蔬菜大棚温湿度管理", "白天棚温控制在25-28℃，夜间15-18℃。湿度超过85%时及时通风，配合滴灌控制土壤含水量。采用黄板诱杀蚜虫，减少化学农药使用。", images[2]},
		{"zhaomy",  "生态鱼塘水质调控方法", "保持溶氧量5mg/L以上，定期补新水。投喂量以30分钟内吃完为宜，避免残饵污染。夏季增开增氧机，防止泛塘。", images[3]},
	}

	expertUsers := map[string]uint{}
	for uname := range experts {
		var u model.User
		db.Where("username = ?", uname).First(&u)
		expertUsers[uname] = u.ID
	}

	for _, d := range defs {
		var count int64
		db.Table("knowledges").Where("title = ?", d.title).Count(&count)
		if count > 0 {
			continue
		}
		k := map[string]interface{}{
			"title":      d.title,
			"content":    d.content,
			"user_id":    expertUsers[d.expertUser],
			"image_url":  d.image,
			"status":     1,
			"likes":      0,
			"created_at": now,
			"updated_at": now,
		}
		if err := db.Table("knowledges").Create(k).Error; err != nil {
			log.Printf("创建知识失败 %s: %v\n", d.title, err)
		} else {
			log.Printf("已创建知识: %s\n", d.title)
		}
	}
}

func seedProducts(db *gorm.DB, users map[string]uint, images []string) {
	defs := []struct {
		publisher string
		title     string
		content   string
		price     float64
		typ       string
		address   string
		stock     int
		image     string
	}{
		{"zhangwei",  "东北优质长粒香大米 5kg", "产自黑土地核心产区，米粒晶莹、口感软糯，当季新米现磨现发。", 39.9, "grain", "黑龙江省五常市", 200, images[0]},
		{"zhangwei",  "有机黄小米 2.5kg", "无农药残留，熬粥绵软香甜，适合老人儿童食用。", 28.5, "grain", "山西省忻州市", 150, images[1]},
		{"lina",  "新鲜红颜草莓 礼盒装", "大棚直采，果形饱满、甜度高，顺丰冷链配送。", 68.0, "fruit", "辽宁省丹东市", 80, images[3]},
		{"lina",  "农家土鸡蛋 30枚装", "散养土鸡产蛋，蛋黄饱满，营养丰富。", 45.0, "livestock", "河北省保定市", 120, images[16]},
		{"zhangwei",  "非转基因压榨花生油 5L", "物理压榨，香味浓郁，适合煎炸炒菜。", 89.0, "oil", "山东省临沂市", 60, images[12]},
		{"lina",  "高山绿茶 250g", "明前采摘，汤色清亮，回甘持久。", 128.0, "tea", "浙江省杭州市", 90, images[2]},
		{"zhangwei",  "复合有机肥 40kg", "改良土壤、促根壮苗，适用于蔬菜果树。", 55.0, "fertilizer", "江苏省宿迁市", 300, images[19]},
		{"lina",  "手工红薯粉条 3斤装", "传统工艺制作，久煮不烂，口感筋道。", 32.0, "processed", "河南省商丘市", 180, images[1]},
	}

	for _, d := range defs {
		var count int64
		db.Table("products").Where("title = ?", d.title).Count(&count)
		if count > 0 {
			continue
		}
		p := map[string]interface{}{
			"title":      d.title,
			"content":    d.content,
			"price":      d.price,
			"type":       d.typ,
			"image_url":  d.image,
			"publisher":  users[d.publisher],
			"partner":    0,
			"address":    d.address,
			"stock":      d.stock,
			"status":     1,
			"created_at": now,
			"updated_at": now,
		}
		if err := db.Table("products").Create(p).Error; err != nil {
			log.Printf("创建商品失败 %s: %v\n", d.title, err)
		} else {
			log.Printf("已创建商品: %s\n", d.title)
		}
	}
}

func seedPosts(db *gorm.DB, users map[string]uint, images []string) {
	defs := []struct {
		user    string
		title   string
		content string
		typ     string
		images  string
		likes   int
		views   int
	}{
		{"yanghm",  "春耕开始了，田野一片生机", "今天和邻居们一起下地育苗，虽然辛苦但充满希望。今年打算多种两亩蔬菜，欢迎大家来交流种植经验！", "share", imgJSON(images[0], images[2]), 12, 86},
		{"wuxl",  "乡村集市见闻", "周末去了镇上的集市，土特产琳琅满目，红薯、腊肉、手工豆腐都很受欢迎。乡村振兴真的在改变我们的生活。", "normal", imgJSON(images[5], images[6]), 8, 54},
		{"zhaojj",  "请教：大棚番茄叶子发黄怎么办？", "最近大棚番茄下部叶片发黄，不知道是缺肥还是病害，有经验的老师傅帮忙看看照片，谢谢！", "question", imgJSON(images[2]), 3, 42},
		{"yanghm",  "晒晒我家果园的收成", "今年苹果产量不错，个头均匀、色泽红润。准备一部分做电商，一部分送亲戚朋友。", "share", imgJSON(images[3]), 15, 120},
		{"wuxl",  "村里新修了产业路", "以前下雨天泥泞难行，现在水泥路通到地头，农机进出方便多了，为村委会点赞！", "normal", imgJSON(images[6]), 20, 98},
		{"zhaojj",  "返乡创业第一步：注册家庭农场", "手续比想象中简单，政务服务中心一站式服务。准备先从特色种植做起，慢慢摸索市场。", "share", imgJSON(images[4]), 6, 35},
		{"lina",  "有机蔬菜配送日记", "今天给城区客户配送了50份蔬菜包，全部当天采摘。绿色健康，从田间到餐桌不超过12小时。", "normal", imgJSON(images[2], images[3]), 18, 76},
		{"zhangwei",  "秋收纪实：颗粒归仓", "联合收割机作业效率真高，半天就收完十亩地。粮食安全，人人有责。", "share", imgJSON(images[0], images[13]), 25, 156},
	}

	for _, d := range defs {
		var count int64
		db.Table("posts").Where("title = ?", d.title).Count(&count)
		if count > 0 {
			continue
		}
		p := map[string]interface{}{
			"user_id":    users[d.user],
			"title":      d.title,
			"content":    d.content,
			"images":     d.images,
			"type":       d.typ,
			"likes":      d.likes,
			"views":      d.views,
			"status":     1,
			"created_at": now,
			"updated_at": now,
		}
		if err := db.Table("posts").Create(p).Error; err != nil {
			log.Printf("创建动态失败 %s: %v\n", d.title, err)
		} else {
			log.Printf("已创建动态: %s\n", d.title)
		}
	}
}

func seedRuralInfos(db *gorm.DB, users map[string]uint, images []string) []uint {
	defs := []struct {
		user    string
		title   string
		content string
		typ     string
		address string
		images  string
		views   int
	}{
		{"zhangwei",  "永安村概况", "永安村位于河东镇中心，全村328户、1260人，耕地面积2100亩。以水稻、蔬菜种植为主，近年来发展乡村旅游和电商产业。", "overview", "河东镇永安村", imgJSON(images[4]), 230},
		{"lina",  "新丰村农业资源介绍", "新丰村拥有优质水田800亩、果园200亩、鱼塘50亩。盛产大米、柑橘和生态鱼，已形成产供销一体化链条。", "resource", "河东镇新丰村", imgJSON(images[0], images[3]), 185},
		{"zhangwei",  "和平村传统文化", "和平村保留了完整的客家围屋建筑群，每年举办丰收节、舞龙舞狮等传统活动，是县级非物质文化遗产传承地。", "culture", "河西镇和平村", imgJSON(images[7]), 142},
		{"lina",  "向阳村交通设施", "村内实现组组通水泥路，距高速入口8公里，县道X102穿村而过。建有农村公交停靠点，日发车6班次。", "transportation", "河西镇向阳村", imgJSON(images[6]), 98},
		{"zhangwei",  "永安村教育医疗资源", "村内有完全小学1所、村级卫生室1个，距镇卫生院3公里。老年人日间照料中心覆盖全村60岁以上老人。", "education", "河东镇永安村", imgJSON(images[4]), 76},
		{"lina",  "河西镇综合介绍", "河西镇辖12个行政村，总人口1.8万人。镇域以特色果蔬和生态养殖为支柱产业，获评省级美丽乡村示范镇。", "village_intro", "河西镇", imgJSON(images[4], images[5]), 310},
		{"zhangwei",  "河东镇水利概况", "全镇建有小型水库2座、灌溉渠系126公里，有效灌溉面积1.2万亩，保障粮食稳产高产。", "resource", "河东镇", imgJSON(images[10]), 67},
		{"lina",  "乡村民宿发展现状", "全镇登记民宿15家，床位380个，年均接待游客2万人次。主打农耕体验和果蔬采摘。", "other", "河东镇", imgJSON(images[5]), 89},
	}

	var ids []uint
	for _, d := range defs {
		var count int64
		db.Model(&model.RuralInfo{}).Where("title = ?", d.title).Count(&count)
		if count > 0 {
			var existing model.RuralInfo
			db.Where("title = ?", d.title).First(&existing)
			ids = append(ids, existing.ID)
			continue
		}
		info := model.RuralInfo{
			UserID:  users[d.user],
			Title:   d.title,
			Content: d.content,
			Type:    d.typ,
			Address: d.address,
			Images:  d.images,
			Views:   d.views,
			Status:  1,
		}
		if err := db.Create(&info).Error; err != nil {
			log.Printf("创建农村信息失败 %s: %v\n", d.title, err)
		} else {
			ids = append(ids, info.ID)
			log.Printf("已创建农村信息: %s\n", d.title)
		}
	}
	return ids
}

func seedPolicyNotices(db *gorm.DB, users map[string]uint, images []string) []uint {
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
		images      string
		isTop       bool
		views       int
	}{
		{ "2026年粮食种植补贴政策", "对种植水稻、小麦的农户按种植面积给予每亩120元补贴。申报时间为3月1日至4月30日，请携带土地承包合同到镇农业服务中心办理。", "subsidy", "河东镇", "县农业农村局", "2026-03-01", imgJSON(images[8]), true, 520},
		{ "农村土地流转管理办法", "规范农村土地流转行为，保障农民合法权益。流转期限不得超过承包期的剩余期限，流转收益归承包方所有。", "land", "全县", "县人民政府", "2026-01-15", imgJSON(images[8]), false, 380},
		{ "畜禽养殖污染防治条例", "规模养殖场须建设粪污处理设施，严禁直排河道。散养户应做好粪污资源化利用，违者将依法处罚。", "environmental", "河西镇", "县生态环境局", "2026-02-10", imgJSON(images[16]), false, 210},
		{ "智慧农业设备购置补贴方案", "对购置无人机植保、智能灌溉等设备的农业经营主体，按购置金额30%给予补贴，最高不超过5万元。", "technology", "全县", "县农业农村局", "2026-04-01", imgJSON(images[18]), true, 445},
		{ "农村医疗保险缴费通知", "2026年度农村居民医保个人缴费标准为每人380元，集中缴费期为9月1日至12月31日。", "healthcare", "全县", "县医疗保障局", "2026-09-01", imgJSON(images[8]), false, 890},
		{ "乡村教师支持计划实施方案", "对乡村学校教师给予生活补助和职称倾斜，鼓励优秀人才到农村任教。", "education", "全县", "县教育局", "2026-03-15", imgJSON(images[8]), false, 156},
		{ "高标准农田建设政策解读", "2026年计划新建高标准农田5000亩，完善灌排设施和田间道路，项目区优先纳入补贴范围。", "comprehensive", "河东镇", "县农业农村局", "2026-05-01", imgJSON(images[0]), false, 278},
		{ "农村危房改造补助标准", "对符合条件的农村低收入群体危房改造，按C级2万元、D级3.5万元标准补助。", "comprehensive", "全县", "县住建局", "2026-06-01", imgJSON(images[4]), false, 334},
	}

	var ids []uint
	for _, d := range defs {
		var count int64
		db.Model(&model.PolicyNotice{}).Where("title = ?", d.title).Count(&count)
		if count > 0 {
			var existing model.PolicyNotice
			db.Where("title = ?", d.title).First(&existing)
			ids = append(ids, existing.ID)
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
			Images:      d.images,
			Views:       d.views,
			IsTop:       d.isTop,
			Status:      1,
		}
		if err := db.Create(&notice).Error; err != nil {
			log.Printf("创建政策公告失败 %s: %v\n", d.title, err)
		} else {
			ids = append(ids, notice.ID)
			log.Printf("已创建政策公告: %s\n", d.title)
		}
	}
	return ids
}

func seedRuralAffairs(db *gorm.DB, users map[string]uint, processors map[string]uint, images []string) {
	processTime := now.Add(-5 * 24 * time.Hour)
	completeTime := now.Add(-2 * 24 * time.Hour)
	firstResponse := now.Add(-4 * 24 * time.Hour)

	type affairDef struct {
		submitter  string
		handler    string
		title      string
		content    string
		typ        string
		address    string
		images     string
		status     int
		process    string
		procImages string
	}

	defs := []affairDef{
		{"yanghm", "liucl",  "永安村3组道路破损", "村道多处坑洼，雨天积水严重，影响村民出行和农产品运输，请尽快修缮。", "infrastructure", "河东镇永安村3组", imgJSON(images[6]), 4, "已组织施工队填补坑洼，预计3日内完成。", imgJSON(images[6])},
		{"wuxl", "chenxh",  "和平村灌溉渠堵塞", "主灌溉渠被杂草淤泥堵塞，下游200亩农田灌溉受阻。", "hardware", "河西镇和平村", imgJSON(images[10]), 4, "已安排清淤作业，疏通完成并检查闸门。", imgJSON(images[10])},
		{"zhaojj", "liucl",  "新丰村垃圾桶损坏", "村口垃圾分类站3个垃圾桶损坏，垃圾露天堆放影响环境卫生。", "env", "河东镇新丰村", imgJSON(images[5]), 5, "已更换新垃圾桶并加强日常清运。", imgJSON(images[5])},
		{"yanghm", "chenxh",  "向阳村电线杆倾斜", "村口电线杆倾斜，大风天气存在倒伏风险，请电力部门处理。", "safety", "河西镇向阳村", imgJSON(images[6]), 5, "已联系电力公司更换电杆，消除安全隐患。", imgJSON(images[6])},
		{"wuxl", "liucl",  "永安村路灯不亮", "主干道5盏太阳能路灯连续一周不亮，夜间出行不便。", "hardware", "河东镇永安村", imgJSON(images[6]), 2, "", ""},
		{"zhaojj", "chenxh",  "和平村水渠护坡坍塌", "暴雨后水渠护坡局部坍塌，需加固防止进一步损坏。", "infrastructure", "河西镇和平村", imgJSON(images[10]), 2, "", ""},
		{"yanghm", "liucl",  "新丰村农药包装废弃物回收点缺失", "村内缺少农药包装废弃物回收点，农户随意丢弃影响环境。", "env", "河东镇新丰村", imgJSON(images[5]), 1, "", ""},
		{"wuxl", "chenxh",  "向阳村广场健身器材损坏", "村文化广场健身器材老化损坏，存在安全隐患。", "safety", "河西镇向阳村", imgJSON(images[7]), 1, "", ""},
	}

	for _, d := range defs {
		var count int64
		db.Model(&model.RuralAffair{}).Where("title = ?", d.title).Count(&count)
		if count > 0 {
			continue
		}

		var handler model.User
		db.Where("username = ?", d.handler).First(&handler)

		affair := model.RuralAffair{
			Title:    d.title,
			Content:  d.content,
			Type:     d.typ,
			Address:  d.address,
			Images:   d.images,
			UserID:   users[d.submitter],
			Status:   d.status,
			AuditTime: &auditAt,
			AuditName: "admin",
		}

		if d.status >= 4 {
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
		} else {
			log.Printf("已创建农村事务: %s (status=%d)\n", d.title, d.status)
		}
	}
}

func seedComments(db *gorm.DB, users map[string]uint, ruralInfoIDs, policyIDs []uint) {
	// 知识评论
	var knowledgeIDs []uint
	db.Table("knowledges").Where("user_id IN ?", []uint{users["wangzj"], users["zhaomy"], users["sunjy"]}).Pluck("id", &knowledgeIDs)
	for i, kid := range knowledgeIDs {
		commenter := users["yanghm"]
		if i%2 == 1 {
			commenter = users["wuxl"]
		}
		content :=  "非常实用的经验分享，学到了！"
		if i%3 == 0 {
			content =  "请问有没有更详细的操作视频？"
		}
		var count int64
		db.Table("comment_knowledges").Where("knowledge_id = ? AND content = ?", kid, content).Count(&count)
		if count > 0 {
			continue
		}
		db.Table("comment_knowledges").Create(map[string]interface{}{
			"user_id":      commenter,
			"knowledge_id": kid,
			"content":      content,
			"parent_id":    0,
			"level":        1,
			"likes":        i + 1,
			"created_at":   now,
		})
	}

	// 动态评论
	var postIDs []uint
	db.Table("posts").Where("user_id IN ?", []uint{users["yanghm"], users["wuxl"], users["zhaojj"], users["lina"], users["zhangwei"]}).Pluck("id", &postIDs)
	for i, pid := range postIDs {
		content :=  "说得真好，深有同感！"
		var count int64
		db.Table("comment_posts").Where("post_id = ? AND content = ?", pid, content).Count(&count)
		if count > 0 {
			continue
		}
		cid := map[string]interface{}{
			"user_id":    users["zhaojj"],
			"post_id":    pid,
			"content":    content,
			"parent_id":  0,
			"level":      1,
			"likes":      i,
			"created_at": now,
		}
		db.Table("comment_posts").Create(cid)
		// 回复评论
		if i < 5 {
			reply :=  "谢谢支持，一起加油！"
			db.Table("comment_posts").Create(map[string]interface{}{
				"user_id":          users["yanghm"],
				"post_id":          pid,
				"content":          reply,
				"parent_id":        0,
				"level":            2,
				"reply_to_user_id": users["zhaojj"],
				"likes":            1,
				"created_at":       now,
			})
		}
	}

	// 商品评论
	var productIDs []uint
	db.Table("products").Where("publisher IN ?", []uint{users["zhangwei"], users["lina"]}).Pluck("id", &productIDs)
	comments := []string{
 "品质很好，下次还会回购！",
 "包装严实，物流很快。",
 "口感不错，家人都喜欢。",
 "价格实惠，性价比高。",
 "新鲜度很高，推荐购买。",
	}
	for i, pid := range productIDs {
		content := comments[i%len(comments)]
		var count int64
		db.Table("comment_products").Where("product_id = ? AND content = ?", pid, content).Count(&count)
		if count > 0 {
			continue
		}
		db.Table("comment_products").Create(map[string]interface{}{
			"user_id":    users["yanghm"],
			"product_id": pid,
			"content":    content,
			"parent_id":  0,
			"level":      1,
			"likes":      i + 2,
			"created_at": now,
		})
	}

	// 农村信息与政策评论
	for i, rid := range ruralInfoIDs {
		content :=  "信息很全面，对我们村很有参考价值。"
		var count int64
		db.Model(&model.CommentRuralInfo{}).Where("target_type = ? AND target_id = ? AND content = ?", "rural_info", rid, content).Count(&count)
		if count > 0 {
			continue
		}
		db.Create(&model.CommentRuralInfo{
			UserID:     users["wuxl"],
			TargetType: "rural_info",
			TargetID:   rid,
			Content:    content,
			ParentID:   0,
			Level:      1,
			Likes:      i + 1,
		})
		if i < 4 {
			db.Create(&model.CommentRuralInfo{
				UserID:         users["zhangwei"],
				TargetType:     "rural_info",
				TargetID:       rid,
				Content:         "欢迎来我们村实地考察！",
				ParentID:       0,
				Level:          2,
				ReplyToUserID:  users["wuxl"],
				Likes:          2,
			})
		}
	}

	for i, pid := range policyIDs {
		content :=  "政策很及时，正好需要了解这方面的信息。"
		var count int64
		db.Model(&model.CommentRuralInfo{}).Where("target_type = ? AND target_id = ? AND content = ?", "policy_notice", pid, content).Count(&count)
		if count > 0 {
			continue
		}
		db.Create(&model.CommentRuralInfo{
			UserID:     users["zhaojj"],
			TargetType: "policy_notice",
			TargetID:   pid,
			Content:    content,
			ParentID:   0,
			Level:      1,
			Likes:      i + 3,
		})
	}

	log.Println("已创建各类评论数据")
}
