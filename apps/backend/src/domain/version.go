package domain

import "time"

type VersionType string

const (
	VersionTypeEpic  VersionType = "epic"
	VersionTypeStory VersionType = "story"
)

type EpicVersion struct {
	ID        int       `json:"id"`
	EpicID    int       `json:"epic_id"`
	Number    int       `json:"number"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type StoryVersion struct {
	ID        int       `json:"id"`
	StoryID   int       `json:"story_id"`
	Number    int       `json:"number"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}
