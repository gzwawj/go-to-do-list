/*
 * @Description:
 * @Author: gzwawj
 * @Date: 2019-09-11 08:25:56
 */
package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/kataras/iris"
	_ "github.com/mattn/go-sqlite3"
)

var dbpath = "./task.db"
var tableName = "task"

// func newApp() *iris.Application {
// 	app := iris.New()
// 	app.StaticEmbedded("/", "./views", Asset, AssetNames)
// 	return app
// }

func main() {

	app := iris.New()
	app.OnErrorCode(iris.StatusNotFound, ontFound)
	app.OnErrorCode(iris.StatusInternalServerError, internalServerError)

	tmp := iris.HTML("./views", ".html")
	// tmp.Binary(Asset, AssetNames) //新添加这一行代码

	app.RegisterView(tmp)

	app.Get("/", getAll("index"))

	app.Get("/task", getAll("json"))
	app.Get("/task/:id", getOne())
	app.Post("/task/create", create())
	app.Post("/task/editor/:id", editor())
	app.Post("/task/delete/:id", delete())
	app.Run(iris.Addr(":8080"))

}

//连接数据库
func connection(path string) *sql.DB {
	_, err := os.Stat(path)
	if err != nil {
		fmt.Println("没有文件，进行创建")
		file, err := os.Create(path)
		if err != nil {
			fmt.Println(err)
		}
		defer file.Close()
	}
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		panic(err)
	}
	createTable(db)
	return db
}

//查询数据库表字段
func selectTable() {
	db := connection(dbpath)
	sqlstr := fmt.Sprintf("select * from sqlite_master where type='table' and name='%s'", tableName)
	rows, err := db.Query(sqlstr)
	checkErr(err)
	defer db.Close()
	defer rows.Close()
	fmt.Println(rows)
}

//创建数据库表
func createTable(db *sql.DB) {

	sqlstr :=
		`
		CREATE TABLE "main"."task" (
			"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
			"content" TEXT NOT NULL,
			"level" integer(3),
			"progress" TEXT,
			"editortime" integer(11) NOT NULL
		  );
		`

	db.Exec(sqlstr)
	// defer db.Close()
}

//修改数据库表
func alterTable(field string, types string) {
	db := connection(dbpath)
	fields := fmt.Sprintf("%s %s", field, types)

	default_fields := fmt.Sprintf("%s", fields)

	sqlstr := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", tableName, default_fields)

	db.Exec(sqlstr)
	defer db.Close()
}

//获取所有数据
func getAll(str string) iris.Handler {
	return func(ctx iris.Context) {
		db := connection(dbpath)
		sqlstr := fmt.Sprintf("SELECT * FROM %s", tableName)
		rows, err := db.Query(sqlstr)
		checkErr(err)
		defer db.Close()
		defer rows.Close()
		var task []Task
		for rows.Next() {
			task = append(task, randTask(rows))
		}
		if str == "json" {
			ctx.JSON(iris.Map{"code": 200, "msg": "test", "data": task})
		}
		if str == "index" {
			ctx.ViewData("message", task)
			ctx.View("todolist.html")
		}
	}
}
func getOne() iris.Handler {
	return func(ctx iris.Context) {
		id := ctx.Params().Get("id")
		db := connection(dbpath)
		sqlstr := fmt.Sprintf("SELECT * FROM %s WHERE id = %s ", tableName, id)
		rows, err := db.Query(sqlstr)
		checkErr(err)
		defer rows.Close()
		var task []Task
		for rows.Next() {
			task = append(task, randTask(rows))
		}
		ctx.JSON(iris.Map{"code": 2001, "msg": "请求成功", "data": task})
	}
}
func create() iris.Handler {
	return func(ctx iris.Context) {
		db := connection(dbpath)

		col := fmt.Sprintf("%s, %s, %s, %s", "content", "level", "progress", "editortime")
		val := fmt.Sprintf("'%s','%s','%s','%d'", ctx.PostValue("content"), ctx.PostValue("level"), ctx.PostValue("progress"), time.Now().Unix())
		sqlstr := fmt.Sprintf("INSERT INTO  %s (%s) VALUES(%s)", tableName, col, val)
		res, err := db.Exec(sqlstr)
		defer db.Close()

		if err != nil {
			//panic(err.Error())
			ctx.JSON(iris.Map{"code": 2002, "msg": err.Error(), "data": ""})
		} else {
			lastId, err := res.LastInsertId()
			if err != nil {
				log.Fatal(err)
			}

			ctx.JSON(iris.Map{"code": 2001, "msg": "添加成功", "data": lastId})
		}
	}
}
func delete() iris.Handler {
	return func(ctx iris.Context) {
		id := ctx.Params().Get("id")
		db := connection(dbpath)
		sqlstr := fmt.Sprintf("DELETE FROM %s WHERE id = %s", tableName, id)
		res, err := db.Exec(sqlstr)
		defer db.Close()
		if err != nil {
			//panic(err.Error())
			ctx.JSON(iris.Map{"code": 2002, "msg": err.Error(), "data": ""})
		} else {
			rowCnt, err := res.RowsAffected()
			if err != nil {
				log.Fatal(err)
			}
			if rowCnt == 0 {
				ctx.JSON(iris.Map{"code": 2002, "msg": "删除失败", "data": ""})
			} else {
				ctx.JSON(iris.Map{"code": 2001, "msg": "删除成功", "data": ""})
			}

		}
	}
}
func editor() iris.Handler {
	return func(ctx iris.Context) {
		db := connection(dbpath)

		id := ctx.Params().Get("id")

		val := fmt.Sprintf("%s='%s',%s='%s',%s='%s',%s='%d'", "content", ctx.PostValue("content"), "level", ctx.PostValue("level"), "progress", ctx.PostValue("progress"), "editortime", time.Now().Unix())
		sqlstr := fmt.Sprintf("UPDATE %s SET %s WHERE id=%s", tableName, val, id)
		res, err := db.Exec(sqlstr)
		defer db.Close()

		if err != nil {
			//panic(err.Error())
			ctx.JSON(iris.Map{"code": 2002, "msg": err.Error(), "data": ""})
		} else {
			rowCnt, err := res.RowsAffected()
			if err != nil {
				log.Fatal(err)
			}
			if rowCnt == 0 {
				ctx.JSON(iris.Map{"code": 2001, "msg": "内容无变化", "data": ""})
			} else {
				ctx.JSON(iris.Map{"code": 2001, "msg": "更新成功", "data": ""})
			}
		}
	}
}

//错误验证
func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
func ontFound(ctx iris.Context) {
	ctx.HTML("访问错误")
}
func internalServerError(ctx iris.Context) {
	ctx.WriteString("Oups something went wrong, try again")
}

type Task struct {
	Id         int
	Content    string
	Level      string
	Progress   string
	Editortime string
}

//数据渲染
func randTask(rows *sql.Rows) Task {
	var (
		id         int
		content    string
		level      string
		progress   string
		editortime string
	)
	err := rows.Scan(&id, &content, &level, &progress, &editortime)
	checkErr(err)
	return Task{id, content, level, progress, editortime}
}
