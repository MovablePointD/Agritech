package config

import (
	"fmt"
	"log"

	"rxtcloud/common/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	dsn := "root:root@tcp(127.0.0.1:3306)/gorxt_db?charset=utf8mb4&parseTime=True&loc=Local"

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true, // 禁用外键约束自动迁移
	})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	fmt.Println("数据库连接成功")

	// 自动迁移：确保Go模型与数据库表结构同步
	if err := DB.AutoMigrate(
		&model.User{},
		&model.Address{},
		&model.Expert{},
		&model.RuralAffair{},
		&model.AffairProcessor{},
		&model.AffairModification{},
		&model.AffairFollowUp{},
		&model.AffairAppeal{},
		&model.AffairAppealMaterial{},
		&model.RuralInfo{},
		&model.PolicyNotice{},
		&model.CommentRuralInfo{},
		&model.SensitiveWord{},
		&model.Like{},
		&model.Notification{},
	); err != nil {
		log.Fatal("数据库自动迁移失败:", err)
	}
	fmt.Println("数据库自动迁移完成")
}
