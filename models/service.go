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

func (m ServiceModel) All(q string, sortBy string, sort string) (services []Service, err error) {
	db := db.GetDB()
	services = make([]Service, 0) // Initialize as empty slice instead of nil
	tx := db.Model(&Service{})

	// Search filter
	if q != "" {
		tx = tx.Where("name ILIKE ?", fmt.Sprintf("%%%s%%", q))
	}

	// Sorting with validation
	validSortFields := map[string]bool{
		"name":       true,
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

	if err := tx.Find(&services).Error; err != nil {
		return []Service{}, err
	}
	return services, err
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
