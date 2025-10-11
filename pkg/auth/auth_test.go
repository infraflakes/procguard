package auth

import "testing"

func TestPasswordHashing(t *testing.T) {
	password := "mysecretpassword"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	if !CheckPasswordHash(password, hash) {
		t.Errorf("CheckPasswordHash() = false, want true")
	}

	if CheckPasswordHash("wrongpassword", hash) {
		t.Errorf("CheckPasswordHash() with wrong password = true, want false")
	}
}
