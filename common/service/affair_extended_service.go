package service

import (
	"errors"
	"fmt"
	"rxtcloud/common/config"
	"rxtcloud/common/model"
	"time"
)

// ============================================================
// 一、事务修改（Modify）
// 用户可在 status=1(待审核) 时修改事务详情
// 每次修改会保存完整变更记录到 AffairModification 表
// ============================================================

// ModifyRuralAffair 用户修改待审核状态的事务，保存修改前后对比记录
// 权限：仅事务提交者 + status=1(待审核) 时可修改
func ModifyRuralAffair(affairID uint, userID uint, userName string, updateData model.RuralAffair) error {
	var affair model.RuralAffair
	if err := config.DB.First(&affair, affairID).Error; err != nil {
		return errors.New("事务不存在")
	}

	// 只有事务提交者可以修改
	if affair.UserID != userID {
		return errors.New("只有事务提交者可以修改")
	}

	// 只有状态=1(待审核)时可以修改
	if affair.Status != 1 {
		return fmt.Errorf("当前状态为「%s」，不允许修改", model.AffairStatusText[affair.Status])
	}

	// 保存修改记录
	modification := model.AffairModification{
		AffairID:    affairID,
		UserID:      userID,
		UserName:    userName,
		OldTitle:    affair.Title,
		NewTitle:    updateData.Title,
		OldContent:  affair.Content,
		NewContent:  updateData.Content,
		OldAddress:  affair.Address,
		NewAddress:  updateData.Address,
		OldImages:   affair.Images,
		NewImages:   updateData.Images,
		OldVideoURL: affair.VideoURL,
		NewVideoURL: updateData.VideoURL,
	}
	if err := config.DB.Create(&modification).Error; err != nil {
		return errors.New("保存修改记录失败")
	}

	// 更新事务字段
	now := time.Now()
	updates := map[string]interface{}{
		"last_modified_at": now,
	}
	if updateData.Title != "" {
		updates["title"] = updateData.Title
	}
	if updateData.Content != "" {
		updates["content"] = updateData.Content
	}
	if updateData.Address != "" {
		updates["address"] = updateData.Address
	}
	if updateData.Images != "" {
		updates["images"] = updateData.Images
	}
	if updateData.VideoURL != "" {
		updates["video_url"] = updateData.VideoURL
	}
	updates["type"] = updateData.Type

	return config.DB.Model(&affair).Updates(updates).Error
}

// GetAffairModifications 获取事务的修改记录列表
func GetAffairModifications(affairID uint) ([]model.AffairModification, error) {
	var modifications []model.AffairModification
	err := config.DB.Where("affair_id = ?", affairID).
		Order("created_at desc").
		Find(&modifications).Error
	return modifications, err
}

// GetLastModificationTime 获取事务最后一次修改时间
func GetLastModificationTime(affairID uint) (*time.Time, error) {
	var affair model.RuralAffair
	if err := config.DB.First(&affair, affairID).Error; err != nil {
		return nil, err
	}
	return affair.LastModifiedAt, nil
}

// ============================================================
// 二、追问/追答（Follow-Up）
// 处理报告提交后，用户可追问 → 处理人员追答 → 用户再追问...
// 状态切换：追问(status 4→6) ↔ 追答(status 6→4)
// ============================================================

// AddFollowUpQuestion 用户添加追问（status=4/6，需已有处理记录），追问后事务状态→6
func AddFollowUpQuestion(affairID uint, userID uint, userName string, content string, images string) (*model.AffairFollowUp, error) {
	var affair model.RuralAffair
	if err := config.DB.First(&affair, affairID).Error; err != nil {
		return nil, errors.New("事务不存在")
	}

	// 只有事务提交者可以追问
	if affair.UserID != userID {
		return nil, errors.New("只有事务提交者可以追问")
	}

	// 必须在处理中(4)或追问中(6)状态，且已经有处理记录
	if affair.Status != 4 && affair.Status != 6 {
		return nil, fmt.Errorf("当前状态为「%s」，不允许追问", model.AffairStatusText[affair.Status])
	}

	if affair.ProcessContent == "" {
		return nil, errors.New("处理人员尚未提交处理记录，暂不可追问")
	}

	followUp := model.AffairFollowUp{
		AffairID:  affairID,
		Type:      "question",
		UserID:    userID,
		UserName:  userName,
		Content:   content,
		Images:    images,
		CreatedAt: time.Now(),
	}

	if err := config.DB.Create(&followUp).Error; err != nil {
		return nil, errors.New("添加追问失败")
	}

	// 将事务状态改为追问中
	config.DB.Model(&affair).Update("status", 6)

	// 通知事务处理人员：事务被追问
	NotifyAffairFollowUp(&affair)

	return &followUp, nil
}

