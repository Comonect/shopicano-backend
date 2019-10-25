package repositories

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/shopicano/shopicano-backend/app"
	"github.com/shopicano/shopicano-backend/models"
	"github.com/shopicano/shopicano-backend/utils"
	"time"
)

type UserRepositoryImpl struct {
}

var userRepository UserRepository

func NewUserRepository() UserRepository {
	if userRepository == nil {
		userRepository = &UserRepositoryImpl{}
	}
	return userRepository
}

func (uu *UserRepositoryImpl) Register(u *models.User) error {
	db := app.DB()
	if err := db.Model(u).Create(u).Error; err != nil {
		return err
	}
	return nil
}

func (uu *UserRepositoryImpl) Login(email, password string) (*models.Session, error) {
	u := models.User{}

	db := app.DB()
	if err := db.Model(&u).Where("email = ? AND status = ?", email, models.UserActive).First(&u).Error; err != nil {
		return nil, err
	}

	if err := utils.CheckPassword(u.Password, password); err != nil {
		return nil, gorm.ErrRecordNotFound
	}

	s := models.Session{
		ID:           utils.NewUUID(),
		UserID:       u.ID,
		AccessToken:  utils.NewToken(),
		RefreshToken: utils.NewToken(),
		CreatedAt:    time.Now().UTC(),
		ExpireOn:     time.Now().Add(time.Hour * 48).Unix(),
	}

	if err := db.Model(&s).Create(&s).Error; err != nil {
		return nil, err
	}

	return &s, nil
}

func (uu *UserRepositoryImpl) Logout(token string) error {
	s := models.Session{}

	db := app.DB()
	if err := db.Model(&s).Where("access_token = ?", token).Delete(&s).Error; err != nil {
		return err
	}
	return nil
}

func (uu *UserRepositoryImpl) RefreshToken(token string) (*models.Session, error) {
	os := models.Session{}

	db := app.DB().Begin()
	if err := db.Model(&os).Where("refresh_token = ?", token).First(&os).Error; err != nil {
		return nil, err
	}

	s := models.Session{
		ID:           utils.NewUUID(),
		UserID:       os.UserID,
		AccessToken:  utils.NewToken(),
		RefreshToken: utils.NewToken(),
		CreatedAt:    time.Now().UTC(),
		ExpireOn:     time.Now().Add(time.Hour * 48).Unix(),
	}

	if err := db.Model(&s).Create(&s).Error; err != nil {
		return nil, err
	}

	if err := db.Model(&os).Where("refresh_token = ?", token).Delete(&os).Error; err != nil {
		return nil, err
	}

	if err := db.Commit().Error; err != nil {
		return nil, err
	}

	return &s, nil
}

func (uu *UserRepositoryImpl) Update(userID string, ud *models.User) (*models.User, error) {
	u := models.User{}

	db := app.DB()
	if err := db.Model(&u).Where("id = ?", userID).First(&u).Error; err != nil {
		return nil, err
	}

	u.Name = ud.Name
	u.ProfilePicture = ud.ProfilePicture
	u.Phone = ud.Phone

	if err := db.Model(&u).Where("id = ?", userID).Save(&u).Error; err != nil {
		return nil, err
	}

	return &u, nil
}

func (uu *UserRepositoryImpl) GetPermission(token string) (string, *models.Permission, error) {
	db := app.DB()

	s := models.Session{}
	u := models.User{}
	up := models.UserPermission{}

	result := struct {
		ID         string             `json:"id"`
		Permission *models.Permission `json:"permission"`
	}{}

	if err := db.Table(fmt.Sprintf("%s AS s", s.TableName())).
		Select("u.id, up.permission").
		Joins(fmt.Sprintf("JOIN %s AS u ON u.id = s.user_id", u.TableName())).
		Joins(fmt.Sprintf("JOIN %s AS up ON u.permission_id = up.id", up.TableName())).
		Where("s.access_token = ? AND u.status = ?", token, models.UserActive).Scan(&result).Error; err != nil {
		return "", nil, err
	}
	return result.ID, result.Permission, nil
}

func (uu *UserRepositoryImpl) GetPermissionByUserID(userID string) (string, *models.Permission, error) {
	db := app.DB()

	u := models.User{}
	up := models.UserPermission{}

	result := struct {
		ID         string             `json:"id"`
		Permission *models.Permission `json:"permission"`
	}{}

	if err := db.Table(fmt.Sprintf("%s AS u", u.TableName())).
		Select("u.id, up.permission").
		Joins(fmt.Sprintf("JOIN %s AS up ON u.permission_id = up.id", up.TableName())).
		Where("u.id = ? AND u.status = ?", userID, models.UserActive).Scan(&result).Error; err != nil {
		return "", nil, err
	}
	return result.ID, result.Permission, nil
}

func (uu *UserRepositoryImpl) Get(userID string) (*models.User, error) {
	u := models.User{}

	db := app.DB()
	if err := db.Model(&u).Where("id = ?", userID).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (uu *UserRepositoryImpl) IsSignUpEnabled() (bool, error) {
	s := models.Settings{}
	db := app.DB()
	if err := db.Model(&s).Where("id = ?", "1").First(&s).Error; err != nil {
		return false, err
	}
	return s.IsSignUpEnabled, nil
}

func (uu *UserRepositoryImpl) IsStoreCreationEnabled() (bool, error) {
	s := models.Settings{}
	db := app.DB()
	if err := db.Model(&s).Where("id = ?", "1").First(&s).Error; err != nil {
		return false, err
	}
	return s.IsStoreCreationEnabled, nil
}
