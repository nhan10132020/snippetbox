package models

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID             int       `gorm:"column:id"`
	Name           string    `gorm:"column:name"`
	Email          string    `gorm:"column:email"`
	HashedPassword []byte    `gorm:"column:hashed_password"`
	Created        time.Time `gorm:"column:created"`
}

func (User) TableName() string { return "users" }

type UserModel struct {
	DB *gorm.DB
}

func (m *UserModel) Insert(name, email, password string) error {
	// bcrypt hash of plain-text password and returns 60-character long hash
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}
	user := User{
		Name:           name,
		Email:          email,
		HashedPassword: hashedPassword,
		Created:        time.Now(),
	}
	err = m.DB.Create(&user).Error
	if err != nil {
		var mySQLError *mysql.MySQLError
		if errors.As(err, &mySQLError) {
			if mySQLError.Number == 1062 && strings.Contains(mySQLError.Message, "users_uc_email") {
				return ErrDuplicateEmail
			}
		}
		return err
	}
	return nil
}

func (m *UserModel) Authenticate(email, password string) (int, error) {
	var user User
	err := m.DB.Where("email = ?", email).First(&user).Error
	if err != nil {
		// check is email exist
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	// check matching password
	err = bcrypt.CompareHashAndPassword(user.HashedPassword, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	return user.ID, nil
}

// check if user exists with a specific ID
func (m *UserModel) Exists(id int) (bool, error) {
	var exists bool
	err := m.DB.
		Model(&User{}).
		Select("count(*) > 0").
		Where("id = ?", id).
		Find(&exists).
		Error
	return exists, err
}

type UserGet struct {
	Name    string    `gorm:"column:name"`
	Email   string    `gorm:"column:email"`
	Created time.Time `gorm:"column:created"`
}

func (UserGet) TableName() string { return "users" }

func (m *UserModel) Get(id int) (*UserGet, error) {
	var user *UserGet
	err := m.DB.Where("id = ?", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNoRecord
		}
		return nil, err
	}
	return user, nil
}

func (m *UserModel) PasswordUpdate(id int, currentPassword, newPassword string) error {
	var user *User

	err := m.DB.
		Where("id = ?", id).
		First(&user).
		Error

	if err != nil {
		return err
	}

	currentHashedPassword := user.HashedPassword

	err = bcrypt.CompareHashAndPassword(currentHashedPassword, []byte(currentPassword))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrInvalidCredentials
		} else {
			return err
		}
	}

	newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return err
	}

	err = m.DB.Model(&User{}).Where("id = ?", id).Update("hashed_password", newHashedPassword).Error
	return err
}