// AddFollowUpAnswer 处理人员追答（status=6时），追答后事务状态恢复→4
func AddFollowUpAnswer(affairID uint, handlerID uint, handlerName string, content string, images string) (*model.AffairFollowUp, error) {
	var affair model.RuralAffair
	if err := config.DB.First(&affair, affairID).Error; err != nil {
		return nil, errors.New("事务不存在")
	}

	// 只有当前处理人员可以追答
	if affair.HandlerID != handlerID {
		return nil, errors.New("只有当前事务处理人员可以追答")
	}

	// 必须在追问中(6)状态
	if affair.Status != 6 {
		return nil, fmt.Errorf("当前状态为「%s」，不允许追答", model.AffairStatusText[affair.Status])
	}

	followUp := model.AffairFollowUp{
		AffairID:  affairID,
		Type:      "answer",
		UserID:    handlerID,
		UserName:  handlerName,
		Content:   content,
		Images:    images,
		CreatedAt: time.Now(),
	}

	if err := config.DB.Create(&followUp).Error; err != nil {
		return nil, errors.New("添加追答失败")
	}

	// 追答后将状态恢复为处理中（用户可以继续追问或确认完成）
	config.DB.Model(&affair).Update("status", 4)

	// 通知事务提交者：追问已回复
	NotifyAffairFollowUpAnswered(&affair)

	return &followUp, nil
}

// GetAffairFollowUps 获取事务的追问追答列表
func GetAffairFollowUps(affairID uint) ([]model.AffairFollowUp, error) {
	var followUps []model.AffairFollowUp
	err := config.DB.Where("affair_id = ?", affairID).
		Order("created_at asc").
		Find(&followUps).Error
	return followUps, err
}

// ============================================================
// 三、申诉（Appeal）
// 用户申诉：已完成(status=5)的事务，由用户发起
// 处理人员申诉：处理中(status=4/6)且有首次处理记录，由处理人员发起
// 申诉由 admin/sysadmin 处理，处理期间事务状态→7(申诉中)
// 同一事务同时只允许一个进行中的申诉
// ============================================================

// CreateAppeal 创建申诉（applicantType: "user"/"handler"），检查是否有进行中的申诉防重复
func CreateAppeal(affairID uint, applicantType string, applicantID uint, applicantName string, reason string, images string) (*model.AffairAppeal, error) {
	var affair model.RuralAffair
	if err := config.DB.First(&affair, affairID).Error; err != nil {
		return nil, errors.New("事务不存在")
	}

	// 权限检查
	switch applicantType {
	case "user":
		if affair.UserID != applicantID {
			return nil, errors.New("只有事务提交者可以发起用户申诉")
		}
		if affair.Status != 5 {
			return nil, fmt.Errorf("当前状态为「%s」，用户只能对已完成的事务发起申诉", model.AffairStatusText[affair.Status])
		}
	case "handler":
		if affair.HandlerID != applicantID {
			return nil, errors.New("只有当前事务处理人员可以发起申诉")
		}
		if affair.Status != 4 && affair.Status != 6 {
			return nil, fmt.Errorf("当前状态为「%s」，处理人员只能对处理中的事务发起申诉", model.AffairStatusText[affair.Status])
		}
		// 必须有首次处理记录
		if affair.ProcessContent == "" {
			return nil, errors.New("暂未提交处理记录，不可发起申诉")
		}
	default:
		return nil, errors.New("无效的申诉人类型")
	}

	// 检查是否已有进行中的申诉
	var existingAppeal model.AffairAppeal
	err := config.DB.Where("affair_id = ? AND status IN ('pending', 'processing')", affairID).
		First(&existingAppeal).Error
	if err == nil {
		return nil, errors.New("该事务已有进行中的申诉，请等待处理完成")
	}

	appeal := model.AffairAppeal{
		AffairID:      affairID,
		ApplicantType: applicantType,
		ApplicantID:   applicantID,
		ApplicantName: applicantName,
		Reason:        reason,
		Images:        images,
		Status:        "pending",
		CreatedAt:     time.Now(),
	}

	if err := config.DB.Create(&appeal).Error; err != nil {
		return nil, errors.New("创建申诉失败")
	}

	// 将事务状态改为申诉中
	config.DB.Model(&affair).Update("status", 7)

	// 通知 admin 和 sysadmin：有新申诉
	NotifyAffairAppeal(&appeal, &affair)

	return &appeal, nil
}

