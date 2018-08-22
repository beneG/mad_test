package user

import "../currency"

// User : structure for user
type User struct {
	ID           int
	IsAdmin      bool
	UserName     string
	PasswordHash string
	Email        string
	Balance      currency.Money
	FrozenAmount currency.Money
}
