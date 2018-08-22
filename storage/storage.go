package storage

import (
	"crypto/md5"
	"crypto/rand"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"../currency"
	"../task"
	"../user"
)

// LoggedInUserStruct : structure to keep user and loggedin time
type LoggedInUserStruct struct {
	User      user.User
	LoginTime time.Time
}

var users map[int]user.User
var tasks map[int]task.Task
var usersMaxID uint32
var tasksMaxID uint32

var loggedInUsers map[string]LoggedInUserStruct

// Init : this func has to be called before using storage
func Init() {
	users = make(map[int]user.User)
	tasks = make(map[int]task.Task)
	loggedInUsers = make(map[string]LoggedInUserStruct)

	// init 3 users
	users[1] = user.User{
		ID:           1,
		IsAdmin:      true,
		UserName:     "admin",
		PasswordHash: "482c811da5d5b4bc6d497ffa98491e38", // md5 hash for "password123" string
		Email:        "admin@domain.com",
		Balance:      currency.MoneyCtr(0.0)}
	users[2] = user.User{
		ID:           2,
		IsAdmin:      false,
		UserName:     "nurbek",
		PasswordHash: "482c811da5d5b4bc6d497ffa98491e38",
		Email:        "nasanbekov@gmail.com",
		Balance:      currency.MoneyCtr(1.000000123)}
	users[3] = user.User{
		ID:           3,
		IsAdmin:      false,
		UserName:     "emil",
		PasswordHash: "482c811da5d5b4bc6d497ffa98491e38",
		Email:        "emilasanbekov@gmail.com",
		Balance:      currency.MoneyCtr(100.12)}

	usersMaxID = 3

	// init 4 tasks
	tasks[1] = task.Task{
		ID:            1,
		CustomerID:    2,
		ExecutionerID: 0,
		Title:         "Make online shop",
		State:         task.StateFree,
		Cost:          currency.MoneyCtr(100.12)}
	tasks[2] = task.Task{
		ID:            2,
		CustomerID:    3,
		ExecutionerID: 0,
		Title:         "Fix bug in network library",
		State:         task.StateFree,
		Cost:          currency.MoneyCtr(200.0001)}
	tasks[3] = task.Task{
		ID:            3,
		CustomerID:    2,
		ExecutionerID: 0,
		Title:         "Create dating site",
		State:         task.StateFree,
		Cost:          currency.MoneyCtr(321.000000123)}
	tasks[4] = task.Task{
		ID:            4,
		CustomerID:    3,
		ExecutionerID: 0,
		Title:         "Create company logo",
		State:         task.StateFree,
		Cost:          currency.MoneyCtr(412.512)}

	tasksMaxID = 4

	log.Println("storage has been filled with sample data")
}

// GetNextUserID : for user.User id field in new records
func GetNextUserID() int {
	atomic.AddUint32(&usersMaxID, 1)
	return int(usersMaxID)
}

// GetNextTaskID : for task.Task id field in new records
func GetNextTaskID() int {
	atomic.AddUint32(&tasksMaxID, 1)
	return int(tasksMaxID)
}

func generateToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// TokenLookup : lookup value in loggedInUsers by token key
func TokenLookup(token string) (u *LoggedInUserStruct, isTokenExist bool) {
	l, exist := loggedInUsers[token]
	if exist {
		return &l, true
	}
	return nil, false
}

func loggedInUserLookup(username string) (token string, isLoggedIn bool) {
	for tok, u := range loggedInUsers {
		if u.User.UserName == username {
			return tok, true
		}
	}
	return "", false
}

// Login : logging in a user
func Login(username, password string) (u *LoggedInUserStruct, token string, errorCode int) {
	for /*id*/ _, user := range users {
		if user.UserName == username {
			hash := fmt.Sprintf("%x", md5.Sum([]byte(password)))
			if hash == user.PasswordHash {
				token, isLoggedIn := loggedInUserLookup(username)
				if !isLoggedIn {
					token = generateToken()
				}
				loggedInUser := LoggedInUserStruct{User: user, LoginTime: time.Now()}
				loggedInUsers[token] = loggedInUser
				return &loggedInUser, token, 0 // errorCode OK
			}
			return nil, "", 1 // errorCode username and password mismatch
		}
	}
	return nil, "", 2 // errorCode username is not registred
}

// IsUserRegistred : check if user is in our storage
func IsUserRegistred(username string) bool {
	for /*id*/ _, user := range users {
		if user.UserName == username {
			return true
		}
	}
	return false
}

// CreateUser : remove user from storage
func CreateUser(login, passwordHash, email string) bool {
	if IsUserRegistred(login) {
		return false
	}
	id := GetNextUserID()
	newUser := user.User{
		ID:           id,
		IsAdmin:      false,
		UserName:     login,
		PasswordHash: passwordHash,
		Email:        email,
		Balance:      currency.MoneyCtr(0.0)}
	users[id] = newUser
	return true
}

// DeleteUserByName : remove user from storage
func DeleteUserByName(username string) bool {
	for id, user := range users {
		if user.UserName == username {
			delete(users, id)
			return true
		}
	}
	return false
}

// UpdateUser : just update user in storage
func UpdateUser(username string, user *user.User) bool {
	for id, u := range users {
		if u.UserName == username {
			users[id] = *user
			return true
		}
	}
	return false
}

// GetUserByName : get user struct by his username
func GetUserByName(username string) *user.User {
	for /*id*/ _, u := range users {
		if u.UserName == username {
			return &u
		}
	}
	return nil
}

// GetUserByID : get user struct by his ID
func GetUserByID(userID int) *user.User {
	user, exist := users[userID]
	if exist {
		return &user
	}
	return nil
}

// GetAllTasks : retruns pointer to tasks map
func GetAllTasks() *map[int]task.Task {
	return &tasks
}

// GetTask : lookup for specific task
func GetTask(taskID int) *task.Task {
	if task, exist := tasks[taskID]; exist {
		return &task
	}
	return nil
}
