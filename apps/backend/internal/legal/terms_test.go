package legal

import (
	"crypto/sha256"
	"os"
	"testing"
)

func TestCurrentTermsContentHashMatchesVersionedDocument(t *testing.T) {
	t.Parallel()
	content, err := os.ReadFile("../../../../docs/legal/TERMS_OF_USE_2026-08-22.md")
	if err != nil {
		t.Fatalf("read terms document: %v", err)
	}
	want := sha256.Sum256(content)
	if string(CurrentTermsContentHash()) != string(want[:]) {
		t.Fatal("current terms hash does not match its versioned document")
	}
}
