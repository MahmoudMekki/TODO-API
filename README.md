# Go Todo

TODO list

## Introduction

Back-End development for TODO REST API

## Requirements
* MySQL installed
* Go installed

## Installation

* Clone this repo 

```bash
git clone https://github.com/MahmoudMekki/TODO-API.git
```

* Change Directory

```bash
cd TODO-API
```

* Modify `.env` file with your correct database credentials and desired Port

## Usage

To run this application, execute:

```bash
go run main.go
```

You should be able to access this application at `localhost:8080`

>**NOTE**<br>
>If you modified the port in the `.env` file, you should access the application for the port you set

## Usage 101
on Postman ,this API provids Methods like:
.POST -> localhost:8080/user/register -> {"username":"YOUR username","password": "YOUR password","maxtodo": number of your tasks per day}
.POST -> localhost:8080/user/login -> {"username":"YOUR username","password": "YOUR password"}
.GET -> localhost:8080/task -> responds with  all your tasks
.POST -> localhost:8080/task -> {"content":"your content","assignee":"name of the one you wanna assign this task to","duedate":"deadline for this task","state":"state of the task completed||open||overdue"}
.GET -> localhost:8080/task/taskid -> get the task of this id
.PUT ->localhost:8080/task/taskid -> to update the task
.DELETE -> localhost:8080/task/taskid -> to delete task
.GET -> localhost:8080/dashboard -> get your dashboard (completed tasks, pending tasks,over due tasks,top assigners,top assignees,top resolvers)

## Conclusion 

If you have anything to add to this, please send in a PR as it will no longer be actively maintained by [me](https://github.com/MahmoudMekki).