// AddAppealMaterial 提交申诉辅助材料（只有申诉对立方可提交：用户申诉→处理人员提交，反之亦然）
func AddAppealMaterial(appealID uint, submitterType string, submitterID uint, submitterName string, content string, images string) (*model.AffairAppealMaterial, error) {
	var appeal model.AffairAppeal
	if err := config.DB.First(&appeal, appealID).Error; err != nil {
		return nil, errors.New("申诉不存在")
	}

	if appeal.Status != "pending" && appeal.Status != "processing" {
		return nil, errors.New("申诉已处理完成，不可再提交材料")
	}

	// 权限检查：只能由申诉的对立方提交材料
	var affair model.RuralAffair
	config.DB.First(&affair, appeal.AffairID)

	allowed := false
	if submitterType == "handler" && appeal.ApplicantType == "user" {
		allowed = true // 用户申诉 → 处理人员可提交材料
	} else if submitterType == "user" && appeal.ApplicantType == "handler" {
		allowed = true // 处理人员申诉 → 用户可提交材料
	}

	if !allowed {
		return nil, errors.New("您无权为此申诉提交辅助材料，只有申诉对立方可提交")
	}

	material := model.AffairAppealMaterial{
		AppealID:      appealID,
		SubmitterType: submitterType,
		SubmitterID:   submitterID,
		SubmitterName: submitterName,
		Content:       content,
		Images:        images,
		CreatedAt:     time.Now(),
	}

	if err := config.DB.Create(&material).Error; err != nil {
		return nil, errors.New("提交材料失败")
	}

	return &material, nil
}

// ProcessAppeal 管理员处理申诉：设定结果(approve=通过)，恢复事务状态到申诉前
// 用户申诉(approve后)→status=5(已完成)，处理人员申诉→status=4(处理中)
func ProcessAppeal(appealID uint, handlerID uint, handlerName string, result string, approve bool) error {
	var appeal model.AffairAppeal
	if err := config.DB.First(&appeal, appealID).Error; err != nil {
		return errors.New("申诉不存在")
	}

	if appeal.Status != "pending" && appeal.Status != "processing" {
		return errors.New("申诉已处理完成")
	}

	now := time.Now()
	appeal.Status = "resolved"
	appeal.Result = result
	appeal.HandlerID = handlerID
	appeal.HandlerName = handlerName
	appeal.ProcessedAt = &now

	if err := config.DB.Save(&appeal).Error; err != nil {
		return errors.New("处理申诉失败")
	}

	// 恢复事务状态
	var affair model.RuralAffair
	config.DB.First(&affair, appeal.AffairID)

	if appeal.ApplicantType == "user" {
		// 用户申诉后回到已完成状态
		affair.Status = 5
	} else {
		// 处理人员申诉后回到处理中状态
		affair.Status = 4
	}
	config.DB.Save(&affair)

	return nil
}

// GetAffairAppeal 获取事务的申诉（最新一条）
func GetAffairAppeal(affairID uint) (*model.AffairAppeal, error) {
	var appeal model.AffairAppeal
	err := config.DB.Where("affair_id = ?", affairID).
		Order("created_at desc").
		First(&appeal).Error
	if err != nil {
		return nil, err
	}
	return &appeal, nil
}

// GetAppealMaterials 获取申诉的所有辅助材料
func GetAppealMaterials(appealID uint) ([]model.AffairAppealMaterial, error) {
	var materials []model.AffairAppealMaterial
	err := config.DB.Where("appeal_id = ?", appealID).
		Order("created_at asc").
		Find(&materials).Error
	return materials, err
}

