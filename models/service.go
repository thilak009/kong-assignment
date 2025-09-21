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
	Name        string `json:"name"`
	Description string `json:"description"`
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

func (m ServiceModel) Create(form forms.CreateServiceForm) (service Service, err error) {
	db := db.GetDB()
	service = Service{
		Name:        form.Name,
		Description: form.Description,
	}
	if err := db.Model(&Service{}).Create(&service).Error; err != nil {
		return Service{}, err
	}
	return service, err
}

// returns isFound as false when there is either an error running the query or if the record is not found
// caller must first check if err is not nil to know whether it is a record not found error
// or some other error and not directly rely on isFound for record not found case
func (m ServiceModel) One(id string) (service Service, isFound bool, err error) {
	db := db.GetDB()
	if err := db.Model(&Service{}).Where("id = ?", id).First(&service).Error; err != nil {
		return Service{}, !errors.Is(err, gorm.ErrRecordNotFound), err
	}
	return service, true, err
}

func (m ServiceModel) All(q string, sortBy string, sort string, page int, limit int) (result PaginatedResult[Service], err error) {
	db := db.GetDB()
	services := make([]*Service, 0) // Initialize as empty slice of pointers
	tx := db.Model(&Service{})

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

	return BuildPaginatedResult(services, totalCount, page, limit), nil
}

func (m ServiceModel) Update(id string, form forms.CreateServiceForm) (service Service, err error) {
	db := db.GetDB()
	service = Service{
		Name:        form.Name,
		Description: form.Description,
	}
	service.ID = id
	if err := db.Model(&Service{}).Where("id = ?", id).Save(&service).Error; err != nil {
		return Service{}, err
	}
	return service, err
}

func (m ServiceModel) Delete(id string) (err error) {
	db := db.GetDB()
	if err := db.Where("id = ?", id).Delete(&Service{}).Error; err != nil {
		return err
	}
	return err
}
