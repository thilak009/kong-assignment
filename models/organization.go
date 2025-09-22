package models

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/thilak009/kong-assignment/db"
	"github.com/thilak009/kong-assignment/forms"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Organization struct {
	BaseWithId
	Name        string `json:"name" gorm:"index"`
	Description string `json:"description"`
	CreatedBy   string `json:"createdBy"`
	// Relationships
	Creator User `json:"-" gorm:"foreignKey:CreatedBy"`
}

func (o *Organization) BeforeCreate(tx *gorm.DB) (err error) {
	o.ID = uuid.New().String()
	o.CreatedAt = time.Now()
	o.UpdatedAt = time.Now()
	return
}

func (o *Organization) BeforeUpdate(tx *gorm.DB) (err error) {
	o.UpdatedAt = time.Now()
	return
}

// UserOrganizationMap represents the many-to-many relationship
type UserOrganizationMap struct {
	Base
	UserID         string `json:"userId" gorm:"primaryKey"`
	OrganizationID string `json:"organizationId" gorm:"primaryKey"`
}

func (o *UserOrganizationMap) BeforeCreate(tx *gorm.DB) (err error) {
	o.CreatedAt = time.Now()
	o.UpdatedAt = time.Now()
	return
}

func (o *UserOrganizationMap) BeforeUpdate(tx *gorm.DB) (err error) {
	o.UpdatedAt = time.Now()
	return
}

type OrganizationModel struct{}

var organizationValidSortFields = map[string]bool{
	"name":       true,
	"created_at": true,
	"updated_at": true,
}

func GetOrganizationValidSortFields() map[string]bool {
	return organizationValidSortFields
}

func (m OrganizationModel) Create(form forms.CreateOrganizationForm, createdBy string) (organization Organization, err error) {
	db := db.GetDB()

	// Start transaction
	tx := db.Begin()

	organization = Organization{
		Name:        form.Name,
		Description: form.Description,
		CreatedBy:   createdBy,
	}

	if err := tx.Create(&organization).Error; err != nil {
		tx.Rollback()
		return Organization{}, err
	}

	// Add creator to organization
	userOrg := UserOrganizationMap{
		UserID:         createdBy,
		OrganizationID: organization.ID,
	}

	if err := tx.Create(&userOrg).Error; err != nil {
		tx.Rollback()
		return Organization{}, err
	}

	tx.Commit()
	return organization, nil
}

func (m OrganizationModel) One(id string) (organization Organization, isFound bool, err error) {
	db := db.GetDB()
	if err := db.Model(&Organization{}).Where("id = ?", id).First(&organization).Error; err != nil {
		return Organization{}, !errors.Is(err, gorm.ErrRecordNotFound), err
	}
	return organization, true, nil
}

func (m OrganizationModel) GetUserOrganizations(userID string, q string, sortBy string, sort string, page int, limit int) (result PaginatedResult[Organization], err error) {
	db := db.GetDB()
	organizations := make([]*Organization, 0)

	tx := db.Model(&Organization{}).
		Joins("JOIN user_organization_maps ON organizations.id = user_organization_maps.organization_id").
		Where("user_organization_maps.user_id = ?", userID)

	// Search filter
	if q != "" {
		tx = tx.Where("organizations.name ILIKE ?", fmt.Sprintf("%%%s%%", q))
	}

	// Get total count for pagination
	var totalCount int64
	if err := tx.Count(&totalCount).Error; err != nil {
		return PaginatedResult[Organization]{}, err
	}

	// Apply sorting
	tx = tx.Order(fmt.Sprintf("organizations.%s %s", sortBy, sort))

	// Pagination
	offset := page * limit
	if err := tx.Limit(limit).Offset(offset).Find(&organizations).Error; err != nil {
		return PaginatedResult[Organization]{}, err
	}

	return BuildPaginatedResult(organizations, totalCount, page, limit), nil
}

func (m OrganizationModel) Update(id string, form forms.CreateOrganizationForm) (organization Organization, err error) {
	db := db.GetDB()

	if err := db.Model(&Organization{}).Where("id = ?", id).First(&organization).Error; err != nil {
		return Organization{}, err
	}

	organization.Name = form.Name
	organization.Description = form.Description

	if err := db.Save(&organization).Error; err != nil {
		return Organization{}, err
	}

	return organization, nil
}

func (m OrganizationModel) Delete(id string) (err error) {
	db := db.GetDB()

	// Start transaction
	tx := db.Begin()

	// TODO: figure out cascade deletes
	// Delete user-organization relationships
	if err := tx.Where("organization_id = ?", id).Delete(&UserOrganizationMap{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	services := []Service{}
	// Delete org services
	if err := tx.Where("organization_id = ?", id).Clauses(clause.Returning{}).Delete(&services).Error; err != nil {
		tx.Rollback()
		return err
	}

	serviceIds := []string{}
	for _, service := range services {
		serviceIds = append(serviceIds, service.ID)
	}
	// Delete versions of the services
	if err := tx.Where("service_id IN (?)", serviceIds).Delete(&ServiceVersion{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete organization
	if err := tx.Where("id = ?", id).Delete(&Organization{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func (m OrganizationModel) IsUserMember(orgID string, userID string) (bool, error) {
	db := db.GetDB()
	var count int64

	err := db.Model(&UserOrganizationMap{}).
		Where("organization_id = ? AND user_id = ?", orgID, userID).
		Count(&count).Error

	return count > 0, err
}
