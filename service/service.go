package service

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/TODO-API/view"

	"github.com/julienschmidt/httprouter"

	config "github.com/TODO-API/db_config"
)

var db = config.Database()

//Register ...
func Register(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	u := view.User{}
	json.NewDecoder(req.Body).Decode(&u)
	if taken(u.Userid) {
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusNotAcceptable)
		json.NewEncoder(res).Encode(map[string]string{"error": "user name is already taken"})
		return
	}
	hash, _ := hashPassword(u.Password)
	u.Password = string(hash)
	stm, _ := db.Prepare(`INSERT INTO Users VALUES (?,?,?);`)
	_, _ = stm.Exec(u.Userid, u.Password, u.Max)
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(map[string]string{"status": "user created successfully"})
}

//Login ...
func Login(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	u := view.User{}
	json.NewDecoder(req.Body).Decode(&u)
	if !validateUser(u.Userid, u.Password) {
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusNonAuthoritativeInfo)
		json.NewEncoder(res).Encode(map[string]string{"status": "Wrong username of password"})
		return
	}
	claims := view.UserClaims{
		UserName: u.Userid,
		Password: getpass(u.Userid),
		MaxTODO:  getMax(u.Userid),
	}
	token, _ := createToken(&claims)
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(map[string]string{"status": "you are signed in", "token": token})
}

// ShowAllTasks ...
func ShowAllTasks(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	claims, ok := checkToken(req)
	if !ok {
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusNotAcceptable)
		json.NewEncoder(res).Encode(map[string]string{"error": "Authorization missing"})
		return
	}
	tasks := []view.Task{}
	rows, _ := db.Query("SELECT * FROM Task WHERE assigner =?;", claims.UserName)
	for rows.Next() {
		task := view.Task{}
		err := rows.Scan(&task.TaskID, &task.Assigner, &task.Content, &task.State, &task.Assignee, &task.IssueDate, &task.DueDate)
		if err != nil {
			res.Header().Set("Content-Type", "application/json")
			res.WriteHeader(http.StatusNoContent)
			json.NewEncoder(res).Encode(map[string]string{"error": "You dont have any tasks to show!"})
			return
		}
		tasks = append(tasks, task)
	}
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(tasks)

}

//AddTask ...
func AddTask(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	claims, ok := checkToken(req)
	if !ok {
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusNotAcceptable)
		json.NewEncoder(res).Encode(map[string]string{"error": "Authorization missing"})
		return
	}
	if !checkMaxPerDay(claims.UserName) {
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusNotAcceptable)
		json.NewEncoder(res).Encode(map[string]string{"error": "You reached your limit today, you can start add more tomorrow!"})
		return
	}
	task := view.Task{}
	json.NewDecoder(req.Body).Decode(&task)
	task.Assigner = claims.UserName
	task.IssueDate = time.Now().Format("01-02-2006")
	stmt, _ := db.Prepare("INSERT INTO Task (assigner,content,state,assignee,issue_date,due_date) VALUES (?,?,?,?,?,?);")
	stmt.Exec(task.Assigner, task.Content, task.State, task.Assignee, task.IssueDate, task.DueDate)
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusAccepted)
	json.NewEncoder(res).Encode(map[string]string{"state": "Your task has been added succesfully!"})
}

// ShowSingleTask ...
func ShowSingleTask(res http.ResponseWriter, req *http.Request, p httprouter.Params) {
	claims, ok := checkToken(req)
	if !ok {
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusNotAcceptable)
		json.NewEncoder(res).Encode(map[string]string{"error": "Authorization missing"})
		return
	}
	taskid := p.ByName("id")
	task := view.Task{}
	row := db.QueryRow("SELECT * FROM Task WHERE assigner =? AND task_id =?;", claims.UserName, taskid)
	row.Scan(&task.TaskID, &task.Assigner, &task.Content, &task.State, &task.Assignee, &task.IssueDate, &task.DueDate)
	if task.TaskID <= 0 {
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusNotAcceptable)
		json.NewEncoder(res).Encode(map[string]string{"error": "No task with this id!"})
		return
	}
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(task)

}

