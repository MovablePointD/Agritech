module rxtcloud/knowledge-service

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/nacos-group/nacos-sdk-go/v2 v2.2.5
	gorm.io/driver/mysql v1.5.4
	gorm.io/gorm v1.25.7
	rxtcloud/common v0.0.0
)

replace rxtcloud/common => ../common
