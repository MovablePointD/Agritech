package config

import (
	"fmt"
	"log"

	"rxtcloud/message-service/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB(dsn string) *gorm.DB {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		log.Fatalf("[Message-Service] 数据库连接失败: %v", err)
	}
	DB = db
	fmt.Println("[Message-Service] 数据库连接成功")

	// 自动迁移：确保Go模型与数据库表结构同步
	if err := db.AutoMigrate(
		&model.Conversation{},
		&model.Message{},
		&model.Notification{},
		&model.Like{},
	); err != nil {
		log.Fatalf("[Message-Service] 数据库自动迁移失败: %v", err)
	}
	fmt.Println("[Message-Service] 数据库自动迁移完成")
	return db
}