// UpdateTask ...
func UpdateTask(res http.ResponseWriter, req *http.Request, p httprouter.Params) {
	claims, ok := checkToken(req)
	if !ok {
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusNotAcceptable)
		json.NewEncoder(res).Encode(map[string]string{"error": "Authorization missing"})
		return
	}
	taskid := p.ByName("id")
	task := view.Task{}
	json.NewDecoder(req.Body).Decode(&task)
	task.Assigner = claims.UserName
	task.IssueDate = time.Now().Format("01-02-2006")
	stmt, _ := db.Prepare("UPDATE Task SET assigner=?,assignee=?,content=?,issue_date=?,due_date=?,state=? WHERE task_id=?;")
	n, _ := stmt.Exec(&task.Assigner, &task.Assignee, &task.Content, &task.IssueDate, &task.DueDate, &task.State, taskid)
	r, _ := n.RowsAffected()
	if r <= 0 {
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusNotAcceptable)
		json.NewEncoder(res).Encode(map[string]string{"error": "No task with this id!"})
		return
	}
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(map[string]string{"state": "Task updated successfully!"})
}

//DeleteTask ...
func DeleteTask(res http.ResponseWriter, req *http.Request, p httprouter.Params) {
	claims, ok := checkToken(req)
	if !ok {
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusNotAcceptable)
		json.NewEncoder(res).Encode(map[string]string{"error": "Authorization missing"})
		return
	}
	taskid := p.ByName("id")
	r, _ := db.Exec("DELETE FROM Task WHERE task_id=? AND assigner=?;", taskid, claims.UserName)
	n, _ := r.RowsAffected()
	if n <= 0 {
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusNotAcceptable)
		json.NewEncoder(res).Encode(map[string]string{"error": "No task with this id!"})
		return
	}
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(map[string]string{"state": "Task Deleted successfully!"})
}

//Dashboard ...
func Dashboard(res http.ResponseWriter, req *http.Request, p httprouter.Params) {

	claims, ok := checkToken(req)
	if !ok {
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusNotAcceptable)
		json.NewEncoder(res).Encode(map[string]string{"error": "Authorization missing"})
		return
	}
	dashboard := view.Dashboard{}
	rows, _ := db.Query("SELECT * FROM Task WHERE assigner=? AND state = ?;", claims.UserName, "completed")
	for rows.Next() {
		task := view.Task{}
		rows.Scan(&task.TaskID, &task.Assigner, &task.Content, &task.State, &task.Assignee, &task.IssueDate, &task.DueDate)
		dashboard.Completed = append(dashboard.Completed, task)
	}

	rows, _ = db.Query("SELECT * FROM Task WHERE assigner=? AND NOT state = ? AND NOT state=?;", claims.UserName, "overdue", "completed")
	for rows.Next() {
		task := view.Task{}
		rows.Scan(&task.TaskID, &task.Assigner, &task.Content, &task.State, &task.Assignee, &task.IssueDate, &task.DueDate)
		dashboard.Pending = append(dashboard.Pending, task)
	}
	rows, _ = db.Query("SELECT * FROM Task WHERE assigner=? AND state = ?;", claims.UserName, "overdue")
	for rows.Next() {
		task := view.Task{}
		rows.Scan(&task.TaskID, &task.Assigner, &task.Content, &task.State, &task.Assignee, &task.IssueDate, &task.DueDate)
		dashboard.OverDue = append(dashboard.OverDue, task)
	}

	rows, _ = db.Query("SELECT assigner,COUNT(assigner) AS amount FROM Task GROUP BY assigner ORDER BY amount DESC LIMIT 2;")
	for rows.Next() {
		assigner := view.Toppers{}
		rows.Scan(&assigner.Name, &assigner.Amount)
		dashboard.Assigners = append(dashboard.Assigners, assigner)
	}
	rows, _ = db.Query("SELECT assignee,COUNT(assignee) AS amount FROM Task GROUP BY assignee ORDER BY amount DESC LIMIT 2;")
	for rows.Next() {
		assignee := view.Toppers{}
		rows.Scan(&assignee.Name, &assignee.Amount)
		dashboard.Assignee = append(dashboard.Assignee, assignee)
	}

	rows, _ = db.Query("SELECT assigner,assignee,COUNT(state) AS amount FROM Task WHERE state = ? GROUP BY assigner,assignee ORDER BY amount DESC LIMIT 2 ;", "completed")
	for rows.Next() {
		resolver := view.Toppers{}
		rows.Scan(&resolver.Name, &resolver.Amount)
		dashboard.Resolvers = append(dashboard.Resolvers, resolver)
	}
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(dashboard)

}
