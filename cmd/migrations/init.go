package migration

import (
	"github.com/shopicano/shopicano-backend/app"
	"github.com/shopicano/shopicano-backend/log"
	"github.com/shopicano/shopicano-backend/models"
	"github.com/shopicano/shopicano-backend/utils"
	"github.com/shopicano/shopicano-backend/values"
	"github.com/spf13/cobra"
	"time"
)

var MigInitCmd = &cobra.Command{
	Use:   "init",
	Short: "init creates initially required data",
	Run:   initCmd,
}

func initCmd(cmd *cobra.Command, args []string) {
	tx := app.DB().Begin()

	s := models.Settings{
		ID:                     "1",
		Name:                   "Fin Shop",
		URL:                    "http://finshop.com",
		IsActive:               true,
		CompanyName:            "Fin Shop Ltd.",
		CompanyAddress:         "Dhaka",
		CompanyCity:            "Dhaka",
		CompanyCountry:         "Bangladesh",
		CompanyPostcode:        "1207",
		CompanyEmail:           "admin@example.com",
		CompanyPhone:           "0000000000",
		IsSignUpEnabled:        false,
		IsStoreCreationEnabled: false,
		TagLine:                "Do it",
		CreatedAt:              time.Now().UTC(),
		UpdatedAt:              time.Now().UTC(),
	}
	if err := tx.Table(s.TableName()).Create(&s).Error; err != nil {
		tx.Rollback()
		return
	}

	upAdmin := models.UserPermission{
		ID:         values.AdminGroupID,
		Permission: models.AdminPerm,
	}
	if err := tx.Table(upAdmin.TableName()).Create(&upAdmin).Error; err != nil {
		tx.Rollback()
		return
	}
	upUser := models.UserPermission{
		ID:         values.UserGroupID,
		Permission: models.UserPerm,
	}
	if err := tx.Table(upUser.TableName()).Create(&upUser).Error; err != nil {
		tx.Rollback()
		return
	}

	password, _ := utils.GeneratePassword("admin")

	u := models.User{
		ID:             utils.NewUUID(),
		Name:           "Fin Admin",
		Status:         models.UserActive,
		Phone:          nil,
		ProfilePicture: nil,
		Password:       password,
		PermissionID:   upAdmin.ID,
		Email:          "admin@example.com",
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}
	if err := tx.Table(u.TableName()).Create(&u).Error; err != nil {
		tx.Rollback()
		return
	}

	if err := tx.Commit().Error; err != nil {
		log.Log().Infoln("Migration init failed with err : ", err)
		return
	}
	log.Log().Infoln("Migration init completed")
}
