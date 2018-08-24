package storage

import (
	"crypto/md5"
	"crypto/rand"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type LoggedInUserStruct struct {
	User      *User
	LoginTime time.Time
}

var loggedInUsers = make(map[string]LoggedInUserStruct) // key is token

func init() {
	go expireLoggedInUsers()
}

func expireLoggedInUsers() {
	for {
		for key, s := range loggedInUsers {
			elapsed := time.Now().Sub(s.LoginTime)
			if elapsed.Minutes() > 30 {
				delete(loggedInUsers, key)
			}
		}
		time.Sleep(10 * time.Second)
	}
}

func loggedInUserLookup(username string) (token string, isLoggedIn bool) {
	for tok, u := range loggedInUsers {
		if u.User.UserName == username {
			u.LoginTime = time.Now()
			return tok, true
		}
	}
	return "", false
}

func generateToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func Login(username, password string) (user *User, token string, errCode int) {
	if tok, isUserLoggedIn := loggedInUserLookup(username); isUserLoggedIn {
		return loggedInUsers[tok].User, tok, 0
	}
	user, isUserPresent := GetUserByName(username)
	if !isUserPresent {
		return nil, "", 1 // there is no such user
	}
	hash := fmt.Sprintf("%x", md5.Sum([]byte(password)))
	if hash != user.PasswordHash {
		return nil, "", 2 // username/password mismatch
	}
	token = generateToken()
	loggedInUsers[token] = LoggedInUserStruct{
		User:      user,
		LoginTime: time.Now()}
	return user, token, 0
}

func Auth(r *http.Request) (user *User, isAuthorized bool) {
	var token string
	authStrings := strings.Split(r.Header.Get("Authorization"), " ")
	if len(authStrings) == 2 && authStrings[0] == "Bearer" {
		token = authStrings[1]
	} else {
		return nil, false
	}
	if userInfo, ok := loggedInUsers[token]; ok {
		return userInfo.User, true
	}
	return nil, false
}
