package broker

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// SessionEnv names the environment variable carrying a session's sealed
// envelope into its container. It is set per exec, not per container: the
// broker outlives every individual session, so anything session-scoped in its
// own environment is stale for every session but the first.
const SessionEnv = "ASYLUM_SESSION"

// ErrNoEnvelope reports an absent envelope, which a handler distinguishes from
// a rejected one: the session simply has nothing sealed for it.
var ErrNoEnvelope = errors.New("no session envelope")

// ErrBadEnvelope reports an envelope that could not be opened — malformed,
// truncated, tampered with, or sealed under a different key.
var ErrBadEnvelope = errors.New("session envelope could not be opened")

const sessionKeyLen = 32

// Seal encrypts a session-scoped payload for delivery through a container.
// The container holds only the sealed form: it cannot read the payload, cannot
// produce one the broker will open, and cannot alter one without detection.
func Seal(payload any) (string, error) {
	plain, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	gcm, err := sessionGCM()
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(gcm.Seal(nonce, nonce, plain, nil)), nil
}

// Open decrypts an envelope into payload. It returns ErrNoEnvelope for an empty
// blob and ErrBadEnvelope for anything that fails to authenticate.
func Open(blob string, payload any) error {
	if blob == "" {
		return ErrNoEnvelope
	}
	// Strict rejects non-canonical trailing bits, so an envelope that differs
	// from the one issued is refused even when it decodes to the same bytes.
	raw, err := base64.RawURLEncoding.Strict().DecodeString(blob)
	if err != nil {
		return ErrBadEnvelope
	}
	gcm, err := sessionGCM()
	if err != nil {
		return err
	}
	if len(raw) < gcm.NonceSize() {
		return ErrBadEnvelope
	}
	plain, err := gcm.Open(nil, raw[:gcm.NonceSize()], raw[gcm.NonceSize():], nil)
	if err != nil {
		return ErrBadEnvelope
	}
	if err := json.Unmarshal(plain, payload); err != nil {
		return ErrBadEnvelope
	}
	return nil
}

func sessionGCM() (cipher.AEAD, error) {
	key, err := sessionKey()
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

// sessionKeyPath locates the envelope key. It sits at the ~/.asylum top level
// rather than in the per-container directory because that directory is
// bind-mounted into the container on a native Linux engine.
func sessionKeyPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".asylum", "session.key"), nil
}

// sessionKey reads the envelope key, creating it on first use. Concurrent
// creators race harmlessly: the loser's exclusive create fails and it reads the
// winner's key. The key is read per call — a 32-byte read is nothing next to
// the process spawn a handler is about to do, and caching would need
// invalidation it cannot get.
func sessionKey() ([]byte, error) {
	path, err := sessionKeyPath()
	if err != nil {
		return nil, err
	}
	key, err := os.ReadFile(path)
	if err == nil {
		if len(key) != sessionKeyLen {
			return nil, fmt.Errorf("session key %s is %d bytes, want %d", path, len(key), sessionKeyLen)
		}
		return key, nil
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}
	key = make([]byte, sessionKeyLen)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if errors.Is(err, fs.ErrExist) {
		return sessionKey()
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if _, err := f.Write(key); err != nil {
		return nil, err
	}
	return key, nil
}
