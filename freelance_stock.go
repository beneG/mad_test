package main

import (
	"crypto/md5"
	"net/http"
	"strconv"
	"time"

	"./currency"
	"./storage"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"fmt"
)

func main() {
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.POST("/api/v1/login", loginHandler)
	e.GET("/api/v1/users/:slug", usersHandlerGet)
	e.POST("/api/v1/users", usersHandlerCreate)
	e.PUT("/api/v1/users/:slug", usersHandlerUpdate)
	e.DELETE("/api/v1/users/:slug", usersHandlerDelete)

	e.GET("/api/v1/tasks/:task_id", tasksHandlerGet)
	e.GET("/api/v1/tasks", tasksHandlerGet)
	e.POST("/api/v1/tasks", tasksHandlerCreate)
	e.PUT("/api/v1/tasks/:task_id", tasksHandlerUpdate)
	e.DELETE("/api/v1/tasks/:task_id", tasksHandlerDelete)

	e.POST("/api/v1/tasks/:task_id/:command", taskCommandHandler)

	// Start server
	e.Logger.Fatal(e.Start(":8000"))

}

func loginHandler(c echo.Context) error {
	username := c.FormValue("login")
	password := c.FormValue("password")

	_, token, errorCode := storage.Login(username, password)
	var answer string
	var status = http.StatusUnauthorized
	switch errorCode {
	case 0:
		answer = fmt.Sprintf(`{"token": "%s", "errror_message": "OK", "error_code": 0}`, token)
		status = http.StatusOK
	case 1:
		answer = `{"errror_message": "username and password doesn't match", "error_code": 1}`
	case 2:
		answer = `{"errror_message": "user does not registred", "error_code": 2}`
	default:
		answer = `{"errror_message": "unknown error", "error_code": 127}`
	}
	return c.JSON(status, answer)
}

func usersHandlerGet(c echo.Context) error {
	u, isAuthorized := storage.Auth(c.Request())
	if !isAuthorized {
		return c.String(http.StatusUnauthorized, "")
	}
	userName := c.Param("slug")
	if u.UserName == userName || u.IsAdmin {
		answer := fmt.Sprintf(`{"login": "%s", "admin": %t, "email": "%s", "balance": %f, "error_code": 0}`, u.UserName, u.IsAdmin, u.Email, u.Balance.GetVal())
		return c.JSON(http.StatusOK, answer)
	}
	return c.JSON(http.StatusNotModified, `{"error_message": "insufficient permission view user data", "error_code": 3}`)
}

func usersHandlerCreate(c echo.Context) error {
	u, isAuthorized := storage.Auth(c.Request())
	if !isAuthorized {
		return c.String(http.StatusUnauthorized, "")
	}
	if u.IsAdmin {
		isAdmin, err := strconv.ParseBool(c.FormValue("is_admin"))
		if err != nil {
			return c.JSON(http.StatusNotModified, `{"error_message": "is_admin param must be true or false", "error_code": 127}`)
		}
		userName := c.FormValue("user_name")
		passwordHash := fmt.Sprintf("%x", md5.Sum([]byte(c.FormValue("password"))))
		email := c.FormValue("email")

		id, created := storage.CreateNewUser(isAdmin, userName, passwordHash, email)
		if created {
			answer := fmt.Sprintf(`{"id": "%d", "error_message": "new user created", "error_code": 0}`, id)
			return c.JSON(http.StatusCreated, answer)
		}
		return c.JSON(http.StatusNotModified, `{"error_message": "user not created", "error_code": 127}`)
	}
	return c.JSON(http.StatusNotModified, `{"error_message": "insufficient permission to create new user", "error_code": 3}`)
}

