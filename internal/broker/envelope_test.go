package broker

import (
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

type testPayload struct {
	Session string `json:"session"`
	Cookie  string `json:"cookie,omitempty"`
}

func TestSealOpenRoundTrip(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	want := testPayload{Session: "w0t1p0:ABC-123", Cookie: "secret"}
	blob, err := Seal(want)
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}
	if blob == "" {
		t.Fatal("Seal returned an empty envelope")
	}

	var got testPayload
	if err := Open(blob, &got); err != nil {
		t.Fatalf("Open: %v", err)
	}
	if got != want {
		t.Errorf("round trip: got %+v, want %+v", got, want)
	}
}

func TestSealHidesPayload(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	blob, err := Seal(testPayload{Session: "w0t1p0:ABC-123"})
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}
	raw, err := base64.RawURLEncoding.DecodeString(blob)
	if err != nil {
		t.Fatalf("envelope is not base64url: %v", err)
	}
	if bytesContain(raw, "w0t1p0:ABC-123") {
		t.Error("sealed envelope contains the plaintext session id")
	}
}

func TestSealDiffersPerCall(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	p := testPayload{Session: "w0t1p0:ABC-123"}
	first, err := Seal(p)
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}
	second, err := Seal(p)
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}
	if first == second {
		t.Error("two seals of the same payload produced identical envelopes; nonce is not random")
	}
}

func TestOpenRejects(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	valid, err := Seal(testPayload{Session: "w0t1p0:ABC-123"})
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}
	raw, err := base64.RawURLEncoding.DecodeString(valid)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}

	tampered := append([]byte(nil), raw...)
	tampered[len(tampered)-1] ^= 0xff

	flipped := append([]byte(nil), raw...)
	flipped[0] ^= 0xff

	tests := []struct {
		name string
		blob string
		want error
	}{
		{"absent", "", ErrNoEnvelope},
		{"not base64", "!!!not base64!!!", ErrBadEnvelope},
		{"shorter than nonce", base64.RawURLEncoding.EncodeToString(raw[:4]), ErrBadEnvelope},
		{"truncated ciphertext", base64.RawURLEncoding.EncodeToString(raw[:len(raw)-1]), ErrBadEnvelope},
		{"tampered ciphertext", base64.RawURLEncoding.EncodeToString(tampered), ErrBadEnvelope},
		{"tampered nonce", base64.RawURLEncoding.EncodeToString(flipped), ErrBadEnvelope},
		{"non-canonical trailing bits", valid[:len(valid)-1] + "X", ErrBadEnvelope},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got testPayload
			err := Open(tt.blob, &got)
			if !errors.Is(err, tt.want) {
				t.Fatalf("Open: got %v, want %v", err, tt.want)
			}
			if got.Session != "" {
				t.Errorf("payload written despite error: %+v", got)
			}
		})
	}
}

func TestOpenRejectsForeignKey(t *testing.T) {
	sealer := t.TempDir()
	t.Setenv("HOME", sealer)
	blob, err := Seal(testPayload{Session: "w0t1p0:ABC-123"})
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	t.Setenv("HOME", t.TempDir()) // a different host account, so a different key
	var got testPayload
	if err := Open(blob, &got); !errors.Is(err, ErrBadEnvelope) {
		t.Fatalf("Open under a foreign key: got %v, want %v", err, ErrBadEnvelope)
	}
}

func TestSessionKeyCreatedOnceAndReused(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	first, err := sessionKey()
	if err != nil {
		t.Fatalf("sessionKey: %v", err)
	}
	if len(first) != sessionKeyLen {
		t.Fatalf("key is %d bytes, want %d", len(first), sessionKeyLen)
	}

	second, err := sessionKey()
	if err != nil {
		t.Fatalf("sessionKey again: %v", err)
	}
	if string(first) != string(second) {
		t.Error("key changed between calls")
	}

	path := filepath.Join(home, ".asylum", "session.key")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat key: %v", err)
	}
	if mode := info.Mode().Perm(); mode != 0600 {
		t.Errorf("key mode is %o, want 600", mode)
	}
}

func TestSessionKeyRejectsWrongLength(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir := filepath.Join(home, ".asylum")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "session.key"), []byte("short"), 0600); err != nil {
		t.Fatalf("write key: %v", err)
	}
	if _, err := sessionKey(); err == nil {
		t.Fatal("sessionKey accepted a truncated key file")
	}
}

func bytesContain(haystack []byte, needle string) bool {
	n := []byte(needle)
	for i := 0; i+len(n) <= len(haystack); i++ {
		if string(haystack[i:i+len(n)]) == string(n) {
			return true
		}
	}
	return false
}

// The key must not live under ~/.asylum/projects/<container>: that directory is
// bind-mounted into the container as /run/asylum for the broker socket on a
// native Linux engine, which would hand the sandbox the means to forge any
// envelope it likes.
func TestSessionKeyPathIsNotInAMountedDirectory(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	path, err := sessionKeyPath()
	if err != nil {
		t.Fatalf("sessionKeyPath: %v", err)
	}
	if want := filepath.Join(home, ".asylum"); filepath.Dir(path) != want {
		t.Errorf("session key at %s, want it directly in %s — a per-container subdirectory is bind-mounted into the container", path, want)
	}
}
