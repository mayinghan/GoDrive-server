package db

import (
	"GoDrive/db/mydb"
	"GoDrive/utils"
	"database/sql"
	"fmt"
)

const salt = "&6ty"

// RegInfo is the registration input: username password and email
type RegInfo struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Code     int64  `json:"code,string" binding:"required"`
}

// LoginInfo is the login input : username and password
type LoginInfo struct {
	Username string `json:"input" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// VerifyEmail is the email input during registration
type VerifyEmail struct {
	Email string `json:"email" form:"email" binding:"required"`
}

// CheckEmail checks against the database for an existing email. Returns a bool and server message
func CheckEmail(email *VerifyEmail) (bool, string, error) {
	var compareEmail string
	userEmail := email.Email
	stmt, err := mydb.DBConn().Prepare(
		"select email from tbl_user where email = ?")
	if err != nil {
		e := fmt.Sprint("Internal server error: Failed to retrieve user from DB")
		fmt.Println(e + err.Error())
		return false, e, err
	}
	defer stmt.Close()

	err = stmt.QueryRow(userEmail).Scan(&compareEmail)
	var e string
	if err != nil {
		if err == sql.ErrNoRows {
			e = fmt.Sprint("Email has not been used!")
			return true, e, err
		}
	}
	e = fmt.Sprint("Internal server error: Email exists.")
	fmt.Println(e + err.Error())
	return false, e, err
}

//UserLogin checks against the database for an existing user. Returns a bool and server message
func UserLogin(loginInfo *LoginInfo) (bool, string, error) {

	var comparePwd string
	username := loginInfo.Username
	password := utils.MD5([]byte(loginInfo.Password + salt))

	fmt.Printf("login input, %s\n", username)

	stmt, err := mydb.DBConn().Prepare(
		"select username from tbl_user where (username = ? or email = ?) and password = ?")

	if err != nil {
		e := fmt.Sprint("Internal server error: Failed to retrieve user from DB")
		fmt.Println(e + err.Error())
		return false, e, err
	}
	defer stmt.Close()

	err = stmt.QueryRow(username, username, password).Scan(&comparePwd)

	if err != nil {
		var e string
		if err == sql.ErrNoRows {
			e = fmt.Sprint("Unauthorized error: Failed to find user with that username/password.")
		} else {
			e = fmt.Sprint("Internal server error: Failed to retrieve user from DB.")
		}
		fmt.Println(e + err.Error())
		return false, e, err
	}

	return true, "Successfully logged in!", nil

}

// UserRegister handles user registration. Return a bool and a server message
func UserRegister(regInfo *RegInfo) (bool, string, error) {
	username := regInfo.Username
	password := regInfo.Password
	email := regInfo.Email

	// fmt.Printf("%v\n", regInfo)

	stmt, err := mydb.DBConn().Prepare(
		"insert ignore into tbl_user (`username`, `password`, `email`, `email_validated`) values(?, ?, ?, 1)")

	if err != nil {
		e := fmt.Sprint("Internal server error: Failed to insert to DB.")
		fmt.Println(e + err.Error())
		return false, e, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(username, password, email)
	if err != nil {
		e := fmt.Sprint("Internal server error: Failed to insert to DB.")
		fmt.Println(e + err.Error())
		return false, e, err
	}
	// check how many row is affected
	if ra, err := result.RowsAffected(); err == nil && ra > 0 {
		return true, "Registered Successfully!", nil
	} else if err == nil && ra <= 0 {
		return false, "Failed to register. Duplicated user!", nil
	} else {
		return false, "Internal server error: Failed to insert to DB", err
	}
}
