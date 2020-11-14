package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/TODO-API/view"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

//Store structer
type Store struct {
	DB     *sql.DB
	JWTKey []byte
}

//Signup ...
func (s *Store) Signup(req *http.Request) error {
	u := view.User{}
	json.NewDecoder(req.Body).Decode(&u)
	if s.taken(u.Userid) {
		return errors.New("Username is already existed ")
	}
	hash, _ := hashPassword(u.Password)
	u.Password = string(hash)
	stm, _ := s.DB.Prepare(`INSERT INTO Users VALUES (?,?,?);`)
	_, _ = stm.Exec(u.Userid, u.Password, u.Max)
	return nil
}

//Login ...
func (s *Store) Login(req *http.Request) (string, error) {
	u := view.User{}
	json.NewDecoder(req.Body).Decode(&u)
	if !s.validateUser(u.Userid, u.Password) {
		return "", errors.New("Wrong Username or password, try again")
	}
	claims := view.UserClaims{
		UserName: u.Userid,
		Password: s.getpass(u.Userid),
		MaxTODO:  s.getMax(u.Userid),
	}
	token, _ := s.createToken(&claims)
	return token, nil
}

//AllTasks ...
func (s *Store) AllTasks(req *http.Request) ([]view.Task, error) {
	claims, ok := s.checkToken(req)
	if !ok {
		return nil, errors.New("UnAuthorized, Login or register first then try again")
	}
	tasks := []view.Task{}
	rows, _ := s.DB.Query("SELECT * FROM Task WHERE assigner =?;", claims.UserName)
	for rows.Next() {
		task := view.Task{}
		err := rows.Scan(&task.TaskID, &task.Assigner, &task.Content, &task.State, &task.Assignee, &task.IssueDate, &task.DueDate)
		if err != nil {
			return nil, errors.New("No Tasks to show up")
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

//AddTask ...
func (s *Store) AddTask(req *http.Request) error {
	claims, ok := s.checkToken(req)
	if !ok {
		return errors.New("UnAuthorized, Login or register first then try again")
	}
	if !s.checkMaxPerDay(claims.UserName) {
		return errors.New("You reached your maximum tasks per day")
	}
	task := view.Task{}
	json.NewDecoder(req.Body).Decode(&task)
	task.Assigner = claims.UserName
	task.IssueDate = time.Now().Format("01-02-2006")
	stmt, _ := s.DB.Prepare("INSERT INTO Task (assigner,content,state,assignee,issue_date,due_date) VALUES (?,?,?,?,?,?);")
	stmt.Exec(task.Assigner, task.Content, task.State, task.Assignee, task.IssueDate, task.DueDate)
	return nil
}

//TaskByID ...
func (s *Store) TaskByID(req *http.Request, id string) (view.Task, error) {
	task := view.Task{}
	claims, ok := s.checkToken(req)
	if !ok {
		return task, errors.New("UnAuthorized, Login or register first then try again")
	}

	row := s.DB.QueryRow("SELECT * FROM Task WHERE assigner =? AND task_id =?;", claims.UserName, id)
	row.Scan(&task.TaskID, &task.Assigner, &task.Content, &task.State, &task.Assignee, &task.IssueDate, &task.DueDate)
	if task.TaskID <= 0 {
		return task, errors.New("No Tasks with this ID")
	}
	return task, nil
}

//UpdateTask ...
func (s *Store) UpdateTask(req *http.Request, id string) error {
	claims, ok := s.checkToken(req)
	if !ok {
		return errors.New("UnAuthorized, Login or register first then try again")
	}
	task := view.Task{}
	json.NewDecoder(req.Body).Decode(&task)
	task.Assigner = claims.UserName
	task.IssueDate = time.Now().Format("01-02-2006")
	stmt, _ := s.DB.Prepare("UPDATE Task SET assigner=?,assignee=?,content=?,issue_date=?,due_date=?,state=? WHERE task_id=?;")
	n, _ := stmt.Exec(&task.Assigner, &task.Assignee, &task.Content, &task.IssueDate, &task.DueDate, &task.State, id)
	r, _ := n.RowsAffected()
	if r <= 0 {
		return errors.New("No Task with this ID")
	}
	return nil
}

//DeleteTask ...
func (s *Store) DeleteTask(req *http.Request, id string) error {
	claims, ok := s.checkToken(req)
	if !ok {
		return errors.New("UnAuthorized, Login or register first then try again")
	}
	r, _ := s.DB.Exec("DELETE FROM Task WHERE task_id=? AND assigner=?;", id, claims.UserName)
	n, _ := r.RowsAffected()
	if n <= 0 {
		return errors.New("No task with this ID ")
	}
	return nil
}

//DashBoard ...
func (s *Store) DashBoard(req *http.Request) (view.Dashboard, error) {
	dashboard := view.Dashboard{}
	claims, ok := s.checkToken(req)
	if !ok {
		return dashboard, errors.New("UnAuthorized, Login or register first then try again")
	}

	rows, _ := s.DB.Query("SELECT * FROM Task WHERE assigner=? AND state = ?;", claims.UserName, "completed")
	for rows.Next() {
		task := view.Task{}
		rows.Scan(&task.TaskID, &task.Assigner, &task.Content, &task.State, &task.Assignee, &task.IssueDate, &task.DueDate)
		dashboard.Completed = append(dashboard.Completed, task)
	}

	rows, _ = s.DB.Query("SELECT * FROM Task WHERE assigner=? AND NOT state = ? AND NOT state=?;", claims.UserName, "overdue", "completed")
	for rows.Next() {
		task := view.Task{}
		rows.Scan(&task.TaskID, &task.Assigner, &task.Content, &task.State, &task.Assignee, &task.IssueDate, &task.DueDate)
		dashboard.Pending = append(dashboard.Pending, task)
	}
	rows, _ = s.DB.Query("SELECT * FROM Task WHERE assigner=? AND state = ?;", claims.UserName, "overdue")
	for rows.Next() {
		task := view.Task{}
		rows.Scan(&task.TaskID, &task.Assigner, &task.Content, &task.State, &task.Assignee, &task.IssueDate, &task.DueDate)
		dashboard.OverDue = append(dashboard.OverDue, task)
	}

	rows, _ = s.DB.Query("SELECT assigner,COUNT(assigner) AS amount FROM Task GROUP BY assigner ORDER BY amount DESC LIMIT 2;")
	for rows.Next() {
		assigner := view.Toppers{}
		rows.Scan(&assigner.Name, &assigner.Amount)
		dashboard.Assigners = append(dashboard.Assigners, assigner)
	}
	rows, _ = s.DB.Query("SELECT assignee,COUNT(assignee) AS amount FROM Task GROUP BY assignee ORDER BY amount DESC LIMIT 2;")
	for rows.Next() {
		assignee := view.Toppers{}
		rows.Scan(&assignee.Name, &assignee.Amount)
		dashboard.Assignee = append(dashboard.Assignee, assignee)
	}

	rows, _ = s.DB.Query("SELECT assigner,assignee,COUNT(state) AS amount FROM Task WHERE state = ? GROUP BY assigner,assignee ORDER BY amount DESC LIMIT 2 ;", "completed")
	for rows.Next() {
		resolver := view.Toppers{}
		rows.Scan(&resolver.Name, &resolver.Amount)
		dashboard.Resolvers = append(dashboard.Resolvers, resolver)
	}
	return dashboard, nil
}

func hashPassword(password string) ([]byte, error) {
	bs, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("error with converting to hash %w", err)
	}
	return bs, nil
}

func (s *Store) createToken(c *view.UserClaims) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS512, c)
	token, err := t.SignedString(s.JWTKey)
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
func (s *Store) parseToken(token string) (*view.UserClaims, error) {
	claims := &view.UserClaims{}
	t, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		if t.Method.Alg() != jwt.SigningMethodHS512.Alg() {
			return nil, fmt.Errorf("Invalid signing algorithm")
		}
		return s.JWTKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("Error in parsing the token")
	}
	if !t.Valid {
		return nil, fmt.Errorf("Invalid token")
	}
	return t.Claims.(*view.UserClaims), nil
}

