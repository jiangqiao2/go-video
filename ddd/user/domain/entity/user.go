package entity

import (
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// User 用户聚合根
type User struct {
	id        uint64
	uuid      string
	username  string
	password  string
	email     string
	nickname  string
	avatar    string
	status    UserStatus
	createdAt time.Time
	updatedAt time.Time
	deletedAt *time.Time
}

// UserStatus 用户状态值对象
type UserStatus int

const (
	UserStatusActive   UserStatus = 1 // 激活
	UserStatusInactive UserStatus = 2 // 未激活
	UserStatusDisabled UserStatus = 3 // 禁用
	UserStatusDeleted  UserStatus = 4 // 删除
)

// NewUser 创建新用户
func NewUser(username, password, email string) (*User, error) {
	user := &User{
		uuid:      uuid.New().String(),
		username:  username,
		email:     email,
		status:    UserStatusInactive,
		createdAt: time.Now(),
		updatedAt: time.Now(),
	}

	// 验证用户名
	if err := user.ValidateUsername(); err != nil {
		return nil, err
	}

	// 验证邮箱
	if err := user.ValidateEmail(); err != nil {
		return nil, err
	}

	// 设置密码
	if err := user.SetPassword(password); err != nil {
		return nil, err
	}

	return user, nil
}

// ID 获取用户内部ID（仅用于内部逻辑）
func (u *User) ID() uint64 {
	return u.id
}

// UUID 获取用户UUID（对外标识）
func (u *User) UUID() string {
	return u.uuid
}

// Username 获取用户名
func (u *User) Username() string {
	return u.username
}

// Email 获取邮箱
func (u *User) Email() string {
	return u.email
}

// Nickname 获取昵称
func (u *User) Nickname() string {
	return u.nickname
}

// Avatar 获取头像
func (u *User) Avatar() string {
	return u.avatar
}

// Status 获取状态
func (u *User) Status() UserStatus {
	return u.status
}

// CreatedAt 获取创建时间
func (u *User) CreatedAt() time.Time {
	return u.createdAt
}

// UpdatedAt 获取更新时间
func (u *User) UpdatedAt() time.Time {
	return u.updatedAt
}

// DeletedAt 获取删除时间
func (u *User) DeletedAt() *time.Time {
	return u.deletedAt
}

// SetID 设置用户内部ID（仅用于从数据库加载）
func (u *User) SetID(id uint64) {
	u.id = id
}

// SetUUID 设置用户UUID（仅用于从数据库加载）
func (u *User) SetUUID(uuid string) {
	u.uuid = uuid
}

// SetNickname 设置昵称
func (u *User) SetNickname(nickname string) error {
	if len(nickname) > 50 {
		return errors.New("昵称长度不能超过50个字符")
	}
	u.nickname = nickname
	u.updatedAt = time.Now()
	return nil
}

// SetAvatar 设置头像
func (u *User) SetAvatar(avatar string) {
	u.avatar = avatar
	u.updatedAt = time.Now()
}

// SetPassword 设置密码
func (u *User) SetPassword(password string) error {
	if err := u.ValidatePassword(password); err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.password = string(hashedPassword)
	u.updatedAt = time.Now()
	return nil
}

// VerifyPassword 验证密码
func (u *User) VerifyPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.password), []byte(password))
	return err == nil
}

// Activate 激活用户
func (u *User) Activate() {
	u.status = UserStatusActive
	u.updatedAt = time.Now()
}

// Disable 禁用用户
func (u *User) Disable() {
	u.status = UserStatusDisabled
	u.updatedAt = time.Now()
}

// Delete 软删除用户
func (u *User) Delete() {
	now := time.Now()
	u.status = UserStatusDeleted
	u.deletedAt = &now
	u.updatedAt = now
}

// IsActive 检查用户是否激活
func (u *User) IsActive() bool {
	return u.status == UserStatusActive
}

// IsDeleted 检查用户是否已删除
func (u *User) IsDeleted() bool {
	return u.status == UserStatusDeleted || u.deletedAt != nil
}

// ValidateUsername 验证用户名
func (u *User) ValidateUsername() error {
	if len(u.username) < 3 || len(u.username) > 20 {
		return errors.New("用户名长度必须在3-20个字符之间")
	}

	// 只允许字母、数字和下划线
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, u.username)
	if !matched {
		return errors.New("用户名只能包含字母、数字和下划线")
	}

	return nil
}

// ValidatePassword 验证密码强度
func (u *User) ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("密码长度不能少于8位")
	}

	if len(password) > 128 {
		return errors.New("密码长度不能超过128位")
	}

	// 检查是否包含字母和数字
	hasLetter, _ := regexp.MatchString(`[a-zA-Z]`, password)
	hasDigit, _ := regexp.MatchString(`[0-9]`, password)

	if !hasLetter || !hasDigit {
		return errors.New("密码必须包含字母和数字")
	}

	return nil
}

// ValidateEmail 验证邮箱格式
func (u *User) ValidateEmail() error {
	if u.email == "" {
		return errors.New("邮箱不能为空")
	}

	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(emailRegex, u.email)
	if !matched {
		return errors.New("邮箱格式无效")
	}

	return nil
}

// GetHashedPassword 获取加密后的密码（仅用于持久化）
func (u *User) GetHashedPassword() string {
	return u.password
}

// SetHashedPassword 设置加密后的密码（仅用于从数据库加载）
func (u *User) SetHashedPassword(hashedPassword string) {
	u.password = hashedPassword
}

// SetTimestamps 设置时间戳（仅用于从数据库加载）
func (u *User) SetTimestamps(createdAt, updatedAt time.Time, deletedAt *time.Time) {
	u.createdAt = createdAt
	u.updatedAt = updatedAt
	u.deletedAt = deletedAt
}

// SetStatus 设置状态（仅用于从数据库加载）
func (u *User) SetStatus(status UserStatus) {
	u.status = status
}

// SetEmail 设置邮箱（仅用于从数据库加载）
func (u *User) SetEmail(email string) {
	u.email = email
}