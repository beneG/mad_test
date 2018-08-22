package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"./currency"
	"./storage"
	"./task"
	"./user"
	"github.com/gorilla/mux"
)

func loginHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(1000000)
	username := r.FormValue("login")
	password := r.FormValue("password")

	_, token, errorCode := storage.Login(username, password)
	w.Header().Set("Content-Type", "application/json")
	if errorCode == 0 {
		answer := fmt.Sprintf(`{"token": "%s", "errMsg": "OK", "errCode": 0}`, token)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(answer))
	} else if errorCode == 1 {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"errMsg": "username and password doesn't match", "errCode": 1}`))
	} else if errorCode == 2 {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"errMsg": "user does not registred", "errCode": 2}`))
	}
}

func getLoginInfo(r *http.Request) (isAuthorized bool, userInfo *storage.LoggedInUserStruct) {
	var token string

	authStrings := strings.Split(r.Header.Get("Authorization"), " ")
	if len(authStrings) == 2 && authStrings[0] == "Bearer" {
		token = authStrings[1]
	} else {
		return false, nil
	}
	userInfo, tokenExist := storage.TokenLookup(token)
	if !tokenExist {
		return false, nil
	}
	return true, userInfo
}

func usersHandler(w http.ResponseWriter, r *http.Request) {
	isAuthorized, userInfo := getLoginInfo(r)
	if !isAuthorized || userInfo == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	processingUserName := vars["slug"]
	w.Header().Set("Content-Type", "application/json")

	if userInfo.User.IsAdmin || userInfo.User.UserName == processingUserName {
		if r.Method == "PUT" {
			if storage.IsUserRegistred(processingUserName) {
				// TODO: edit user, including balance changes
			} else {
				// TODO: create new user
			}
		} else if r.Method == "DELETE" {
			if storage.DeleteUserByName(processingUserName) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"errMsg": "OK", "errCode": 0}`))
			} else {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`{"errMsg": "Cannot delet user", "errCode": 4}`))
			}
		}
	} else {
		if r.Method == "PUT" || r.Method == "DELETE" {
			w.WriteHeader(http.StatusNotModified)
			w.Write([]byte(`{"errMsg": "insufficient permission to change user data", "errCode": 3}`))
			return
		}
	}

	if r.Method == "GET" {
		var u *user.User
		if processingUserName == "" {
			u = &userInfo.User
		} else {
			u = storage.GetUserByName(processingUserName)
		}
		if u == nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"errMsg": "username not found", "errCode": 4}`))
			return
		}
		answer := fmt.Sprintf(`{"login": "%s", "admin": %t, "email": "%s", "balance": %f, "errCode": 0}`, u.UserName, u.IsAdmin, u.Email, u.Balance.GetVal())
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(answer))
	}
}

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	isAuthorized, userInfo := getLoginInfo(r)
	if !isAuthorized || userInfo == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	processingTaskID := vars["task_id"]

	w.Header().Set("Content-Type", "application/json")
	if r.Method == "GET" {
		if processingTaskID != "" {
			// return specific task
			taskID, _ := strconv.Atoi(processingTaskID)
			task := storage.GetTask(taskID)
			answer := fmt.Sprintf(`{"id": %d, "title": "%s", "customer_id": %d, "executioner_id": %d, "state": %d, "cost": %f}`, task.ID, task.Title, task.CustomerID, task.ExecutionerID, task.State, task.Cost.GetVal())
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(answer))
		} else {
			// return all tasks
			tasks := storage.GetAllTasks()
			answer := "["
			isFirstRecord := true
			for _, task := range *tasks {
				taskStr := fmt.Sprintf(`{"id": %d, "title": "%s", "customer_id": %d, "executioner_id": %d, "state": %d, "cost": %f}`, task.ID, task.Title, task.CustomerID, task.ExecutionerID, task.State, task.Cost.GetVal())
				if isFirstRecord {
					isFirstRecord = false
				} else {
					answer += ","
				}
				answer += taskStr
			}
			answer += "]"
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(answer))
		}
	} else if r.Method == "PUT" {
		if processingTaskID != "" {
			// update task
			taskID, _ := strconv.Atoi(processingTaskID)
			t := storage.GetTask(taskID)

			if userInfo.User.IsAdmin || (userInfo.User.ID == t.CustomerID && t.State == task.StateFree) {
				cost, _ := strconv.ParseFloat(vars["cost"], 64)
				t.Title = vars["title"]
				t.Cost = currency.MoneyCtr(cost)
				t.Problem = vars["problem"]
			}
		} else {
			//create new task
			cost, _ := strconv.ParseFloat(vars["cost"], 64)

			task := task.Task{
				ID:            storage.GetNextTaskID(),
				CustomerID:    userInfo.User.ID,
				ExecutionerID: 0,
				Title:         vars["title"],
				State:         task.StateFree,
				Cost:          currency.MoneyCtr(cost),
				Problem:       vars["problem"]}

			(*storage.GetAllTasks())[task.ID] = task
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errMsg": "task updated", "errCode": 0}`))

	} else if r.Method == "DELETE" {
		// TODO: need to implement deleting tasks
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(`{"errMsg": "task removing is not implemented", "errCode": 5}`))
	}
}