func (s *Store) checkToken(req *http.Request) (*view.UserClaims, bool) {
	token := req.Header.Get("Authorization")
	token = strings.Split(token, "Bearer ")[1]
	claims := &view.UserClaims{}
	claims, err := s.parseToken(token)
	if err != nil {
		return nil, false
	}
	if !s.validAuth(claims.UserName, claims.Password) {
		return nil, false
	}
	return claims, true
}

func (s *Store) validateUser(username, pwd string) bool {
	stmt := `SELECT * FROM Users WHERE userid = ? ;`
	row := s.DB.QueryRow(stmt, username)
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

func (s *Store) taken(username string) bool {
	stmt := `SELECT userid FROM Users WHERE userid = ? ;`
	row := s.DB.QueryRow(stmt, username)
	u := &view.User{}
	err := row.Scan(&u.Userid)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

func (s *Store) getMax(username string) int {
	stmt := `SELECT max FROM Users WHERE userid = ? ;`
	row := s.DB.QueryRow(stmt, username)
	u := &view.User{}
	_ = row.Scan(&u.Max)
	return u.Max
}

func (s *Store) getpass(username string) string {
	stmt := `SELECT password FROM Users WHERE userid = ?;`
	row := s.DB.QueryRow(stmt, username)
	u := &view.User{}
	_ = row.Scan(&u.Password)
	return u.Password
}

func (s *Store) validAuth(id, pwd string) bool {
	stmt := `SELECT * FROM Users WHERE userid = ? AND password =?;`
	row := s.DB.QueryRow(stmt, id, pwd)
	u := &view.User{}
	err := row.Scan(&u.Userid, &u.Password, &u.Max)
	if err != nil {
		return false
	}
	return true
}

func (s *Store) checkMaxPerDay(username string) bool {
	stmt := `SELECT count(task_id) FROM Task WHERE Assigner = ? AND issue_date = ?`
	rows, err := s.DB.Query(stmt, username, time.Now().Format("01-02-2006"))
	if err != nil {
		return true
	}
	var created uint
	for rows.Next() {
		_ = rows.Scan(&created)
	}
	max := s.getMax(username)
	if int(created) >= max {
		return false
	}
	return true
}
