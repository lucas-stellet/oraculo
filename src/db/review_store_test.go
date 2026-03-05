package db

import (
	"errors"
	"testing"

	"github.com/lucas/oraculo/src/domain"
)

func TestReviewStore_Create_Approved(t *testing.T) {
	database := testDB(t)
	epicID := createTestEpic(t, database, "my-epic")
	versionStore := NewVersionStore(database)
	reviewStore := NewReviewStore(database)

	v, _ := versionStore.CreateEpicVersion(epicID, "epic content")

	review, err := reviewStore.Create(v.ID, domain.VersionTypeEpic, domain.ReviewApproved, "looks good")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if review.VersionID != v.ID {
		t.Errorf("VersionID = %d, want %d", review.VersionID, v.ID)
	}
	if review.VersionType != domain.VersionTypeEpic {
		t.Errorf("VersionType = %q, want %q", review.VersionType, domain.VersionTypeEpic)
	}
	if review.Verdict != domain.ReviewApproved {
		t.Errorf("Verdict = %q, want %q", review.Verdict, domain.ReviewApproved)
	}
	if review.Comment != "looks good" {
		t.Errorf("Comment = %q, want %q", review.Comment, "looks good")
	}
	if review.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}

	// Check propagation to epic.
	epic, _ := NewEpicStore(database).GetByName("my-epic")
	if epic.ApprovalStatus != domain.ApprovalApproved {
		t.Errorf("epic ApprovalStatus = %q, want %q", epic.ApprovalStatus, domain.ApprovalApproved)
	}
}

func TestReviewStore_Create_Rejected(t *testing.T) {
	database := testDB(t)
	epicID := createTestEpic(t, database, "my-epic")
	versionStore := NewVersionStore(database)
	reviewStore := NewReviewStore(database)

	v, _ := versionStore.CreateEpicVersion(epicID, "epic content")

	review, err := reviewStore.Create(v.ID, domain.VersionTypeEpic, domain.ReviewRejected, "needs work")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if review.Verdict != domain.ReviewRejected {
		t.Errorf("Verdict = %q, want %q", review.Verdict, domain.ReviewRejected)
	}

	// Check propagation to epic.
	epic, _ := NewEpicStore(database).GetByName("my-epic")
	if epic.ApprovalStatus != domain.ApprovalRejected {
		t.Errorf("epic ApprovalStatus = %q, want %q", epic.ApprovalStatus, domain.ApprovalRejected)
	}
}

func TestReviewStore_Create_Story(t *testing.T) {
	database := testDB(t)
	epicID := createTestEpic(t, database, "my-epic")
	storyID := createTestStory(t, database, epicID, "my-story")
	versionStore := NewVersionStore(database)
	reviewStore := NewReviewStore(database)

	v, _ := versionStore.CreateStoryVersion(storyID, "story content")

	review, err := reviewStore.Create(v.ID, domain.VersionTypeStory, domain.ReviewApproved, "lgtm")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if review.VersionType != domain.VersionTypeStory {
		t.Errorf("VersionType = %q, want %q", review.VersionType, domain.VersionTypeStory)
	}

	// Check propagation to story.
	story, _ := NewStoryStore(database).GetByName(epicID, "my-story")
	if story.ApprovalStatus != domain.ApprovalApproved {
		t.Errorf("story ApprovalStatus = %q, want %q", story.ApprovalStatus, domain.ApprovalApproved)
	}
}

func TestReviewStore_GetByID(t *testing.T) {
	database := testDB(t)
	epicID := createTestEpic(t, database, "my-epic")
	versionStore := NewVersionStore(database)
	reviewStore := NewReviewStore(database)

	v, _ := versionStore.CreateEpicVersion(epicID, "content")
	created, _ := reviewStore.Create(v.ID, domain.VersionTypeEpic, domain.ReviewApproved, "ok")

	got, err := reviewStore.GetByID(created.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("ID = %d, want %d", got.ID, created.ID)
	}
}

func TestReviewStore_GetByID_NotFound(t *testing.T) {
	database := testDB(t)
	reviewStore := NewReviewStore(database)

	_, err := reviewStore.GetByID(999)
	if err == nil {
		t.Fatal("expected error for nonexistent review")
	}
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestReviewStore_ListByVersion(t *testing.T) {
	database := testDB(t)
	epicID := createTestEpic(t, database, "my-epic")
	versionStore := NewVersionStore(database)
	reviewStore := NewReviewStore(database)

	v, _ := versionStore.CreateEpicVersion(epicID, "content")
	reviewStore.Create(v.ID, domain.VersionTypeEpic, domain.ReviewRejected, "first review")
	reviewStore.Create(v.ID, domain.VersionTypeEpic, domain.ReviewApproved, "second review")

	reviews, err := reviewStore.ListByVersion(v.ID, domain.VersionTypeEpic)
	if err != nil {
		t.Fatalf("ListByVersion: %v", err)
	}
	if len(reviews) != 2 {
		t.Fatalf("len = %d, want 2", len(reviews))
	}
	if reviews[0].Comment != "first review" {
		t.Errorf("reviews[0].Comment = %q, want %q", reviews[0].Comment, "first review")
	}
	if reviews[1].Comment != "second review" {
		t.Errorf("reviews[1].Comment = %q, want %q", reviews[1].Comment, "second review")
	}
}
