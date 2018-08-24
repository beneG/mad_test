package storage

import (
	"database/sql"
	"time"

	"../currency"
	"github.com/VividCortex/mysqlerr"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

const timeStringLayout = "2006-01-02 15:04:05"

type dbUser struct {
	ID           int     `db:"id"`
	IsAdmin      bool    `db:"is_admin"`
	UserName     string  `db:"user_name"`
	PasswordHash string  `db:"password_hash"`
	Email        string  `db:"email"`
	Balance      float64 `db:"balance"`
	FrozenAmount float64 `db:"frozen_amount"`
}

type dbTask struct {
	ID            int     `db:"id"`
	CustomerID    int     `db:"customer_id"`
	ExecutionerID int     `db:"executor_id"`
	Title         string  `db:"title"`
	State         int     `db:"status"`
	Cost          float64 `db:"cost"`
	Problem       string  `db:"problem"`
	Solution      string  `db:"solution"`
	BeginTime     string  `db:"begin_time"`
	EndTime       string  `db:"end_time"`
}

var connection *sqlx.DB

func init() {
	if connection == nil {
		conn, err := sqlx.Connect("mysql", "seth:123@tcp(127.0.0.1:3306)/freelance_stock")
		if err != nil {
			panic(err)
		}
		connection = conn
	}
}

func dbUserToUser(val *dbUser) *User {
	retVal := User{
		ID:           val.ID,
		IsAdmin:      val.IsAdmin,
		UserName:     val.UserName,
		PasswordHash: val.PasswordHash,
		Email:        val.Email,
		Balance:      currency.MoneyCtr(val.Balance),
		FrozenAmount: currency.MoneyCtr(val.FrozenAmount)}
	return &retVal
}

func userToDbUser(val *User) dbUser {
	return dbUser{
		ID:           val.ID,
		IsAdmin:      val.IsAdmin,
		UserName:     val.UserName,
		PasswordHash: val.PasswordHash,
		Email:        val.Email,
		Balance:      val.Balance.GetVal(),
		FrozenAmount: val.FrozenAmount.GetVal()}
}

func dbTaskToTask(val *dbTask) *Task {
	beginTime, _ := time.Parse(timeStringLayout, val.BeginTime)
	endTime, _ := time.Parse(timeStringLayout, val.EndTime)
	retVal := Task{
		ID:            val.ID,
		CustomerID:    val.CustomerID,
		ExecutionerID: val.ExecutionerID,
		Title:         val.Title,
		State:         State(val.State),
		Cost:          currency.MoneyCtr(val.Cost),
		Problem:       val.Problem,
		Solution:      val.Solution,
		BeginTime:     beginTime,
		EndTime:       endTime}
	return &retVal
}

func taskToDbTask(val *Task) dbTask {
	return dbTask{
		ID:            val.ID,
		CustomerID:    val.CustomerID,
		ExecutionerID: val.ExecutionerID,
		Title:         val.Title,
		State:         int(val.State),
		Cost:          val.Cost.GetVal(),
		Problem:       val.Problem,
		Solution:      val.Solution,
		BeginTime:     val.BeginTime.Format(timeStringLayout),
		EndTime:       val.EndTime.Format(timeStringLayout)}
}

// GetTaskByID retruns task structure by it's ID
func GetTaskByID(ID int) (task *Task, isTaskPresent bool) {
	var dbT dbTask
	err := connection.Get(&dbT, "SELECT * FROM tasks WHERE id=?", ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, false
		}
		panic(err)
	}
	return dbTaskToTask(&dbT), true
}

/*
func GetAllTasksNew() (tasks []*task.Task) {
	var dbT dbTask
	selectedTasks []*task.Task
	err := connection.Get(&dbT, "SELECT * FROM tasks WHERE id IN ?, ?, ?, ?",
		0, 1, 2, 3)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		panic(err)
	}
	return dbTaskToTask(&dbT)
}*/

func CreateNewTask(customerID int, title string, cost currency.Money, problem string) (taskID int, isTaskCreated bool) {
	res, err := connection.Exec("INSERT INTO tasks (customer_id, executor_id, title, status, cost, problem, solution) VALUES(?, 0, ?, 0, ?, ?, \"\")",
		customerID,
		title,
		cost.GetVal(),
		problem)
	if err != nil {
		panic(err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		panic(err)
	}
	return int(id), true
}

func UpdateTask(task *Task) {
	dbT := taskToDbTask(task)
	_, err := connection.Exec("UPDATE tasks set customer_id=?, executor_id=?, title=?, status=?, cost=?, problem=?, solution=?, begin_time=?, end_time=? where id=?",
		dbT.CustomerID,
		dbT.ExecutionerID,
		dbT.Title,
		dbT.State,
		dbT.Cost,
		dbT.Problem,
		dbT.Solution,
		dbT.BeginTime,
		dbT.EndTime,
		dbT.ID)
	if err != nil {
		panic(err)
	}
}

func DeleteTask(taskID int) (isDeleted bool) {
	_, err := connection.Exec("DELETE FROM tasks WHERE id=?", taskID)
	if err != nil {
		return false
	}
	return true
}

func GetUserByName(userName string) (user *User, isUserPresent bool) {
	var dbU dbUser
	err := connection.Get(&dbU, "SELECT * FROM users WHERE user_name=?", userName)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, false
		}
		panic(err)
	}
	return dbUserToUser(&dbU), true
}

func GetUserByID(userID int) (user *User, isUserPresent bool) {
	var dbU dbUser
	err := connection.Get(&dbU, "SELECT * FROM users WHERE id=?", userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, false
		}
		panic(err)
	}
	return dbUserToUser(&dbU), true
}

func CreateNewUser(isAdmin bool, userName, passwordHash, email string) (userID int, isUserCreated bool) {
	res, err := connection.Exec("INSERT INTO users (is_admin, user_name, password_hash, email) VALUES(?, ?, ?, ?)",
		isAdmin,
		userName,
		passwordHash,
		email)
	if err != nil {
		if driverErr, ok := err.(*mysql.MySQLError); ok {
			if driverErr.Number == mysqlerr.ER_DUP_ENTRY {
				// in this case we trying to create user with existing user_name
				return 0, false
			}
		}
		panic(err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		panic(err)
	}
	return int(id), true
}

func UpdateUser(user *User) (isUpdated bool) {
	dbU := userToDbUser(user)
	_, err := connection.Exec("UPDATE users set is_admin=?, user_name=?, password_hash=?, email=?, balance=?, frozen_amount=? WHERE id=?",
		dbU.IsAdmin,
		dbU.UserName,
		dbU.PasswordHash,
		dbU.Email,
		dbU.Balance,
		dbU.FrozenAmount,
		dbU.ID)
	if err != nil {
		if driverErr, ok := err.(*mysql.MySQLError); ok {
			if driverErr.Number == mysqlerr.ER_DUP_ENTRY {
				// in this case we trying to update user with other existing user_name
				return false
			}
		}
		panic(err)
	}
	return true
}

func DeleteUser(userID int) (isDeleted bool) {
	_, err := connection.Exec("DELETE FROM users WHERE id=?", userID)
	if err != nil {
		return false
	}
	return true
}
