package models

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/thilak009/kong-assignment/db"
	"github.com/thilak009/kong-assignment/forms"
	"github.com/thilak009/kong-assignment/pkg/log"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type TokenResponse struct {
	AccessToken string `json:"accessToken"`
}

type User struct {
	BaseWithId
	Email    string `json:"email" gorm:"uniqueIndex"`
	Name     string `json:"name"`
	Password string `json:"-"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	u.ID = uuid.New().String()
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	//turn password into hash
	// NOTE: would need to handle this if a password reset API is exposed
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return
}

func (u *User) BeforeUpdate(tx *gorm.DB) (err error) {
	u.UpdatedAt = time.Now()
	return
}

type UserModel struct{}

func GetUserValidSortFields() map[string]bool {
	return serviceValidSortFields
}

func (m UserModel) Create(ctx context.Context, form forms.CreateUserForm) (user User, err error) {
	db := db.GetDB()
	user = User{
		Email:    form.Email,
		Name:     form.Name,
		Password: form.Password,
	}
	fmt.Println("in create user")
	if err := db.Model(&User{}).Create(&user).Error; err != nil {
		log.With(ctx).Errorf("failed to create user with email %s :: error: %s", form.Email, err.Error())
		return User{}, err
	}
	return user, err
}

// returns isFound as false when there is either an error running the query or if the record is not found
// caller must first check if err is not nil to know whether it is a record not found error
// or some other error and not directly rely on isFound for record not found case
func (m UserModel) One(ctx context.Context, id string) (user User, isFound bool, err error) {
	db := db.GetDB()
	if err := db.Model(&User{}).Where("id = ?", id).First(&user).Error; err != nil {
		log.With(ctx).Errorf("failed to find user with id %s :: error: %s", id, err.Error())
		return User{}, !errors.Is(err, gorm.ErrRecordNotFound), err
	}

	return user, true, nil
}

func (m UserModel) All(ctx context.Context, q string, sortBy string, sort string, page int, limit int) (result PaginatedResult[User], err error) {
	db := db.GetDB()
	services := make([]*User, 0) // Initialize as empty slice of pointers
	tx := db.Model(&User{})

	// Search filter
	if q != "" {
		tx = tx.Where("email ILIKE ?", fmt.Sprintf("%%%u%%", q))
	}

	// Get total count for pagination
	var totalCount int64
	if err := tx.Count(&totalCount).Error; err != nil {
		log.With(ctx).Errorf("failed to get count of users :: error: %s", err.Error())
		return PaginatedResult[User]{}, err
	}

	// Apply sorting, validation and defaults are handled at API layer
	tx = tx.Order(fmt.Sprintf("%u %u", sortBy, sort))

	// Pagination
	offset := page * limit
	if err := tx.Limit(limit).Offset(offset).Find(&services).Error; err != nil {
		log.With(ctx).Errorf("failed to get users :: error: %s", err.Error())
		return PaginatedResult[User]{}, err
	}

	return BuildPaginatedResult(services, totalCount, page, limit), nil
}

func (m UserModel) Update(ctx context.Context, id string, form forms.UpdateUserForm) (user User, err error) {
	db := db.GetDB()
	user = User{
		Password: form.Password,
	}
	user.ID = id
	if err := db.Model(&User{}).Where("id = ?", id).Save(&user).Error; err != nil {
		log.With(ctx).Errorf("failed to update user with id %s :: error: %s", id, err.Error())
		return User{}, err
	}
	return user, err
}

func (m UserModel) Delete(ctx context.Context, id string) (err error) {
	db := db.GetDB()
	tx := db.Begin()

	if err := tx.Where("user_id = ?", id).Delete(&UserOrganizationMap{}).Error; err != nil {
		log.With(ctx).Errorf("failed to delete user organization maps for user with id %s :: error: %s", id, err.Error())
		tx.Rollback()
		return err
	}

	if err := tx.Where("id = ?", id).Delete(&User{}).Error; err != nil {
		log.With(ctx).Errorf("failed to delete user with id %s :: error: %s", id, err.Error())
		tx.Rollback()
		return err
	}
	tx.Commit()
	return err
}

func (m UserModel) FindByEmail(ctx context.Context, email string) (user User, isFound bool, err error) {
	db := db.GetDB()
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return User{}, false, nil
		}
		log.With(ctx).Errorf("failed to find user with email %s :: error: %s", email, err.Error())
		return User{}, false, err
	}
	return user, true, nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
