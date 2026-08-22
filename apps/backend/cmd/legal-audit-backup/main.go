// Command legal-audit-backup writes an encrypted incremental legal-audit export.
package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type record struct {
	ID, EmailHash, TermsVersion, TermsContentHash, Source string
	AcceptedAt                                            time.Time
	UpdatedAt                                             time.Time
	RetentionUntil                                        *time.Time
}

type recordState struct {
	RetentionUntil *time.Time `json:"retentionUntil"`
}

type fileState struct {
	RecordIDs []string `json:"recordIds"`
}

type checkpoint struct {
	// AcceptedAt is retained only to read state files created before updated_at
	// existed. New state advances by (updated_at, id).
	AcceptedAt time.Time              `json:"acceptedAt,omitempty"`
	UpdatedAt  time.Time              `json:"updatedAt"`
	LastID     string                 `json:"lastId"`
	Records    map[string]recordState `json:"records"`
	Files      map[string]fileState   `json:"files"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	dbURL, destination, keyPath, statePath := os.Getenv("DATABASE_URL"), os.Getenv("LEGAL_AUDIT_BACKUP_DESTINATION"), os.Getenv("LEGAL_AUDIT_BACKUP_KEY_PATH"), os.Getenv("LEGAL_AUDIT_BACKUP_STATE_PATH")
	if dbURL == "" || destination == "" || keyPath == "" || statePath == "" {
		return fmt.Errorf("DATABASE_URL y las variables LEGAL_AUDIT_BACKUP_* son obligatorias")
	}
	keyText, err := os.ReadFile(keyPath)
	if err != nil {
		return err
	}
	key, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(keyText)))
	if err != nil || len(key) != 32 {
		return fmt.Errorf("clave de backup inválida")
	}
	var since checkpoint
	if raw, err := os.ReadFile(statePath); err == nil {
		if err := json.Unmarshal(raw, &since); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		return err
	}
	defer pool.Close()
	if since.UpdatedAt.IsZero() {
		since.UpdatedAt = since.AcceptedAt
	}
	if since.Records == nil {
		since.Records = make(map[string]recordState)
	}
	if since.Files == nil {
		since.Files = make(map[string]fileState)
	}
	rows, err := pool.Query(context.Background(), `SELECT id::text, encode(email_hash,'hex'), terms_version, encode(terms_content_hash,'hex'), source, accepted_at, updated_at, retention_until FROM legal_account_acceptances WHERE updated_at > $1 OR (updated_at = $1 AND id::text > $2) ORDER BY updated_at, id`, since.UpdatedAt, since.LastID)
	if err != nil {
		return err
	}
	defer rows.Close()
	var records []record
	latest, lastID := since.UpdatedAt, since.LastID
	for rows.Next() {
		var r record
		if err := rows.Scan(&r.ID, &r.EmailHash, &r.TermsVersion, &r.TermsContentHash, &r.Source, &r.AcceptedAt, &r.UpdatedAt, &r.RetentionUntil); err != nil {
			return err
		}
		records = append(records, r)
		latest, lastID = r.UpdatedAt, r.ID
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if len(records) > 0 {
		plain, err := json.Marshal(records)
		if err != nil {
			return err
		}
		block, err := aes.NewCipher(key)
		if err != nil {
			return err
		}
		gcm, err := cipher.NewGCM(block)
		if err != nil {
			return err
		}
		nonce := make([]byte, gcm.NonceSize())
		if _, err = rand.Read(nonce); err != nil {
			return err
		}
		sealed := gcm.Seal(nonce, nonce, plain, nil)
		if err := os.MkdirAll(destination, 0700); err != nil {
			return err
		}
		name := fmt.Sprintf("legal-audit-%s.json.aes", time.Now().UTC().Format("20060102T150405Z"))
		if err := os.WriteFile(filepath.Join(destination, name), sealed, 0600); err != nil {
			return err
		}
		ids := make([]string, 0, len(records))
		for _, r := range records {
			ids = append(ids, r.ID)
			since.Records[r.ID] = recordState{RetentionUntil: r.RetentionUntil}
		}
		since.Files[name] = fileState{RecordIDs: ids}
		since.UpdatedAt, since.LastID = latest, lastID
	}
	if err := writeCheckpoint(statePath, since); err != nil {
		return err
	}
	if err := pruneExpiredFiles(destination, &since, time.Now().UTC()); err != nil {
		return err
	}
	return writeCheckpoint(statePath, since)
}

func writeCheckpoint(path string, state checkpoint) error {
	encoded, err := json.Marshal(state)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	return os.WriteFile(path, encoded, 0600)
}

func pruneExpiredFiles(destination string, state *checkpoint, now time.Time) error {
	for name, file := range state.Files {
		expired := len(file.RecordIDs) > 0
		for _, id := range file.RecordIDs {
			record, exists := state.Records[id]
			if !exists || record.RetentionUntil == nil || record.RetentionUntil.After(now) {
				expired = false
				break
			}
		}
		if !expired {
			continue
		}
		if err := os.Remove(filepath.Join(destination, name)); err != nil && !os.IsNotExist(err) {
			return err
		}
		delete(state.Files, name)
	}
	return nil
}
