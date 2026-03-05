package domain

import "time"

type VersionType string

const (
	VersionTypeEpic  VersionType = "epic"
	VersionTypeStory VersionType = "story"
)

type EpicVersion struct {
	ID        int
	EpicID    int
	Number    int
	Content   string
	CreatedAt time.Time
}

type StoryVersion struct {
	ID        int
	StoryID   int
	Number    int
	Content   string
	CreatedAt time.Time
}
