package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Table struct {
	Comment map[string]string
	Name    map[string][]Column
}
type Column struct {
	Field   string
	Type    string
	Comment string
}

func main() {
	tpl := "mysql-to-proto/template/proto.go.tpl"
	file := "mysql-to-proto/sso.proto"
	dbName := "yuedan_user"
	db, err := Connect("mysql", "root:123456@tcp(127.0.0.1:3306)/"+dbName+"?charset=utf8mb4&parseTime=true")
	//Table names to be excluded
	exclude := map[string]int{"user_authority_log": 1}
	if err != nil {
		fmt.Println(err)
	}
	t := Table{}
	t.TableColumn(db, dbName, exclude)
	t.Generate(file, tpl)

	fmt.Println(os.Getenv("GOPATH"))
}

//Extract table field information
func (table *Table) TableColumn(db *sql.DB, dbName string, exclude map[string]int) {
	rows, err := db.Query("SELECT t.TABLE_NAME,t.TABLE_COMMENT,c.COLUMN_NAME,c.COLUMN_TYPE,c.COLUMN_COMMENT FROM information_schema.TABLES t,INFORMATION_SCHEMA.Columns c WHERE c.TABLE_NAME=t.TABLE_NAME AND t.`TABLE_SCHEMA`='" + dbName + "'")
	defer db.Close()
	defer rows.Close()
	var name, comment string
	var column Column
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: ", err)
		return
	}
	table.Comment = make(map[string]string)
	table.Name = make(map[string][]Column)
	for rows.Next() {
		rows.Scan(&name, &comment, &column.Field, &column.Type, &column.Comment)
		if _, ok := exclude[name]; ok {
			continue
		}
		if _, ok := table.Comment[name]; !ok {
			table.Comment[name] = comment
		}
		table.Name[name] = append(table.Name[name], column)
	}

	if err = rows.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: ", err)
		return
	}
	return
}

//Generate proto files in the current directory
func (table *Table) Generate(filepath, tpl string) {
	type RpcServers struct {
		Models string
		Name   string
	}
	rpcservers := RpcServers{Models: "sso", Name: "Sso"}

	tmpl, err := template.ParseFiles(tpl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: ", err)
		return
	}
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY, 0755)

	err = tmpl.Execute(file, rpcservers)
	//err = tmpl.Execute(os.Stdout, rpcservers)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: ", err)
		return
	}
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

//Get the program run path
func GetRunDirectory() (string, error) {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return "", err
	}
	return strings.Replace(dir, "\\", "/", -1), nil
}