// GetAllAppeals 获取所有申诉列表（管理员）
func GetAllAppeals(status string, page, pageSize int) ([]model.AffairAppeal, int64, error) {
	var appeals []model.AffairAppeal
	var total int64

	query := config.DB.Model(&model.AffairAppeal{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Order("created_at desc").
		Offset(offset).
		Limit(pageSize).
		Find(&appeals).Error

	return appeals, total, err
}

// ============================================================
// 四、确认完成
// 用户确认处理结果满意，从处理中(4)或追问中(6)→已完成(5)
// 前提：必须有处理记录(ProcessContent 不为空)
// ============================================================

// ConfirmCompleteRuralAffair 用户确认事务处理完成，状态→5，记录完成时间
func ConfirmCompleteRuralAffair(affairID uint, userID uint) error {
	var affair model.RuralAffair
	if err := config.DB.First(&affair, affairID).Error; err != nil {
		return errors.New("事务不存在")
	}

	// 只有事务提交者可以确认完成
	if affair.UserID != userID {
		return errors.New("只有事务提交者可以确认完成")
	}

	// 处理中(4)或追问中(6)都可以确认完成
	if affair.Status != 4 && affair.Status != 6 {
		return fmt.Errorf("当前状态为「%s」，不允许确认完成", model.AffairStatusText[affair.Status])
	}

	// 必须有处理记录
	if affair.ProcessContent == "" {
		return errors.New("处理人员尚未提交处理记录，暂不可确认完成")
	}

	now := time.Now()
	affair.Status = 5
	affair.CompletedTime = &now

	if err := config.DB.Save(&affair).Error; err != nil {
		return err
	}

	// 通知事务提交者和处理人员：事务已完成
	NotifyAffairCompleted(&affair)

	return nil
}

// ============================================================
// 五、完成事务详情查询（包含所有扩展数据）
// ============================================================

// AffairFullDetail 事务完整详情（包含修改记录、追问、申诉）
type AffairFullDetail struct {
	Affair         *model.RuralAffair            `json:"affair"`
	Modifications  []model.AffairModification    `json:"modifications"`   // 修改记录
	FollowUps      []model.AffairFollowUp        `json:"follow_ups"`      // 追问追答
	Appeal         *model.AffairAppeal           `json:"appeal"`          // 申诉信息
	AppealMaterials []model.AffairAppealMaterial `json:"appeal_materials"` // 申诉辅助材料
	CanModify      bool                          `json:"can_modify"`      // 当前用户是否可以修改
	CanFollowUp    bool                          `json:"can_follow_up"`   // 当前用户是否可以追问
	CanRespond     bool                          `json:"can_respond"`     // 当前处理人员是否可以追答
	CanConfirm     bool                          `json:"can_confirm"`     // 当前用户是否可以确认完成
	CanAppeal      bool                          `json:"can_appeal"`      // 当前用户是否可以申诉
}

// GetAffairFullDetail 获取事务完整详情（包含扩展数据，以及基于 currentUser 的权限判断）
func GetAffairFullDetail(affairID uint, currentUserID uint, currentUserRole string) (*AffairFullDetail, error) {
	affair, err := GetRuralAffair(affairID)
	if err != nil {
		return nil, err
	}

	// 并行获取关联数据
	modifications, _ := GetAffairModifications(affairID)
	followUps, _ := GetAffairFollowUps(affairID)
	appeal, _ := GetAffairAppeal(affairID)
	var appealMaterials []model.AffairAppealMaterial
	if appeal != nil {
		appealMaterials, _ = GetAppealMaterials(appeal.ID)
	}

	// 权限判断
	canModify := false
	if affair.UserID == currentUserID && affair.Status == 1 {
		canModify = true
	}

	canFollowUp := false
	if affair.UserID == currentUserID && (affair.Status == 4 || affair.Status == 6) && affair.ProcessContent != "" {
		canFollowUp = true
	}

	canRespond := false
	if affair.HandlerID == currentUserID && affair.Status == 6 {
		canRespond = true
	}

	canConfirm := false
	if affair.UserID == currentUserID && (affair.Status == 4 || affair.Status == 6) && affair.ProcessContent != "" {
		canConfirm = true
	}

	isAdmin := currentUserRole == "admin" || currentUserRole == "sysadmin"
	canAppeal := false
	if affair.Status == 5 && affair.UserID == currentUserID {
		canAppeal = true // 用户对已完成事务发起申诉
	}
	if (affair.Status == 4 || affair.Status == 6) && affair.HandlerID == currentUserID && affair.ProcessContent != "" {
		canAppeal = true // 处理人员对处理中且有首次记录的事务发起申诉
	}

	_ = isAdmin // admin can process appeals

	return &AffairFullDetail{
		Affair:          affair,
		Modifications:   modifications,
		FollowUps:       followUps,
		Appeal:          appeal,
		AppealMaterials: appealMaterials,
		CanModify:       canModify,
		CanFollowUp:     canFollowUp,
		CanRespond:      canRespond,
		CanConfirm:      canConfirm,
		CanAppeal:       canAppeal,
	}, nil
}
