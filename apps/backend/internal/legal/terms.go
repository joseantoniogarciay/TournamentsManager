// Package legal contains the versioned legal documents required by account creation.
package legal

import "encoding/hex"

const CurrentTermsVersion = "2026-08-22"

// CurrentTermsContentSHA256 is the SHA-256 of docs/legal/TERMS_OF_USE_2026-08-22.md.
const CurrentTermsContentSHA256 = "f5557edc3100acba737728d557492234e794b6e55ba413b3f5ac0e7cf662cf76"

func CurrentTermsContentHash() []byte {
	hash, err := hex.DecodeString(CurrentTermsContentSHA256)
	if err != nil {
		panic("invalid current terms content hash")
	}
	return hash
}
