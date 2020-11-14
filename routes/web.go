package routes

import (
	"github.com/TODO-API/service"
	"github.com/julienschmidt/httprouter"
)

//InitRoutes ...
func InitRoutes() *httprouter.Router {
	r := httprouter.New()
	r.POST("/user/register", service.Register)
	r.GET("/user/login", service.Login)
	r.GET("/task", service.ShowAllTasks)
	r.POST("/task", service.AddTask)
	r.GET("/task/:id", service.ShowSingleTask)
	r.PUT("/task/:id", service.UpdateTask)
	r.DELETE("/task/:id", service.DeleteTask)
	r.GET("/dashboard", service.Dashboard)
	return r
}
