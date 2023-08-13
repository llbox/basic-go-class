package dao

import (
	"basic-go-class/workspace/webook/internal/domain"
	"context"
	"errors"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"log"
	"time"
)

var (
	ErrUserDuplicateEmail = errors.New("邮箱冲突")
	ErrUserNotFound       = gorm.ErrRecordNotFound
)

type UserDAO struct {
	db *gorm.DB
}

type User struct {
	Id           int64  `gorm:"primaryKey,autoIncrement"`
	Email        string `gorm:"unique"`
	Password     string
	Nickname     string
	Birthday     string
	Introduction string
	Ctime        int64
	Utime        int64
}

func NewUserDAO(db *gorm.DB) *UserDAO {
	return &UserDAO{
		db: db,
	}
}

func (d *UserDAO) Insert(ctx context.Context, u User) error {
	now := time.Now().UnixMilli()
	u.Ctime = now
	u.Utime = now
	err := d.db.WithContext(ctx).Create(&u).Error
	if err, ok := err.(*mysql.MySQLError); ok {
		const uniqueConflictsErrorNo uint16 = 1062
		if err.Number == uniqueConflictsErrorNo {
			return ErrUserDuplicateEmail
		}
	}
	return err
}

func (d *UserDAO) FindByEmail(ctx context.Context, email string) (User, error) {
	var user User
	err := d.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	return user, err
}

func (d *UserDAO) UpdatesById(ctx context.Context, user domain.User) error {
	query := map[string]interface{}{
		"nickname":     user.Nickname,
		"birthday":     user.Birthday,
		"introduction": user.Introduction,
	}
	err := d.db.WithContext(ctx).Model(&User{}).Where("id = ?", user.Id).Updates(query).Error
	if err != nil {
		log.Printf("更新错误：%v", err)
	}

	return err
}

func (d *UserDAO) FindById(ctx context.Context, id int64) (User, error) {
	var user User
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	return user, err
}
