package images

import (
	"sort"
	"testing"
)

func TestNaturalLessOrdersNumericSuffixesNumerically(t *testing.T) {
	files := []string{"10.png", "2.png", "1.png", "img-10.png", "img-9.png", "B.png", "a.png"}
	sort.SliceStable(files, func(i, j int) bool { return naturalLess(files[i], files[j]) })

	want := []string{"1.png", "2.png", "10.png", "a.png", "B.png", "img-9.png", "img-10.png"}
	for i := range want {
		if files[i] != want[i] {
			t.Fatalf("position %d: want %q, got %q (full order %v)", i, want[i], files[i], files)
		}
	}
}

func TestValidateImageTypeRejectsUnknownTypes(t *testing.T) {
	if err := validateImageType("phoneScreenshots"); err != nil {
		t.Fatalf("phoneScreenshots should be valid: %v", err)
	}
	if err := validateImageType("promoGraphic"); err == nil {
		t.Fatal("promoGraphic is not an AppImageType and must be rejected")
	}
}
