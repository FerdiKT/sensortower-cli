package cmd

import "testing"

func TestNormalizeCategoryRankingsPage(t *testing.T) {
	limit, offset, err := normalizeCategoryRankingsPage(100, 0, true)
	if err != nil {
		t.Fatalf("normalizeCategoryRankingsPage() error = %v", err)
	}
	if limit != categoryRankingsPageCap || offset != 0 {
		t.Fatalf("normalizeCategoryRankingsPage() = (%d, %d), want (%d, 0)", limit, offset, categoryRankingsPageCap)
	}
}

func TestNormalizeCategoryRankingsPageRejectsLargeOffset(t *testing.T) {
	if _, _, err := normalizeCategoryRankingsPage(25, 25, false); err == nil {
		t.Fatal("expected offset validation error")
	}
}
