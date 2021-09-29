package main

import(
	"testing"
)

func TestAddingUnexistingUser(t *testing.T) {
	userStor := NewInMemoryUserStorage()
	key := "nick"
	user := User{
		Email:		"myemail@gmail.com",
		PasswordDigest:	"QwErTy123",
		FavoriteCake:	"Orange",
	}
	err := userStor.Add(key, user)
	_, ok := userStor.storage[key]
	if err != nil {
		t.Errorf("Add(key, user) = %s; want nil", err)
	}
	if !ok {
		t.Error("The user was not added")
	}
}

func TestAddingExistingUser(t *testing.T) {
	userStor := NewInMemoryUserStorage()
	key := "nick"
	user := User{
		Email:		"myemail@gmail.com",
		PasswordDigest:	"QwErTy123",
		FavoriteCake:	"Orange",
	}
	userStor.Add(key, user)
	err := userStor.Add(key, user)
	if err == nil {
		t.Errorf("Add(key, user) = nil; want error")
	}
}

func TestUpdateExistingUser(t *testing.T) {
	userStor := NewInMemoryUserStorage()
	key := "nick"
	user := User{
		Email:		"myemail@gmail.com",
		PasswordDigest:	"QwErTy123",
		FavoriteCake:	"Orange",
	}
	userStor.Add(key, user)
	user.Email = "email@gmail.com"
	err := userStor.Update(key, user)
	newUser, _ := userStor.storage[key]
	if err != nil {
		t.Errorf("Add(key, user) = %s; want nil", err)
	}
	if newUser.Email != "email@gmail.com" {
		t.Error("The user was not updated")
	}
}

func TestUpdatingUnexistingUser(t *testing.T) {
	userStor := NewInMemoryUserStorage()
	key := "nick"
	user := User{
		Email:		"myemail@gmail.com",
		PasswordDigest:	"QwErTy123",
		FavoriteCake:	"Orange",
	}
	err := userStor.Update(key, user)
	if err == nil {
		t.Errorf("Update(key, user) = nil; want error")
	}
}

func TestGettingExistingUser(t *testing.T) {
	userStor := NewInMemoryUserStorage()
	key := "nick"
	user := User{
		Email:		"myemail@gmail.com",
		PasswordDigest:	"QwErTy123",
		FavoriteCake:	"Orange",
	}
	userStor.Add(key, user)
	gotUser, err := userStor.Get(key)
	if err != nil {
		t.Errorf("Add(key, user) = %s; want nil", err)
	}
	if gotUser != user {
		t.Error("Got incorrect user")
	}
}

func TestGettingUnexistingUser(t *testing.T) {
	userStor := NewInMemoryUserStorage()
	key := "nick"
	_, err := userStor.Get(key)
	if err == nil {
		t.Errorf("Get(key) = nil; want error")
	}
}

func TestDeletingUnexistingUser(t *testing.T) {
	userStor := NewInMemoryUserStorage()
	key := "nick"
	_, err := userStor.Delete(key)
	if err == nil {
		t.Errorf("Delete(key) = nil; want error")
	}
}

func TestDeletingExistingUser(t *testing.T) {
	userStor := NewInMemoryUserStorage()
	key := "nick"
	user := User{
		Email:		"myemail@gmail.com",
		PasswordDigest:	"QwErTy123",
		FavoriteCake:	"Orange",
	}
	userStor.Add(key, user)
	gotUser, err := userStor.Delete(key)
	if err != nil {
		t.Errorf("Add(key, user) = %s; want nil", err)
	}
	if user != gotUser {
		t.Error("Not the right user was deleted")
	}
}
