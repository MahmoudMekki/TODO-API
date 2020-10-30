package config

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/ichtrojan/thoth"
	_ "github.com/joho/godotenv/autoload"
)

var DB *sql.DB

/*Database to connect to our database*/
func Database() *sql.DB {
	logger, _ := thoth.Init("log")

	user, exist := os.LookupEnv("DB_USER")

	if !exist {
		logger.Log(errors.New("DB_USER not set in .env"))
		log.Fatal("DB_USER not set in .env")
	}

	pass, exist := os.LookupEnv("DB_PASS")

	if !exist {
		logger.Log(errors.New("DB_PASS not set in .env"))
		log.Fatal("DB_PASS not set in .env")
	}

	host, exist := os.LookupEnv("DB_HOST")

	if !exist {
		logger.Log(errors.New("DB_HOST not set in .env"))
		log.Fatal("DB_HOST not set in .env")
	}

	credentials := fmt.Sprintf("%s:%s@tcp(%s:3306)/?charset=utf8&parseTime=True", user, pass, host)

	database, err := sql.Open("mysql", credentials)

	if err != nil {
		logger.Log(err)
		log.Fatal(err)
	} else {
		fmt.Println("Database Connection Successful")
	}
	/*	_, err = database.Exec(`DROP DATABASE TODO;`)

		if err != nil {
			fmt.Println(err)
		}
	*/
	_, err = database.Exec(`CREATE DATABASE TODO;`)

	if err != nil {
		fmt.Println(err)
	}

	_, err = database.Exec(`USE TODO;`)

	if err != nil {
		fmt.Println(err)
	}

	_, err = database.Exec(`
	CREATE TABLE Users (
		userid varchar(50) NOT NULL,
		password varchar(100) NOT NULL,
		max int NOT NULL DEFAULT '5',
		PRIMARY KEY (userid)
	  );
	  `)

	if err != nil {
		fmt.Println(err)
	}

	_, err = database.Exec(`
	CREATE TABLE Task (
		task_id int NOT NULL AUTO_INCREMENT,
		assigner varchar(45) NOT NULL,
		content varchar(45) NOT NULL,
		state varchar(40) NOT NULL DEFAULT 'open',
		assignee varchar(45) NOT NULL,
		issue_date varchar(45) NOT NULL,
		due_date varchar(45) NOT NULL,
		PRIMARY KEY (task_id),
		KEY Task_ibfk_1 (assigner),
		CONSTRAINT Task_ibfk_1 FOREIGN KEY (assigner) REFERENCES Users (userid)
	);
	  `)

	if err != nil {
		fmt.Println(err)
	}
	DB = database
	return DB
}

func GetDataBase() *sql.DB {
	return DB
}
