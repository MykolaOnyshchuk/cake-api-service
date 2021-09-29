package main

import(
	"testing"
	// "github.com/openware/rango/pkg/auth"
)

func TestNewJWTService(t *testing.T) {
	_, err := NewJWTService("pubkey.rsa", "privkey.rsa")
	if err != nil {
		t.Errorf("Expected: err = nil; actual err = %s", err)
	}
}

func TestNewJWTServiceInvalid(t *testing.T) {
	jwtService, err := NewJWTService("", "")
	if jwtService != nil {
		t.Errorf("jwtService expected: nil; actual: &JWTService")
	}
	if err == nil {
		t.Errorf("Expected: err = error; actual err = nil")
	}
}

func TestGenerateJWT(t *testing.T) {
	user := User{
		Email:		"myemail@gmail.com",
		PasswordDigest:	"QwErTy123",
		FavoriteCake:	"Orange",
	}
	jwtService, _ := NewJWTService("pubkey.rsa", "privkey.rsa")
	// keys, _ := auth.LoadOrGenerateKeys("pubkey.rsa", "privkey.rsa")
	_, err := jwtService.GenearateJWT(user)
	if err != nil {
		t.Errorf("Expected: err = nil; actual: %s", err)
	}
}

func TestParseJWT(t *testing.T) {
	user := User{
		Email:		"myemail@gmail.com",
		PasswordDigest:	"QwErTy123",
		FavoriteCake:	"Orange",
	}
	jwtService, _ := NewJWTService("pubkey.rsa", "privkey.rsa")
	// keys, _ := auth.LoadOrGenerateKeys("pubkey.rsa", "privkey.rsa")
	token, _ := jwtService.GenearateJWT(user)
	_, err := jwtService.ParseJWT(token)
	if err != nil {
		t.Errorf("Expected: err = nil; actual: %s", err)
	}
}