func usersHandlerUpdate(c echo.Context) error {
	u, isAuthorized := storage.Auth(c.Request())
	if !isAuthorized {
		return c.String(http.StatusUnauthorized, "")
	}
	var isAdmin bool
	userName := c.Param("slug")
	editingUser, isUserPresent := storage.GetUserByName(userName)
	if !isUserPresent {
		return c.JSON(http.StatusNotFound, `{"error_message": "user not found", "error_code": 123}`)
	}
	if u.IsAdmin {
		if c.FormValue("is_admin") == "" {
			isAdmin = editingUser.IsAdmin
		} else if isAdminParam, err := strconv.ParseBool(c.FormValue("is_admin")); err != nil {
			return c.JSON(http.StatusNotModified, `{"error_message": "is_admin param must be true or false", "error_code": 127}`)
		} else {
			isAdmin = isAdminParam
		}
	}
	if u.UserName == userName || u.IsAdmin {
		editingUser.IsAdmin = isAdmin
		if password := c.FormValue("password"); password != "" {
			editingUser.PasswordHash = fmt.Sprintf("%x", md5.Sum([]byte(password)))
		}
		if email := c.FormValue("email"); email != "" {
			editingUser.Email = email
		}
		if balanceStr := c.FormValue("balance"); balanceStr != "" {
			balance, err := strconv.ParseFloat(balanceStr, 64)
			if err != nil {
				return c.JSON(http.StatusNotModified, `{"error_message": "balance param must be in float number format", "error_code": 127}`)
			}
			editingUser.Balance.SetVal(balance)
		}
		if frozenStr := c.FormValue("frozen_amount"); frozenStr != "" {
			frozenAmount, err := strconv.ParseFloat(frozenStr, 64)
			if err != nil {
				return c.JSON(http.StatusNotModified, `{"error_message": "frozen_amount param must be in float number format", "error_code": 127}`)
			}
			editingUser.FrozenAmount.SetVal(frozenAmount)
		}
		storage.UpdateUser(editingUser)
		return c.JSON(http.StatusOK, `{"error_message": "user updated", "error_code": 0}`)
	}
	return c.JSON(http.StatusNotModified, `{"error_message": "insufficient permission to create new user", "error_code": 3}`)
}

func usersHandlerDelete(c echo.Context) error {
	u, isAuthorized := storage.Auth(c.Request())
	if !isAuthorized {
		return c.String(http.StatusUnauthorized, "")
	}
	if !u.IsAdmin {
		return c.JSON(http.StatusOK, `{"error_message": "insufficient privileges to delete user", "error_code": 125}`)
	}
	userName := c.Param("slug")
	editingUser, isUserPresent := storage.GetUserByName(userName)
	if !isUserPresent {
		return c.JSON(http.StatusNotFound, `{"error_message": "user not found", "error_code": 122}`)
	}
	if u.UserName == editingUser.UserName {
		return c.JSON(http.StatusNotModified, `{"error_message": "user can't delete himself", "error_code": 125}`)
	}
	if storage.DeleteUser(editingUser.ID) {
		return c.JSON(http.StatusOK, `{"error_message": "user deleted", "error_code": 0}`)
	}
	return c.JSON(http.StatusNotModified, `{"error_message": "user not deleted", "error_code": 127}`)
}

func tasksHandlerGet(c echo.Context) error {
	/*u*/ _, isAuthorized := storage.Auth(c.Request())
	if !isAuthorized {
		return c.String(http.StatusUnauthorized, "")
	}
	if c.Param("task_id") == "" {
		//return all tasks exept closed ones
	}
	taskID, err := strconv.ParseInt(c.Param("task_id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, `{"error_message": "task_id must be integer type", "error_code": 10}`)
	}
	task, isTaskPresent := storage.GetTaskByID(int(taskID))
	if !isTaskPresent {
		answer := fmt.Sprintf(`{"error_message": "task with id=%d not found in database", "error_code": 11}`, taskID)
		return c.JSON(http.StatusNotFound, answer)
	}
	answer := fmt.Sprintf(`{"id": %d, "title": "%s", "customer_id": %d, "executioner_id": %d, "state": %d, "cost": %f}`,
		task.ID,
		task.Title,
		task.CustomerID,
		task.ExecutionerID,
		task.State,
		task.Cost.GetVal())
	return c.JSON(http.StatusOK, answer)
}

func tasksHandlerCreate(c echo.Context) error {
	u, isAuthorized := storage.Auth(c.Request())
	if !isAuthorized {
		return c.String(http.StatusUnauthorized, "")
	}
	title := c.FormValue("title")
	cost, err := strconv.ParseFloat(c.FormValue("cost"), 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, `{"error_message": "parameter cost must be in float number format", "error_code": 20}`)
	}
	problem := c.FormValue("problem")
	taskID, _ := storage.CreateNewTask(u.ID, title, currency.MoneyCtr(cost), problem)
	answer := fmt.Sprintf(`{"error_message": "task with id=%d has been created", "error_code": 0}`, taskID)
	return c.JSON(http.StatusBadRequest, answer)
}

