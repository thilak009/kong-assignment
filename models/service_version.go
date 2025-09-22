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
	"gorm.io/gorm"
)

type ServiceVersion struct {
	BaseWithId
	Version          string    `json:"version" gorm:"uniqueIndex:idx_service_version"`
	Description      string    `json:"description"`
	ReleaseTimestamp time.Time `json:"releaseTimestamp"`
	ServiceID        string    `json:"serviceId" gorm:"uniqueIndex:idx_service_version"`
	Service          Service   `gorm:"foreignKey:ServiceID" json:"-"`
}

func (sv *ServiceVersion) BeforeCreate(tx *gorm.DB) (err error) {
	sv.ID = uuid.New().String()
	sv.CreatedAt = time.Now()
	sv.UpdatedAt = time.Now()
	return
}

func (sv *ServiceVersion) BeforeUpdate(tx *gorm.DB) (err error) {
	sv.UpdatedAt = time.Now()
	return
}

type ServiceVersionModel struct{}

var serviceVersionValidSortFields = map[string]bool{
	"version":    true,
	"created_at": true,
	"updated_at": true,
}

func GetServiceVersionValidSortFields() map[string]bool {
	return serviceVersionValidSortFields
}

func (m ServiceVersionModel) Create(ctx context.Context, serviceID string, form forms.CreateServiceVersionForm) (serviceVersion ServiceVersion, err error) {
	db := db.GetDB()
	serviceVersion = ServiceVersion{
		Version:          form.Version,
		Description:      form.Description,
		ReleaseTimestamp: form.ReleaseTimestamp,
		ServiceID:        serviceID,
	}
	if err := db.Model(&ServiceVersion{}).Create(&serviceVersion).Error; err != nil {
		log.With(ctx).Errorf("failed to create service version for service with id %s :: error: %s", serviceID, err.Error())
		return ServiceVersion{}, err
	}
	return serviceVersion, err
}

// returns isFound as false when there is either an error running the query or if the record is not found
// caller must first check if err is not nil to know whether it is a record not found error
// or some other error and not directly rely on isFound for record not found case
func (m ServiceVersionModel) One(ctx context.Context, serviceID string, organizationID string, id string) (serviceVersion ServiceVersion, isFound bool, err error) {
	db := db.GetDB()

	// Join with services table to ensure the service belongs to the organization
	if err := db.Model(&ServiceVersion{}).
		Joins("JOIN services ON service_versions.service_id = services.id").
		Where("service_versions.service_id = ? AND service_versions.id = ? AND services.organization_id = ?", serviceID, id, organizationID).
		First(&serviceVersion).Error; err != nil {
		log.With(ctx).Errorf("failed to find service version with id %s for service with id %s :: error: %s", id, serviceID, err.Error())
		return ServiceVersion{}, !errors.Is(err, gorm.ErrRecordNotFound), err
	}
	return serviceVersion, true, nil
}

func (m ServiceVersionModel) All(ctx context.Context, serviceID string, organizationID string, q string, sortBy string, sort string, page int, limit int) (result PaginatedResult[ServiceVersion], err error) {
	db := db.GetDB()
	serviceVersions := make([]*ServiceVersion, 0) // Initialize as empty slice of pointers

	// Join with services table to ensure the service belongs to the organization
	tx := db.Model(&ServiceVersion{}).
		Joins("JOIN services ON service_versions.service_id = services.id").
		Where("service_versions.service_id = ? AND services.organization_id = ?", serviceID, organizationID)

	// Search filter
	if q != "" {
		tx = tx.Where("version ILIKE ?", fmt.Sprintf("%s%%", q))
	}

	// Get total count for pagination
	var totalCount int64
	if err := tx.Count(&totalCount).Error; err != nil {
		log.With(ctx).Errorf("failed to get count of service versions for service with id %s :: error: %s", serviceID, err.Error())
		return PaginatedResult[ServiceVersion]{}, err
	}

	// Apply sorting, validation and defaults are handled at API layer
	tx = tx.Order(fmt.Sprintf("%s %s", sortBy, sort))

	// Pagination
	offset := page * limit
	if err := tx.Limit(limit).Offset(offset).Find(&serviceVersions).Error; err != nil {
		log.With(ctx).Errorf("failed to get service versions for service with id %s :: error: %s", serviceID, err.Error())
		return PaginatedResult[ServiceVersion]{}, err
	}

	return BuildPaginatedResult(serviceVersions, totalCount, page, limit), nil
}

func (m ServiceVersionModel) Update(ctx context.Context, serviceID string, organizationID string, id string, form forms.UpdateServiceVersionForm) (serviceVersion ServiceVersion, err error) {
	db := db.GetDB()

	// First get the existing record with organization validation
	if err := db.Model(&ServiceVersion{}).
		Joins("JOIN services ON service_versions.service_id = services.id").
		Where("service_versions.service_id = ? AND service_versions.id = ? AND services.organization_id = ?", serviceID, id, organizationID).
		First(&serviceVersion).Error; err != nil {
		log.With(ctx).Errorf("failed to find service version with id %s for service with id %s :: error: %s", id, serviceID, err.Error())
		return ServiceVersion{}, err
	}

	// Update only the fields that are provided
	if form.Description != "" {
		serviceVersion.Description = form.Description
	}
	if form.ReleaseTimestamp != nil {
		serviceVersion.ReleaseTimestamp = *form.ReleaseTimestamp
	}

	if err := db.Save(&serviceVersion).Error; err != nil {
		log.With(ctx).Errorf("failed to update service version with id with id %s for service with id %s :: error: %s", id, serviceID, err.Error())
		return ServiceVersion{}, err
	}
	return serviceVersion, nil
}

func (m ServiceVersionModel) Delete(ctx context.Context, id string) (err error) {
	db := db.GetDB()

	if err := db.Where("id = ?", id).Delete(&ServiceVersion{}).Error; err != nil {
		log.With(ctx).Errorf("failed to delete service version with id %s :: error: %s", id, err.Error())
		return err
	}

	return nil
}
