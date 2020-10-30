package view

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

//UserClaims for JWT token ....
type UserClaims struct {
	jwt.StandardClaims
	UserName string
	Password string
	MaxTODO  int
}

//Toppers struct ...
type Toppers struct {
	Name   string `json:"top_ranked"`
	Amount int    `json:"no_attembts"`
}

//User struct ...
type User struct {
	Userid   string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Max      int    `json:"maxtodo,omitempty"`
}

//Task struct...
type Task struct {
	TaskID    int    `json:"taskid,omitempty"`
	Assigner  string `json:"assigner"`
	Content   string `json:"content"`
	IssueDate string `json:"issuedate,omitempty"`
	DueDate   string `json:"duedate,omitempty"`
	State     string `json:"state"`
	Assignee  string `json:"assignee"`
}

//Dashboard struct...
type Dashboard struct {
	Completed []Task    `json:"completed_tasks"`
	Pending   []Task    `json:"penging_tasks"`
	OverDue   []Task    `json:"overdue_tasks"`
	Resolvers []Toppers `json:"Top_resolvers"`
	Assigners []Toppers `json:"Top_assigners"`
	Assignee  []Toppers `json:"Top_assignees"`
}

//Valid the token
func (u *UserClaims) Valid() error {
	if !u.VerifyExpiresAt(time.Now().Unix(), false) {
		return fmt.Errorf("Token is timed out")
	}

	return nil
}
