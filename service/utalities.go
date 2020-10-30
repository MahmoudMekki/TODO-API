package service

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	config "github.com/TODO-API/db_config"

	"github.com/TODO-API/view"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte("MahmoudMekk")

func hashPassword(password string) ([]byte, error) {
	bs, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("error with converting to hash %w", err)
	}
	return bs, nil
}

func createToken(c *view.UserClaims) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS512, c)
	token, err := t.SignedString(jwtKey)
	if err != nil {
		return "", fmt.Errorf("Error while generating token")
	}
	return token, nil
}
func compare(password string, hash []byte) bool {
	err := bcrypt.CompareHashAndPassword(hash, []byte(password))
	if err != nil {
		return false
	}
	return true
}
func parseToken(token string) (*view.UserClaims, error) {
	claims := &view.UserClaims{}
	t, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		if t.Method.Alg() != jwt.SigningMethodHS512.Alg() {
			return nil, fmt.Errorf("Invalid signing algorithm")
		}
		return jwtKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("Error in parsing the token")
	}
	if !t.Valid {
		return nil, fmt.Errorf("Invalid token")
	}
	return t.Claims.(*view.UserClaims), nil
}

func checkToken(req *http.Request) (*view.UserClaims, bool) {
	token := req.Header.Get("Authorization")
	token = strings.Split(token, "Bearer ")[1]
	claims := &view.UserClaims{}
	claims, err := parseToken(token)
	if err != nil {
		return nil, false
	}
	if !validAuth(claims.UserName, claims.Password) {
		return nil, false
	}
	return claims, true
}

func validateUser(username, pwd string) bool {
	stmt := `SELECT * FROM Users WHERE userid = ? ;`
	db := config.GetDataBase()
	row := db.QueryRow(stmt, username)
	u := &view.User{}
	err := row.Scan(&u.Userid, &u.Password, &u.Max)
	if err != nil {
		return false
	}
	if !compare(pwd, []byte(u.Password)) {
		return false
	}

	return true
}

func taken(username string) bool {
	stmt := `SELECT userid FROM Users WHERE userid = ? ;`
	db := config.GetDataBase()
	row := db.QueryRow(stmt, username)
	u := &view.User{}
	err := row.Scan(&u.Userid)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

func getMax(username string) int {
	stmt := `SELECT max FROM Users WHERE userid = ? ;`
	db := config.GetDataBase()
	row := db.QueryRow(stmt, username)
	u := &view.User{}
	_ = row.Scan(&u.Max)
	return u.Max
}

func getpass(username string) string {
	stmt := `SELECT password FROM Users WHERE userid = ?;`
	db := config.GetDataBase()
	row := db.QueryRow(stmt, username)
	u := &view.User{}
	_ = row.Scan(&u.Password)
	return u.Password
}

func validAuth(id, pwd string) bool {
	stmt := `SELECT * FROM Users WHERE userid = ? AND password =?;`
	db := config.GetDataBase()
	row := db.QueryRow(stmt, id, pwd)
	u := &view.User{}
	err := row.Scan(&u.Userid, &u.Password, &u.Max)
	if err != nil {
		return false
	}
	return true
}

func checkMaxPerDay(username string) bool {
	stmt := `SELECT count(task_id) FROM Task WHERE Assigner = ? AND issue_date = ?`
	db := config.GetDataBase()
	rows, err := db.Query(stmt, username, time.Now().Format("01-02-2006"))
	if err != nil {
		return true
	}
	var created uint
	for rows.Next() {
		_ = rows.Scan(&created)
	}
	max := getMax(username)
	if int(created) >= max {
		return false
	}
	return true
}
