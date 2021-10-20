package main

import (
	"net/mail"
	"errors"
	"net/http"
	"crypto/md5"
	"encoding/json"

)

type User struct {
	Email string
	PasswordDigest string
	FavoriteCake string
}

type UserRepository interface {
	Add(string, User) error
	Get(string) (User, error)
	Update(string, User) error
	Delete(string) (User, error)
}


type UserService struct {
	repository UserRepository
}

type UserRegisterParams struct {// If it looks strange, read about golang struct tags
	Email string `json:"email"`
	Password string `json:"password"`
	FavoriteCake string `json:"favorite_cake"`
}

type CakeUpdate struct {
	FavoriteCake string `json:"favorite_cake"`
}

type EmailUpdate struct {
	Email string `json:"email"`
}

type PasswordUpdate struct {
	Password string `json:"password"`
}

func validateRegisterParams(p *UserRegisterParams) error {
	// 1. Email is valid

	//if _, err := mail.ParseAddress(p.Email); err != nil {
	//	return err
	//}

	if validateEmail(p.Email) != nil {
		return validateEmail(p.Email)
	}

	// 2. Password at least 8 symbols

	if validatePassword(p.Password) != nil {
		return validatePassword(p.Password)
	}

	// 3. Favorite cake not empty
	// 4. Favorite cake only alphabetic

	if validateCake(p.FavoriteCake) != nil {
		return validateCake(p.FavoriteCake)
	}
	return nil
}

func validatePassword(password string) error {
	if len([]rune(password)) < 8 {
		return errors.New("The password must be at least 8 symbols")
	}
	return nil
}

func validateEmail(email string) error {
	_, err := mail.ParseAddress(email)
	return err
}

func validateCake(cake string) error {
	if cake == "" {
		return errors.New("The favorite cake field is empty")
	}
	for i := 0; i < len([]rune(cake)); i++ {
		if c := []rune(cake)[i]; c < 65 || (c > 90 && c < 97) || c > 122 {
			return errors.New("Favorite cake field should not contain only alphabetic values")
		}
	}
	return nil
}

func (u *UserService) Register(w http.ResponseWriter, r *http.Request) {
	params := &UserRegisterParams{}
	err := json.NewDecoder(r.Body).Decode(params)
	if err != nil {
		handleError(errors.New("could not read params"), w)
		return
	}
	if err := validateRegisterParams(params); err != nil {
		handleError(err, w)
		return
	}
	passwordDigest := md5.New().Sum([]byte(params.Password))
	newUser := User{
		Email:		params.Email,
		PasswordDigest:	string(passwordDigest),
		FavoriteCake:	params.FavoriteCake,
	}
	err = u.repository.Add(params.Email, newUser)
	if err != nil {
		handleError(err, w)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("registered"))
}

func handleError(err error, w http.ResponseWriter) {
	w.WriteHeader(http.StatusUnprocessableEntity)
	w.Write([]byte(err.Error()))
}

func (us *UserService) UpdateCake(w http.ResponseWriter, r *http.Request, user User) {
	params := &CakeUpdate{}
	err := json.NewDecoder(r.Body).Decode(params)
	if err != nil {
		handleError(errors.New("could not read params"), w)
		return
	}

	err = validateCake(params.FavoriteCake)
	if err != nil {
		handleError(err, w)
		return
	}

	user.FavoriteCake = params.FavoriteCake
	err = us.repository.Update(user.Email, user)
	if err != nil {
		handleError(err, w)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("updated"))
}

func (us *UserService) UpdateEmail(w http.ResponseWriter, r *http.Request, user User) {
	params := &EmailUpdate{}
	err := json.NewDecoder(r.Body).Decode(params)
	if err != nil {
		handleError(errors.New("could not read params"), w)
		return
	}

	err = validateEmail(params.Email)
	if err != nil {
		handleError(err, w)
		return
	}

	_, err = us.repository.Delete(user.Email)
	if err != nil {
		handleError(err, w)
		return
	}
	user.Email = params.Email
	err = us.repository.Add(user.Email, user)
	if err != nil {
		handleError(err, w)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("updated"))
}

func (us *UserService) UpdatePassword(w http.ResponseWriter, r *http.Request, user User) {
	params := &PasswordUpdate{}
	err := json.NewDecoder(r.Body).Decode(params)
	if err != nil {
		handleError(errors.New("could not read params"), w)
		return
	}

	if err := validatePassword(params.Password); err != nil {
		handleError(err, w)
		return
	}

	passwordDigest := md5.New().Sum([]byte(params.Password))
	user.PasswordDigest = string(passwordDigest)

	err = us.repository.Update(user.Email, user)
	if err != nil {
		handleError(err, w)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("updated"))
}

func (us *UserService) GetCake(wr http.ResponseWriter, req *http.Request, user User) {
	user.PasswordDigest = ""
	out, err := json.Marshal(user)
	if err != nil {
		handleError(errors.New("could not encode response"), wr)
		return
	}

	wr.WriteHeader(http.StatusOK)
	wr.Write(out)
}

