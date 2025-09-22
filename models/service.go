package models

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/thilak009/kong-assignment/db"
	"github.com/thilak009/kong-assignment/forms"
	"gorm.io/gorm"
)

type Service struct {
	BaseWithId
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	OrganizationID string          `json:"organization_id"`
	Metadata       ServiceMetadata `json:"metadata" gorm:"-"`
	// Relationships
	Organization Organization `json:"-" gorm:"foreignKey:OrganizationID"`
}

type ServiceMetadata struct {
	VersionCount *int `json:"versionCount,omitempty"`
}

func (s *Service) BeforeCreate(tx *gorm.DB) (err error) {
	s.ID = uuid.New().String()
	s.CreatedAt = time.Now()
	s.UpdatedAt = time.Now()
	return
}

func (s *Service) BeforeUpdate(tx *gorm.DB) (err error) {
	s.UpdatedAt = time.Now()
	return
}

type ServiceModel struct{}

var serviceValidSortFields = map[string]bool{
	"name":       true,
	"created_at": true,
	"updated_at": true,
}

func GetServiceValidSortFields() map[string]bool {
	return serviceValidSortFields
}

func (m ServiceModel) Create(form forms.CreateServiceForm, organizationID string) (service Service, err error) {
	db := db.GetDB()
	service = Service{
		Name:           form.Name,
		Description:    form.Description,
		OrganizationID: organizationID,
	}
	if err := db.Model(&Service{}).Create(&service).Error; err != nil {
		return Service{}, err
	}
	return service, err
}

// returns isFound as false when there is either an error running the query or if the record is not found
// caller must first check if err is not nil to know whether it is a record not found error
// or some other error and not directly rely on isFound for record not found case
func (m ServiceModel) One(id string, organizationID string, includeVersionCount bool) (service Service, isFound bool, err error) {
	db := db.GetDB()
	if err := db.Model(&Service{}).Where("id = ? AND organization_id = ?", id, organizationID).First(&service).Error; err != nil {
		return Service{}, !errors.Is(err, gorm.ErrRecordNotFound), err
	}

	// Populate version count (only if requested)
	if includeVersionCount {
		var versionCount int64
		if err := db.Model(&ServiceVersion{}).Where("service_id = ?", service.ID).Count(&versionCount).Error; err != nil {
			return Service{}, true, err
		}
		versionCountInt := int(versionCount)
		service.Metadata.VersionCount = &versionCountInt
	}

	return service, true, nil
}

func (m ServiceModel) All(organizationID string, q string, sortBy string, sort string, page int, limit int, includeVersionCount bool) (result PaginatedResult[Service], err error) {
	db := db.GetDB()
	services := make([]*Service, 0) // Initialize as empty slice of pointers
	tx := db.Model(&Service{}).Where("organization_id = ?", organizationID)

	// Search filter
	if q != "" {
		tx = tx.Where("name ILIKE ?", fmt.Sprintf("%%%s%%", q))
	}

	// Get total count for pagination
	var totalCount int64
	if err := tx.Count(&totalCount).Error; err != nil {
		return PaginatedResult[Service]{}, err
	}

	// Apply sorting, validation and defaults are handled at API layer
	tx = tx.Order(fmt.Sprintf("%s %s", sortBy, sort))

	// Pagination
	offset := page * limit
	if err := tx.Limit(limit).Offset(offset).Find(&services).Error; err != nil {
		return PaginatedResult[Service]{}, err
	}

	// Populate version counts for all services efficiently (only if requested)
	if includeVersionCount && len(services) > 0 {
		// Get all service IDs
		serviceIds := make([]string, len(services))
		for i, service := range services {
			serviceIds[i] = service.ID
		}

		// Single query to get all version counts
		type VersionCount struct {
			ServiceID string `gorm:"column:service_id"`
			Count     int64  `gorm:"column:count"`
		}
		// went with this approach as the API is paginated and wouldn't have too many IDs in the IN clause
		var versionCounts []VersionCount
		if err := db.Model(&ServiceVersion{}).
			Select("service_id, COUNT(*) as count").
			Where("service_id IN ?", serviceIds).
			Group("service_id").
			Scan(&versionCounts).Error; err != nil {
			return PaginatedResult[Service]{}, err
		}

		countMap := make(map[string]int64)
		for _, vc := range versionCounts {
			countMap[vc.ServiceID] = vc.Count
		}

		// Assign counts to services
		for _, service := range services {
			versionCountInt := int(countMap[service.ID])
			service.Metadata.VersionCount = &versionCountInt
		}
	}

	return BuildPaginatedResult(services, totalCount, page, limit), nil
}

func (m ServiceModel) Update(id string, organizationID string, form forms.CreateServiceForm) (service Service, err error) {
	db := db.GetDB()

	// First check if service exists and belongs to organization
	if err := db.Model(&Service{}).Where("id = ? AND organization_id = ?", id, organizationID).First(&service).Error; err != nil {
		return Service{}, err
	}

	// Update the service
	service.Name = form.Name
	service.Description = form.Description

	if err := db.Save(&service).Error; err != nil {
		return Service{}, err
	}
	return service, err
}

func (m ServiceModel) Delete(id string, organizationID string) (err error) {
	db := db.GetDB()
	tx := db.Begin()
	if err := tx.Where("service_id = ?", id).Delete(&ServiceVersion{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("id = ? AND organization_id = ?", id, organizationID).Delete(&Service{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return err
}