func tasksHandlerUpdate(c echo.Context) error {
	u, isAuthorized := storage.Auth(c.Request())
	if !isAuthorized {
		return c.String(http.StatusUnauthorized, "")
	}
	taskID, err := strconv.ParseInt(c.Param("task_id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, `{"error_message": "task_id must be integer type", "error_code": 10}`)
	}
	t, isTaskPresent := storage.GetTaskByID(int(taskID))
	if !isTaskPresent {
		answer := fmt.Sprintf(`{"error_message": "task with id=%d not found in database", "error_code": 11}`, taskID)
		return c.JSON(http.StatusNotFound, answer)
	}
	if u.IsAdmin {
		// full control
		if customerID := c.FormValue("customer_id"); customerID != "" {
			cID, _ := strconv.ParseInt(customerID, 10, 64)
			t.CustomerID = int(cID)
		}
		if executionerID := c.FormValue("executor_id"); executionerID != "" {
			eID, _ := strconv.ParseInt(executionerID, 10, 64)
			t.ExecutionerID = int(eID)
		}
		if title := c.FormValue("title"); title != "" {
			t.Title = title
		}
		if stateStr := c.FormValue("state"); stateStr != "" {
			state, _ := strconv.ParseInt(stateStr, 10, 64)
			t.ExecutionerID = int(state)
		}
		if costStr := c.FormValue("cost"); costStr != "" {
			cost, err := strconv.ParseFloat(costStr, 64)
			if err != nil {
				return c.JSON(http.StatusNotModified, `{"error_message": "cost param must be in float number format", "error_code": 127}`)
			}
			t.Cost.SetVal(cost)
		}
		if problem := c.FormValue("problem"); problem != "" {
			t.Problem = problem
		}
		if solution := c.FormValue("solution"); solution != "" {
			t.Solution = solution
		}
		if beginTimeStr := c.FormValue("begin_time"); beginTimeStr != "" {
			t.BeginTime, err = time.Parse("2006-01-02 15:04:05", beginTimeStr)
			if err != nil {
				return c.JSON(http.StatusNotModified, `{"error_message": "begin_time must be like 2006-01-02 15:04:05", "error_code": 120}`)
			}
		}
		if endTimeStr := c.FormValue("end_time"); endTimeStr != "" {
			t.EndTime, err = time.Parse("2006-01-02 15:04:05", endTimeStr)
			if err != nil {
				return c.JSON(http.StatusNotModified, `{"error_message": "end_time must be like 2006-01-02 15:04:05", "error_code": 121}`)
			}
		}
	} // if u.IsAdmin
	if u.ID == t.CustomerID {
		// partial control
		if title := c.FormValue("title"); title != "" {
			t.Title = title
		}
		if t.State == storage.StateFree {
			if costStr := c.FormValue("cost"); costStr != "" {
				cost, err := strconv.ParseFloat(costStr, 64)
				if err != nil {
					return c.JSON(http.StatusNotModified, `{"error_message": "cost param must be in float number format", "error_code": 127}`)
				}
				t.Cost.SetVal(cost)
			}
			if problem := c.FormValue("problem"); problem != "" {
				t.Problem = problem
			}
		}
	} // u.ID == t.CustomerID
	storage.UpdateTask(t)
	answer := fmt.Sprintf(`{"error_message": "task with id=%d has been updated", "error_code": 0}`, t.ID)
	return c.JSON(http.StatusOK, answer)
}

func tasksHandlerDelete(c echo.Context) error {
	u, isAuthorized := storage.Auth(c.Request())
	if !isAuthorized {
		return c.String(http.StatusUnauthorized, "")
	}
	taskID, err := strconv.ParseInt(c.Param("task_id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, `{"error_message": "task_id must be integer type", "error_code": 10}`)
	}
	t, isTaskPresent := storage.GetTaskByID(int(taskID))
	if !isTaskPresent {
		answer := fmt.Sprintf(`{"error_message": "task with id=%d not found in database", "error_code": 11}`, taskID)
		return c.JSON(http.StatusNotFound, answer)
	}
	if !u.IsAdmin || u.ID != t.CustomerID {
		return c.JSON(http.StatusNotModified, `{"error_message": "insufficient privileges to delete task", "error_code": 125}`)
	}
	if t.State != storage.StateFree && !u.IsAdmin {
		return c.JSON(http.StatusNotModified, `{"error_message": "only tasks with status free(0) can be deleted", "error_code": 115}`)
	}
	if storage.DeleteTask(int(taskID)) {
		return c.JSON(http.StatusOK, `{"error_message": "task deleted", "error_code": 0}`)
	}
	return c.JSON(http.StatusNotModified, `{"error_message": "task not deleted", "error_code": 64}`)
}

func taskCommandHandler(c echo.Context) error {
	// allowed commands: acquire, finish, accept, close
	u, isAuthorized := storage.Auth(c.Request())
	if !isAuthorized {
		return c.String(http.StatusUnauthorized, "")
	}
	taskID, err := strconv.ParseInt(c.Param("task_id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, `{"error_message": "task_id must be integer type", "error_code": 10}`)
	}
	t, isTaskPresent := storage.GetTaskByID(int(taskID))
	if !isTaskPresent {
		answer := fmt.Sprintf(`{"error_message": "task with id=%d not found in database", "error_code": 11}`, taskID)
		return c.JSON(http.StatusNotFound, answer)
	}

	command := c.Param("command")

	switch command {
	case "acquire":
		if t.State != storage.StateFree {
			return c.JSON(http.StatusNotModified, `{"error_message": "task is not in free status", "error_code": 32}`)
		}
		activeAmount := u.Balance
		activeAmount.Sub(u.FrozenAmount)
		if t.Cost.IsGreaterThan(activeAmount) {
			return c.JSON(http.StatusNotModified, `{"error_message": "isufficient amout of money on users account", "error_code": 33}`)
		}
		u.FrozenAmount.Add(t.Cost)
		t.State = storage.StateExecuting
		t.ExecutionerID = u.ID
		t.BeginTime = time.Now()

		tx := storage.BeginTransaction()
		storage.UpdateTask(t)
		storage.UpdateUser(u)
		storage.CommitTransaction(tx)

		return c.JSON(http.StatusOK, `{"error_message": "task acquired", "error_code": 0}`)
	case "finish":
		if t.ExecutionerID != u.ID {
			return c.JSON(http.StatusNotModified, `{"error_message": "task is not acquired previously by user", "error_code": 34}`)
		}
		if t.State != storage.StateExecuting {
			return c.JSON(http.StatusNotModified, `{"error_message": "task is not in executing status", "error_code": 35}`)
		}
		t.State = storage.StateCompleted
		t.Solution = c.FormValue("solution")
		t.EndTime = time.Now()
		storage.UpdateTask(t)
		return c.JSON(http.StatusOK, `{"error_message": "task finished", "error_code": 0}`)
	case "accept":
		if t.CustomerID != u.ID {
			return c.JSON(http.StatusNotModified, `{"error_message": "task is not created by this user", "error_code": 36}`)
		}
		if t.State != storage.StateCompleted {
			return c.JSON(http.StatusNotModified, `{"error_message": "task is not in completed status", "error_code": 37}`)
		}
		t.State = storage.StateAccepted
		executioner, _ := storage.GetUserByID(t.ExecutionerID)
		executioner.Balance.Add(t.Cost)
		u.Balance.Sub(t.Cost)
		u.FrozenAmount.Sub(t.Cost)

		tx := storage.BeginTransaction()
		storage.UpdateTask(t)
		storage.UpdateUser(u)
		storage.UpdateUser(executioner)
		storage.CommitTransaction(tx)

		return c.JSON(http.StatusOK, `{"error_message": "task accepted", "error_code": 0}`)
	case "close":
		if t.CustomerID != u.ID {
			return c.JSON(http.StatusNotModified, `{"error_message": "task is not created by this user", "error_code": 46}`)
		}
		if t.State != storage.StateFree {
			return c.JSON(http.StatusNotModified, `{"error_message": "task is not in free status", "error_code": 47}`)
		}
		t.State = storage.StateClosed
		return c.JSON(http.StatusOK, `{"error_message": "task successfully closed", "error_code": 0}`)
	}
	return c.JSON(http.StatusBadRequest, `{"error_message": "unexeptable command", "error_code": 66}`)
}
