package convertor

import (
	"time"

	"go-video/ddd/user/domain/entity"
	"go-video/ddd/user/infrastructure/database/po"
)

// UserConvertor 用户转换器
type UserConvertor struct{}

// NewUserConvertor 创建用户转换器
func NewUserConvertor() *UserConvertor {
	return &UserConvertor{}
}

// ToEntity 将PO转换为实体
func (c *UserConvertor) ToEntity(userPO *po.UserPO) *entity.User {
	if userPO == nil {
		return nil
	}

	user := &entity.User{}
	user.SetID(userPO.Id)
	user.SetUUID(userPO.UUID)
	user.SetEmail(userPO.Email)
	user.SetHashedPassword(userPO.Password)
	user.SetStatus(entity.UserStatus(userPO.Status))

	// 设置时间戳
	var deletedAt *time.Time
	if userPO.IsDeleted > 0 {
		// 如果IsDeleted > 0，表示已删除
		now := time.Now()
		deletedAt = &now
	}
	var createdAt, updatedAt time.Time
	if userPO.CreatedAt != nil {
		createdAt = *userPO.CreatedAt
	}
	if userPO.UpdatedAt != nil {
		updatedAt = *userPO.UpdatedAt
	}
	user.SetTimestamps(createdAt, updatedAt, deletedAt)

	// 设置昵称和头像
	if userPO.Nickname != "" {
		user.SetNickname(userPO.Nickname)
	}
	if userPO.Avatar != "" {
		user.SetAvatar(userPO.Avatar)
	}

	return user
}

// ToPO 将实体转换为PO
func (c *UserConvertor) ToPO(user *entity.User) *po.UserPO {
	if user == nil {
		return nil
	}

	userPO := &po.UserPO{
		BaseModel: po.BaseModel{
			Id: user.ID(),
		},
		UUID:     user.UUID(),
		Username: user.Username(),
		Password: user.GetHashedPassword(),
		Email:    user.Email(),
		Nickname: user.Nickname(),
		Avatar:   user.Avatar(),
		Status:   int(user.Status()),
	}

	// 设置时间戳
	createdAt := user.CreatedAt()
	updatedAt := user.UpdatedAt()
	userPO.CreatedAt = &createdAt
	userPO.UpdatedAt = &updatedAt

	// 设置删除标记
	if user.DeletedAt() != nil {
		userPO.IsDeleted = 1
	} else {
		userPO.IsDeleted = 0
	}

	return userPO
}

// ToEntities 将PO列表转换为实体列表
func (c *UserConvertor) ToEntities(userPOs []*po.UserPO) []*entity.User {
	if userPOs == nil {
		return nil
	}

	users := make([]*entity.User, 0, len(userPOs))
	for _, userPO := range userPOs {
		if user := c.ToEntity(userPO); user != nil {
			users = append(users, user)
		}
	}

	return users
}

// ToPOs 将实体列表转换为PO列表
func (c *UserConvertor) ToPOs(users []*entity.User) []*po.UserPO {
	if users == nil {
		return nil
	}

	userPOs := make([]*po.UserPO, 0, len(users))
	for _, user := range users {
		if userPO := c.ToPO(user); userPO != nil {
			userPOs = append(userPOs, userPO)
		}
	}

	return userPOs
}