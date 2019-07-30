package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
)

type Table struct {
	Name []string
}
type Column struct {
	Field string
	Type  string
}

func main() {
	dbName := "yuedan_user"
	db, err := Connect("mysql", "root:123456@tcp(127.0.0.1:3306)/"+dbName+"?charset=utf8mb4&parseTime=true")
	if err != nil {
		fmt.Println(err)
	}
	TableColumn(db, dbName)

	//fmt.Println(db, err)

}

func TableColumn(db *sql.DB, dbName string) {
	rows, err := db.Query("SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = '" + dbName + "'")
	defer db.Close()
	defer rows.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: ", err)
		return
	}
	for rows.Next() {
		var name string
		rows.Scan(&name)
		rowsTable, err := db.Query("SHOW COLUMNS FROM " + name + "")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Fatal error: ", err)
			return
		}
		for rows.Next() {

		}

	}
	if err = rows.Err(); err != nil {
		return
	}
	return
}

func Connect(driverName, dsn string) (*sql.DB, error) {
	db, err := sql.Open(driverName, dsn)

	if err != nil {
		log.Fatalln(err)
	}
	//用于设置闲置的连接数。如果 <= 0, 则没有空闲连接会被保留
	db.SetMaxIdleConns(0)
	//用于设置最大打开的连接数,默认值为0表示不限制。
	db.SetMaxOpenConns(30)
	if err := db.Ping(); err != nil {
		log.Fatalln(err)
	}
	return db, err
}
