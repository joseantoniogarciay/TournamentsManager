// Command legal-audit-restore decrypts one legal-audit backup into an explicit destination.
package main

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	source, destination, keyPath := os.Getenv("LEGAL_AUDIT_BACKUP_SOURCE"), os.Getenv("LEGAL_AUDIT_BACKUP_RESTORE_DESTINATION"), os.Getenv("LEGAL_AUDIT_BACKUP_KEY_PATH")
	if source == "" || destination == "" || keyPath == "" {
		return fmt.Errorf("LEGAL_AUDIT_BACKUP_SOURCE, LEGAL_AUDIT_BACKUP_RESTORE_DESTINATION y LEGAL_AUDIT_BACKUP_KEY_PATH son obligatorias")
	}
	keyText, err := os.ReadFile(keyPath)
	if err != nil {
		return err
	}
	key, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(keyText)))
	if err != nil || len(key) != 32 {
		return fmt.Errorf("clave de backup inválida")
	}
	sealed, err := os.ReadFile(source)
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
	if len(sealed) < gcm.NonceSize() {
		return fmt.Errorf("backup cifrado inválido")
	}
	plain, err := gcm.Open(nil, sealed[:gcm.NonceSize()], sealed[gcm.NonceSize():], nil)
	if err != nil {
		return fmt.Errorf("no se pudo descifrar el backup: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0700); err != nil {
		return err
	}
	return os.WriteFile(destination, plain, 0600)
}
