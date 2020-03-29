package web

import (
	"database/sql"
	"fmt"
	"time"

	"webtest/config"
	//
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func init() {
	db, _ = sql.Open("mysql", config.DB)

	// mysql image starts need time.
	for {
		err := db.Ping()
		if err == nil {
			break
		}
		fmt.Println(err)
		time.Sleep(2 * time.Second)
	}

	// https://github.com/go-sql-driver/mysql/issues/674
	db.SetMaxIdleConns(0)

	//db.Exec(`
	//	CREATE TABLE IF NOT EXISTS users (
	//		id INT(10) NOT NULL AUTO_INCREMENT,
	//		username VARCHAR(16) NULL DEFAULT NULL,
	//		password VARCHAR(64) NULL DEFAULT NULL,
	//		email VARCHAR(64) NULL DEFAULT NULL,
	//		PRIMARY KEY (id)
	//	);`)
	// db.Exec(`INSERT INTO users(username, password, email) values("admin","adminn","a@a.com");`)
}

func getPassword(username string) (string, int) {
	var password string
	var id int
	rows, _ := db.Query("SELECT password,id from users where username=?", username)
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&password, &id)

	}
	return password, id
}

func getChallenges() []Challenge {

	var Challenges []Challenge
	var challenge Challenge

	rows, _ := db.Query("SELECT * from challenge")
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&challenge.Id, &challenge.Name, &challenge.Img, &challenge.Description, &challenge.Type)
		Challenges = append(Challenges, challenge)
	}
	return Challenges
}

func getChallengesStatus(uId int) []Challenge {
	var Challenges []Challenge
	var challenge Challenge

	rows, _ := db.Query("SELECT * from tasks where userid=?", uId)
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&challenge.Id, &challenge.Id, &challenge.StartTime, &challenge.Uid, &challenge.Url) //获取当前用户开启的实验信息
		Challenges = append(Challenges, challenge)
	}
	return Challenges
}

func startChallenges(cId, uId int) bool {

	//启动后 返回对应的url
	url := "http://127.0.0.1"
	// fmt.Printf("INSERT INTO tasks(challenge,start,userid,url) values (%d,%d,%d,'%s')", cId, time.Now().Unix(), uId, url)
	rows, err := db.Query("INSERT INTO tasks(challengeId,start,userid,url) VALUES (?,?,?,?)", int(cId), int(time.Now().Unix()), int(uId), url)
	defer rows.Close()
	if err != nil {
		return false
	} else {
		return true
	}
}
func getItems(uid int) []Items {
	var item Items
	var items []Items
	i := 0
	rows, _ := db.Query("SELECT id,content,time,uid,status from items where uid=?", uid)
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&item.Id, &item.Content, &item.Time, &item.Uid, &item.Status)
		items = append(items, item)
		i++
	}
	return items
}

func updateItems(id int, status int) int {
	rows, err := db.Query("UPDATE items set status=? from items where uid=?", status, id)
	defer rows.Close()
	if err != nil {
		return 0
	} else {
		return 1
	}
}

func addItems(item Items) bool {
	fmt.Println(item.Content)
	rows, err := db.Query("INSERT INTO items(content,time,uid,status) values (?,?,?,?)", item.Content, item.Time, item.Uid, item.Status)
	defer rows.Close()
	if err != nil {
		return false
	} else {
		return true
	}
}

func delItems(id int) int {
	rows, err := db.Query("delete from items where id=?", id)
	defer rows.Close()
	if err != nil {
		return 0
	} else {
		return 1
	}
}

type Items struct {
	Id      int
	Content string
	Time    int64
	Uid     int
	Status  int
}
