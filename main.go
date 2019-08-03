package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"html/template"
	"log"
	"os"
	"strings"
)

var typeArr = map[string]string{
	"int":       "int32",
	"tinyint":   "int32",
	"smallint":  "int32",
	"mediumint": "int32",
	"enum":      "int32",
	"bigint":    "sint64",
	"varchar":   "string",
	"timestamp": "string",
	"date":      "string",
	"text":      "string",
	"double":    "double",
	"decimal":   "double",
	"float":     "float",
}

type Table struct {
	PackageModels string
	ServiceName   string
	Method        map[string]MethodDetail
	Comment       map[string]string
	Name          map[string][]Column
	Message       map[string]Detail
}
type MethodDetail struct {
	Request  Detail
	Response Detail
}
type Column struct {
	Field   string
	Type    string
	Comment string
}
type RpcServers struct {
	Models      string
	Name        string
	Funcs       []FuncParam
	MessageList []Message
}
type FuncParam struct {
	Name         string
	RequestName  string
	ResponseName string
}
type Message struct {
	Name          string
	MessageDetail []Field
}
type Field struct {
	TypeName string
	AttrName string
	Num      int
}

type Detail struct {
	Name string
	Cat  string
	Attr []AttrDetail
}

type AttrDetail struct {
	TypeName string
	AttrName string
}

func main() {
	tpl := "d:/gopath/src/mysql-to-proto/template/proto.go.tpl"
	file := "d:/gopath/src/mysql-to-proto/sso.proto"
	dbName := "yuedan_user"
	db, err := Connect("mysql", "root:123456@tcp(127.0.0.1:3306)/"+dbName+"?charset=utf8mb4&parseTime=true")
	//Table names to be excluded
	exclude := map[string]int{"user_authority_log": 1}
	if err != nil {
		fmt.Println(err)
	}
	if IsFile(file) {
		fmt.Fprintf(os.Stderr, "Fatal error: ", "file already exist")
		return
	}
	t := Table{}
	t.Message = map[string]Detail{
		"Filter": {
			Name: "Filter",
			Cat:  "custom",
			Attr: []AttrDetail{{
				TypeName: "uint64",
				AttrName: "id",
			}},
		},
		"Request": {
			Name: "Request",
			Cat:  "all",
		},
		"Response": {
			Name: "Response",
			Cat:  "custom",
			Attr: []AttrDetail{
				{
					TypeName: "uint64",
					AttrName: "id",
				},
				{
					TypeName: "bool",
					AttrName: "success",
				},
			},
		},
	}

	t.PackageModels = "sso"
	t.ServiceName = "Sso"
	t.Method = map[string]MethodDetail{
		"Get":    {Request: t.Message["Filter"], Response: t.Message["Request"]},
		"Create": {Request: t.Message["Request"], Response: t.Message["Response"]},
		"Update": {Request: t.Message["Request"], Response: t.Message["Response"]},
	}
	t.TableColumn(db, dbName, exclude)
	t.Generate(file, tpl)
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

		n := strings.Index(column.Type, "(")
		if n > 0 {
			column.Type = column.Type[0:n]
		}
		n = strings.Index(column.Type, " ")
		if n > 0 {
			column.Type = column.Type[0:n]
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
	rpcservers := RpcServers{Models: table.PackageModels, Name: table.ServiceName}
	rpcservers.HandleFuncs(table)
	rpcservers.HandleMessage(table)
	tmpl, err := template.ParseFiles(tpl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: ", err)
		return
	}
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY, 0755)
	err = tmpl.Execute(file, rpcservers)
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
	db.SetMaxIdleConns(0)
	db.SetMaxOpenConns(30)
	if err := db.Ping(); err != nil {
		log.Fatalln(err)
	}
	return db, err
}

func (r *RpcServers) HandleFuncs(t *Table) {
	var funcParam FuncParam
	for key, _ := range t.Comment {
		k := StrFirstToUpper(key)
		for n, m := range t.Method {
			funcParam.Name = n + k
			funcParam.ResponseName = k + m.Response.Name
			funcParam.RequestName = k + m.Request.Name
			r.Funcs = append(r.Funcs, funcParam)
		}
	}
}
func (r *RpcServers) HandleMessage(t *Table) {
	message := Message{}
	field := Field{}
	var i int

	for key, value := range t.Name {
		k := StrFirstToUpper(key)

		for name, detail := range t.Message {
			message.Name = k + name
			message.MessageDetail = nil
			if detail.Cat == "all" {
				i = 1
				for _, f := range value {
					field.AttrName = f.Field
					field.TypeName = TypeMToP(f.Type)
					field.Num = i
					message.MessageDetail = append(message.MessageDetail, field)
					i++
				}
			} else if detail.Cat == "custom" {
				i = 1
				for _, f := range detail.Attr {
					field.TypeName = f.TypeName
					field.AttrName = f.AttrName
					field.Num = i
					message.MessageDetail = append(message.MessageDetail, field)
					i++
				}
			}
			r.MessageList = append(r.MessageList, message)
		}

	}

}
func TypeMToP(m string) string {
	if _, ok := typeArr[m]; ok {
		return typeArr[m]
	}
	return "string"
}
func StrFirstToUpper(str string) string {
	temp := strings.Split(str, "_")
	var upperStr string
	for _, v := range temp {
		if len(v) > 0 {
			runes := []rune(v)
			upperStr += string(runes[0]-32) + string(runes[1:])
		}
	}
	return upperStr
}
func IsFile(f string) bool {
	fi, e := os.Stat(f)
	if e != nil {
		return false
	}
	return !fi.IsDir()
}
