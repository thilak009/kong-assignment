package models

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/thilak009/kong-assignment/db"
	"github.com/thilak009/kong-assignment/forms"
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

func (m ServiceVersionModel) Create(serviceID string, form forms.CreateServiceVersionForm) (serviceVersion ServiceVersion, err error) {
	db := db.GetDB()
	serviceVersion = ServiceVersion{
		Version:          form.Version,
		Description:      form.Description,
		ReleaseTimestamp: form.ReleaseTimestamp,
		ServiceID:        serviceID,
	}
	if err := db.Model(&ServiceVersion{}).Create(&serviceVersion).Error; err != nil {
		return ServiceVersion{}, err
	}
	return serviceVersion, err
}

// returns isFound as false when there is either an error running the query or if the record is not found
// caller must first check if err is not nil to know whether it is a record not found error
// or some other error and not directly rely on isFound for record not found case
func (m ServiceVersionModel) One(serviceID string, id string) (serviceVersion ServiceVersion, isFound bool, err error) {
	db := db.GetDB()
	if err := db.Model(&ServiceVersion{}).Where("service_id = ? AND id = ?", serviceID, id).First(&serviceVersion).Error; err != nil {
		return ServiceVersion{}, !errors.Is(err, gorm.ErrRecordNotFound), err
	}
	return serviceVersion, true, err
}

func (m ServiceVersionModel) All(serviceID string, q string, sortBy string, sort string) (serviceVersions []ServiceVersion, err error) {
	db := db.GetDB()
	serviceVersions = make([]ServiceVersion, 0) // Initialize as empty slice instead of nil
	tx := db.Model(&ServiceVersion{}).Where("service_id = ?", serviceID)

	// Search filter
	if q != "" {
		tx = tx.Where("version ILIKE ?", fmt.Sprintf("%s%%", q))
	}

	// Sorting with validation
	validSortFields := map[string]bool{
		"version":    true,
		"created_at": true,
		"updated_at": true,
	}
	validSortOrder := map[string]bool{
		"asc":  true,
		"desc": true,
	}

	if validSortFields[sortBy] && validSortOrder[sort] {
		tx = tx.Order(fmt.Sprintf("%s %s", sortBy, strings.ToUpper(sort)))
	} else {
		tx = tx.Order("updated_at DESC") // default
	}

	if err := tx.Find(&serviceVersions).Error; err != nil {
		return []ServiceVersion{}, err
	}
	return serviceVersions, err
}

func (m ServiceVersionModel) Update(serviceID string, id string, form forms.UpdateServiceVersionForm) (serviceVersion ServiceVersion, err error) {
	db := db.GetDB()

	// First get the existing record
	if err := db.Model(&ServiceVersion{}).Where("service_id = ? AND id = ?", serviceID, id).First(&serviceVersion).Error; err != nil {
		return ServiceVersion{}, err
	}

	// Update only the fields that are provided
	if form.Description != "" {
		serviceVersion.Description = form.Description
	}
	if form.ReleaseTimestamp != nil {
		serviceVersion.ReleaseTimestamp = *form.ReleaseTimestamp
	}

	if err := db.Model(&ServiceVersion{}).Where("id = ?", id).Save(&serviceVersion).Error; err != nil {
		return ServiceVersion{}, err
	}
	return serviceVersion, err
}

func (m ServiceVersionModel) Delete(serviceID string, id string) (err error) {
	db := db.GetDB()
	if err := db.Where("service_id = ? AND id = ?", serviceID, id).Delete(&ServiceVersion{}).Error; err != nil {
		return err
	}
	return err
}
