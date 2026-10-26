package service

import (
	"backend/models"
	"testing"
)

func TestDroppedPublicIDsFindsRemovePhoto(t *testing.T) {
	before := []models.PostMedia{
		{PublicID: "A"},
		{PublicID: "B"},
	}
	after := []models.PostMedia{
		{PublicID: "A"},
	}
	dropped := droppedPublicIDs(before, after)

	if len(dropped) != 1 || dropped[0] != "B" {
		t.Errorf("expected [B] to be dropped, got %v", dropped)
	}
}