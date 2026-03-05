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
	ID          int
	VersionID   int
	VersionType VersionType
	Verdict     ReviewVerdict
	Comment     string
	CreatedAt   time.Time
}
