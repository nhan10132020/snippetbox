package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type Snippet struct {
	ID      int       `gorm:"column:id"`
	Title   string    `gorm:"column:title"`
	Content string    `gorm:"column:content"`
	Created time.Time `gorm:"column:created"`
	Expires time.Time `gorm:"column:expires"`
}

func (Snippet) TableName() string { return "snippets" }

type SnippetModel struct {
	DB *gorm.DB
}

// This will insert a new snippet into the database.
func (m *SnippetModel) Insert(title string, content string, expires int) (int, error) {
	snippet := Snippet{
		Title:   title,
		Content: content,
		Created: time.Now(),
		Expires: time.Now().AddDate(0, 0, expires),
	}
	if err := m.DB.Create(&snippet).Error; err != nil {
		return 0, err
	}
	return snippet.ID, nil
}

// This will return a specific snippet based on its id.
func (m *SnippetModel) Get(id int) (*Snippet, error) {
	var snippet *Snippet
	err := m.DB.Where("id = ? AND expires > UTC_TIMESTAMP()", id).First(&snippet).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNoRecord
		}
		return nil, err
	}
	return snippet, nil
}

// This will return the 10 most recently created and haven't expired snippets.
func (m *SnippetModel) Latest() ([]*Snippet, error) {
	var snippet []*Snippet
	err := m.DB.Order("id desc").Where("expires > UTC_TIMESTAMP()").Limit(10).Find(&snippet).Error
	if err != nil {
		return nil, err
	}
	return snippet, nil
}
