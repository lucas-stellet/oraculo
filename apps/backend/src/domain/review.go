package domain

import "time"

type ReviewVerdict string

const (
	ReviewApproved ReviewVerdict = "approved"
	ReviewRejected ReviewVerdict = "rejected"
)

var validReviewVerdicts = map[ReviewVerdict]bool{
	ReviewApproved: true, ReviewRejected: true,
}

func (v ReviewVerdict) Valid() bool { return validReviewVerdicts[v] }

type Review struct {
	ID          int           `json:"id"`
	VersionID   int           `json:"version_id"`
	VersionType VersionType   `json:"version_type"`
	Verdict     ReviewVerdict `json:"verdict"`
	Comment     string        `json:"comment"`
	CreatedAt   time.Time     `json:"created_at"`
}