// This function processess available commands
func taskCommandHandler(w http.ResponseWriter, r *http.Request) {
	// allowed commands: acquire, finish, accept, close
	isAuthorized, userInfo := getLoginInfo(r)
	if !isAuthorized || userInfo == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	processingTaskID := vars["task_id"]
	command := vars["command"]
	taskID, _ := strconv.Atoi(processingTaskID)

	t := storage.GetTask(taskID)
	if t == nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"errMsg": "task not found", "errCode": 8}`))
	}

	w.Header().Set("Content-Type", "application/json")

	if command == "acquire" {
		if t.State != task.StateFree {
			w.WriteHeader(http.StatusNotModified)
			w.Write([]byte(`{"errMsg": "task is not in free status", "errCode": 7}`))
			return
		}
		activeAmount := userInfo.User.Balance
		activeAmount.Sub(userInfo.User.FrozenAmount)
		if t.Cost.IsGreaterThan(activeAmount) {
			w.WriteHeader(http.StatusNotModified)
			w.Write([]byte(`{"errMsg": "isufficient amout of money on users account", "errCode": 9}`))
			return
		}
		u := storage.GetUserByID(userInfo.User.ID)
		u.FrozenAmount.Add(t.Cost)
		t.State = task.StateExecuting
		t.ExecutionerID = u.ID
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errMsg": "OK", "errCode": 0}`))
	} else if command == "finish" {
		if t.ExecutionerID != userInfo.User.ID {
			w.WriteHeader(http.StatusNotModified)
			w.Write([]byte(`{"errMsg": "task is not acquired by user", "errCode": 11}`))
			return
		}
		if t.State != task.StateExecuting {
			w.WriteHeader(http.StatusNotModified)
			w.Write([]byte(`{"errMsg": "task is not in executing status", "errCode": 10}`))
			return
		}
		t.State = task.StateCompleted
		t.Solution = vars["solution"]
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errMsg": "OK", "errCode": 0}`))
	} else if command == "accept" {
		if t.CustomerID != userInfo.User.ID {
			w.WriteHeader(http.StatusNotModified)
			w.Write([]byte(`{"errMsg": "task is not created by user", "errCode": 12}`))
			return
		}
		if t.State != task.StateCompleted {
			w.WriteHeader(http.StatusNotModified)
			w.Write([]byte(`{"errMsg": "task is not in completed status", "errCode": 13}`))
			return
		}
		t.State = task.StateAccepted
		client := storage.GetUserByID(userInfo.User.ID)
		executioner := storage.GetUserByID(t.ExecutionerID)
		executioner.Balance.Add(t.Cost)
		client.Balance.Sub(t.Cost)
		client.FrozenAmount.Sub(t.Cost)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errMsg": "OK", "errCode": 0}`))
	} else if command == "close" {
		if t.CustomerID != userInfo.User.ID {
			w.WriteHeader(http.StatusNotModified)
			w.Write([]byte(`{"errMsg": "task is not created by user", "errCode": 12}`))
			return
		}
		if t.State != task.StateFree {
			w.WriteHeader(http.StatusNotModified)
			w.Write([]byte(`{"errMsg": "task is not in free status", "errCode": 7}`))
			return
		}
		t.State = task.StateClosed
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errMsg": "OK", "errCode": 0}`))
	} else {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(`{"errMsg": "unexeptable command", "errCode": 6}`))
	}
}

func main() {
	storage.Init()
	r := mux.NewRouter()
	r.HandleFunc("/api/v1/login", loginHandler).Methods("POST")
	r.HandleFunc("/api/v1/users/{slug}", usersHandler).Methods("GET", "PUT", "POST", "DELETE")
	r.HandleFunc("/api/v1/tasks", tasksHandler).Methods("GET")
	r.HandleFunc("/api/v1/tasks/{task_id}", tasksHandler).Methods("GET", "PUT", "POST", "DELETE")
	r.HandleFunc("/api/v1/tasks/{task_id}/{command}", taskCommandHandler).Methods("POST")

	log.Println("starting server at 8000 port")
	log.Fatal(http.ListenAndServe(":8000", r))
}
