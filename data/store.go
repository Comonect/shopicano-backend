package data

import (
	"github.com/jinzhu/gorm"
	"github.com/shopicano/shopicano-backend/models"
)

type StoreRepository interface {
	GetStoreUserProfile(db *gorm.DB, userID string) (*models.StoreUserProfile, error)
	CreateStore(db *gorm.DB, s *models.Store) error
	FindStoreByID(db *gorm.DB, ID string) (*models.Store, error)
	AddStoreStuff(db *gorm.DB, staff *models.Staff) error
	ListStaffs(db *gorm.DB, storeID string, from, limit int) ([]models.StoreUserProfile, error)
	SearchStaffs(db *gorm.DB, storeID, query string, from, limit int) ([]models.StoreUserProfile, error)
	UpdateStoreStuffPermission(db *gorm.DB, staff *models.Staff) error
	DeleteStoreStuffPermission(db *gorm.DB, storeID, userID string) error
	IsAlreadyStaff(db *gorm.DB, userID string) (bool, error)
}
