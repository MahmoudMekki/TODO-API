package service

import (
	"encoding/json"
	"net/http"

	storage "github.com/TODO-API/db_config"

	"github.com/julienschmidt/httprouter"
)

var todo = storage.Store{
	DB:     storage.New(),
	JWTKey: []byte("MahmoudMekki"),
}

//Register ...
func Register(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {

	err := todo.Signup(req)

	if err != nil {
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusNotAcceptable)
		json.NewEncoder(res).Encode(map[string]string{"error": err.Error()})
		return
	}
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(map[string]string{"status": "user created successfully"})
}

//Login ...
func Login(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	token, err := todo.Login(req)
	if err != nil {
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusNonAuthoritativeInfo)
		json.NewEncoder(res).Encode(map[string]string{"status": err.Error()})
		return
	}
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(map[string]string{"status": "you are signed in", "token": token})
}

// ShowAllTasks ...
func ShowAllTasks(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	tasks, err := todo.AllTasks(req)
	if err != nil {
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusNotAcceptable)
		json.NewEncoder(res).Encode(map[string]string{"error": err.Error()})
		return
	}
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(tasks)

}

//AddTask ...
func AddTask(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	err := todo.AddTask(req)
	if err != nil {
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusNotAcceptable)
		json.NewEncoder(res).Encode(map[string]string{"error": err.Error()})
		return
	}
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusAccepted)
	json.NewEncoder(res).Encode(map[string]string{"state": "Your task has been added succesfully!"})
}

// ShowSingleTask ...
func ShowSingleTask(res http.ResponseWriter, req *http.Request, p httprouter.Params) {
	taskid := p.ByName("id")
	task, err := todo.TaskByID(req, taskid)
	if err != nil {
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusNotAcceptable)
		json.NewEncoder(res).Encode(map[string]string{"error": err.Error()})
		return
	}
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(task)

}

// UpdateTask ...
func UpdateTask(res http.ResponseWriter, req *http.Request, p httprouter.Params) {
	taskid := p.ByName("id")
	err := todo.UpdateTask(req, taskid)
	if err != nil {
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusNotAcceptable)
		json.NewEncoder(res).Encode(map[string]string{"error": err.Error()})
		return
	}
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(map[string]string{"state": "Task updated successfully!"})
}

//DeleteTask ...
func DeleteTask(res http.ResponseWriter, req *http.Request, p httprouter.Params) {
	taskid := p.ByName("id")
	err := todo.DeleteTask(req, taskid)
	if err != nil {
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusNotAcceptable)
		json.NewEncoder(res).Encode(map[string]string{"error": err.Error()})
		return
	}
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(map[string]string{"state": "Task Deleted successfully!"})
}

//Dashboard ...
func Dashboard(res http.ResponseWriter, req *http.Request, p httprouter.Params) {

	dashboard, err := todo.DashBoard(req)
	if err != nil {
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusNotAcceptable)
		json.NewEncoder(res).Encode(map[string]string{"error": err.Error()})
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(dashboard)

}
