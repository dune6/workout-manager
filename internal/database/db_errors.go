package database

import "errors"

var (
	ErrorUserExist         = errors.New("user already exists")
	ErrorUserNotFound      = errors.New("user not found")
	ErrorUserInsert        = errors.New("something went wrong with inserting user")
	ErrorSomethingGetWrong = errors.New("something went wrong with getting user")
)
