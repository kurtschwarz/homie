package secrets

import (
	"bytes"
	"testing"
)

func TestBoxRoundTripAndAuthentication(t *testing.T) {
	box, err := New(bytes.Repeat([]byte{7}, 32))
	if err != nil {
		t.Fatal(err)
	}
	sealed, err := box.Seal([]byte("calendar password"))
	if err != nil {
		t.Fatal(err)
	}
	opened, err := box.Open(sealed)
	if err != nil {
		t.Fatal(err)
	}
	if string(opened) != "calendar password" {
		t.Fatalf("got %q", opened)
	}
	sealed[len(sealed)-1] ^= 1
	if _, err := box.Open(sealed); err == nil {
		t.Fatal("tampered value decrypted")
	}
}

func TestBoxRejectsWrongKeyAndTruncatedValue(t *testing.T) {
	one, _ := New(bytes.Repeat([]byte{1}, 32))
	two, _ := New(bytes.Repeat([]byte{2}, 32))
	sealed, _ := one.Seal([]byte("secret"))
	if _, err := two.Open(sealed); err == nil {
		t.Fatal("wrong key decrypted")
	}
	if _, err := one.Open([]byte{1}); err == nil {
		t.Fatal("truncated value decrypted")
	}
}
