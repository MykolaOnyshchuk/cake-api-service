package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"crypto/md5"
	"regexp"
)

type parsedResponse struct {
	status	int
	body	[]byte
}

func createRequester(t *testing.T) func(req *http.Request, err error) parsedResponse {
	return func(req *http.Request, err error) parsedResponse {
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return parsedResponse{}
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return parsedResponse{}
		}
		resp, err := io.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return parsedResponse{}
		}
		return parsedResponse{res.StatusCode, resp}
	}
}

func prepareParams(t *testing.T, params map[string]interface{}) io.Reader {
	body, err := json.Marshal(params)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	return bytes.NewBuffer(body)
}
func newTestUserService() *UserService {
	return &UserService{
		repository: NewInMemoryUserStorage(),
	}
}

func assertStatus(t *testing.T, expected int, r parsedResponse) {
	if r.status != expected {
		t.Errorf("Unexpected response status. Expected: %d, actual: %d", expected, r.status)
	}
}

func assertBody(t *testing.T, expected string, r parsedResponse) {
	actual := string(r.body)
	if actual != expected {
		t.Errorf("Unexpected response body. Expected: %s, actual: %s", expected, actual)
	}
}

func assertBodyRegex(t *testing.T, rx string, r parsedResponse) {
	rex, err := regexp.Compile(rx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !rex.Match(r.body) {
		t.Errorf("Unexpected response body. Expected regexp: %s, actual: %s", rx, string(r.body))
	}
}

func TestUsers_JWT(t *testing.T) {
	doRequest := createRequester(t)
	t.Run("user does not exist", func(t *testing.T) {
		u := newTestUserService()
		j, err := NewJWTService("pubkey.rsa", "privkey.rsa")
		if err != nil {
			t.FailNow()
		}
		ts := httptest.NewServer(http.HandlerFunc(wrapJwt(j, u.JWT)))
		defer ts.Close()
		params := map[string]interface{}{
			"email":
			"test@mail.com",
			"password": "somepass",
		}
		resp := doRequest(http.NewRequest(http.MethodPost, ts.URL, prepareParams(t, params)))
		assertStatus(t, 422, resp)
		assertBody(t, "invalid login params", resp)
	})
	t.Run("wrong password", func(t *testing.T) {
		t.Skip()
	})
}

func TestRegisterWithInvalidEmail(t *testing.T) {
	createReq := createRequester(t)

	t.Run("invalid email", func(t *testing.T) {
		user := newTestUserService()

		ts := httptest.NewServer(http.HandlerFunc(user.Register))
		defer ts.Close()

		params := map[string]interface{}{
			"email":         "email",
			"password":      "qwerty123",
			"favorite_cake": "citrus",
		}

		request, err := http.NewRequest(http.MethodPost, ts.URL, prepareParams(t, params))
		resp := createReq(request, err)
		assertStatus(t, 422, resp)
		assertBodyRegex(t, "missing '@' or angle-addr", resp)
	})
}

func TestRegisterWithInvalidCake(t *testing.T) {
	createReq := createRequester(t)

	t.Run("invalid cake field", func(t *testing.T) {
		user := newTestUserService()

		ts := httptest.NewServer(http.HandlerFunc(user.Register))
		defer ts.Close()

		params := map[string]interface{}{
			"email":         "email@gmail.com",
			"password":      "qwerty123",
			"favorite_cake": "",
		}

		request, err := http.NewRequest(http.MethodPost, ts.URL, prepareParams(t, params))
		resp := createReq(request, err)
		assertStatus(t, 422, resp)
		assertBodyRegex(t, "cake field is empty", resp)
	})
}

func TestRegisterWithInvalidPassword(t *testing.T) {
	createReq := createRequester(t)

	t.Run("not secure password", func(t *testing.T) {
		user := newTestUserService()

		ts := httptest.NewServer(http.HandlerFunc(user.Register))
		defer ts.Close()

		params := map[string]interface{}{
			"email":         "email@gmail.com",
			"password":      "1234",
			"favorite_cake": "citrus",
		}

		request, err := http.NewRequest(http.MethodPost, ts.URL, prepareParams(t, params))
		resp := createReq(request, err)
		assertStatus(t, 422, resp)
		assertBodyRegex(t, "password length is less than 8", resp)
	})
}

func TestCakeUpdate(t *testing.T) {
	createReq := createRequester(t)

	t.Run("updates cake", func(t *testing.T) {
		us := newTestUserService()
		j, err := NewJWTService("public.rsa", "privkey.rsa")
		if err != nil {
			t.FailNow()
		}

		userParams := UserRegisterParams{
			Email:        "email@gmail.com",
			Password:     "qwerty123",
			FavoriteCake: "citrus",
		}

		passwordHashing := md5.New().Sum([]byte(userParams.Password))
		user := User{
			Email:          userParams.Email,
			PasswordDigest: string(passwordHashing),
			FavoriteCake:   userParams.FavoriteCake,
		}
		err = us.repository.Add(userParams.Email, user)
		if err != nil {
			t.Errorf(err.Error())
		}

		jwt, err := j.GenearateJWT(user)
		if err != nil {
			t.Errorf(err.Error())
		}

		ts := httptest.NewServer(http.HandlerFunc(j.AuthenticationJWT(us.repository, us.UpdateCake)))
		defer ts.Close()

		params := map[string]interface{}{
			"favorite_cake": "toffee",
		}

		request, err := http.NewRequest(http.MethodGet, ts.URL, prepareParams(t, params))
		resp := createReq(request, err)
		assertStatus(t, 200, resp)
		assertBody(t, "updated", resp)

		usr, err := us.repository.Get(userParams.Email)
		if err != nil {
			t.Errorf(err.Error())
		}
		if params["favorite_cake"] != usr.FavoriteCake {
			t.Errorf("updated info do not match")
		}
	})
}

func TestEmailUpdate(t *testing.T) {
	createReq := createRequester(t)

	t.Run("updates cake", func(t *testing.T) {
		us := newTestUserService()
		j, err := NewJWTService("public.rsa", "privkey.rsa")
		if err != nil {
			t.FailNow()
		}

		userParams := UserRegisterParams{
			Email:        "email@gmail.com",
			Password:     "qwerty123",
			FavoriteCake: "citrus",
		}

		passwordHashing := md5.New().Sum([]byte(userParams.Password))
		user := User{
			Email:          userParams.Email,
			PasswordDigest: string(passwordHashing),
			FavoriteCake:   userParams.FavoriteCake,
		}
		err = us.repository.Add(userParams.Email, user)
		if err != nil {
			t.Errorf(err.Error())
		}

		jwt, err := j.GenearateJWT(user)
		if err != nil {
			t.Errorf(err.Error())
		}

		ts := httptest.NewServer(http.HandlerFunc(j.AuthenticationJWT(us.repository, us.UpdateEmail)))
		defer ts.Close()

		params := map[string]interface{}{
			"email": "newemail@gmail.com",
		}

		request, err := http.NewRequest(http.MethodPut, ts.URL, prepareParams(t, params))
		resp := createReq(request, err)
		assertStatus(t, 200, resp)
		assertBody(t, "updated", resp)

		_, err = us.repository.Get(params["email"].(string))
		if err != nil {
			t.Errorf(err.Error())
		}
	})
}

func TestPasswordUpdate(t *testing.T) {
	createReq := createRequester(t)

	t.Run("updates password", func(t *testing.T) {
		us := newTestUserService()
		j, err := NewJWTService("public.rsa", "privkey.rsa")
		if err != nil {
			t.FailNow()
		}

		userParams := UserRegisterParams{
			Email:        "email@gmail.com",
			Password:     "qwerty123",
			FavoriteCake: "citrus",
		}

		PasswordHashing1 := md5.New().Sum([]byte(userParams.Password))
		user := User{
			Email:          userParams.Email,
			PasswordDigest: string(PasswordHashing1),
			FavoriteCake:   userParams.FavoriteCake,
		}
		err = us.repository.Add(userParams.Email, user)
		if err != nil {
			t.Errorf(err.Error())
		}

		jwt, err := j.GenearateJWT(user)
		if err != nil {
			t.Errorf(err.Error())
		}

		ts := httptest.NewServer(http.HandlerFunc(j.AuthenticationJWT(us.repository, us.UpdatePassword)))
		defer ts.Close()

		params := map[string]interface{}{
			"password": "QWERTy123",
		}

		request, err := http.NewRequest(http.MethodPut, ts.URL, prepareParams(t, params))
		resp := createReq(request, err)
		assertStatus(t, 200, resp)
		assertBody(t, "updated", resp)

		usr, err := us.repository.Get(userParams.Email)
		if err != nil {
			t.Errorf(err.Error())
		}

		newPasswordHashing := md5.New().Sum([]byte(params["password"].(string)))
		if usr.PasswordDigest != string(newPasswordHashing) {
			t.Errorf("password is not updated")
		}
	})
}

func TestCakeGetting(t *testing.T) {
	doRequest := createRequester(t)

	t.Run("", func(t *testing.T) {
		us := newTestUserService()
		j, err := NewJWTService("public.rsa", "privkey.rsa")
		if err != nil {
			t.FailNow()
		}

		userParams := UserRegisterParams{
			Email:        "example@gmail.com",
			Password:     "qwerty123",
			FavoriteCake: "citrus",
		}

		passwordDigest := md5.New().Sum([]byte(userParams.Password))
		user := User{
			Email:          userParams.Email,
			PasswordDigest: string(passwordDigest),
			FavoriteCake:   userParams.FavoriteCake,
		}
		err = us.repository.Add(userParams.Email, user)
		if err != nil {
			t.Errorf(err.Error())
		}

		jwt, err := j.GenearateJWT(user)
		if err != nil {
			t.Errorf(err.Error())
		}

		ut, err := us.repository.Get(user.Email)

		ut.PasswordDigest = ""
		out, err := json.Marshal(ut)

		ts := httptest.NewServer(http.HandlerFunc(j.AuthenticationJWT(us.repository, us.GetCake)))
		defer ts.Close()

		params := map[string]interface{}{}

		request, err := http.NewRequest(http.MethodPut, ts.URL, prepareParams(t, params))
		resp := doRequest(request, err)
		assertStatus(t, 200, resp)
		assertBody(t, string(out), resp)
	})
}
