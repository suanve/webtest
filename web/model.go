package web

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"webtest/config"
	"webtest/engine"

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

//获取用户的密码
func getPassword(username string) (string, int) {
	// ret 密码 id
	var password string
	var id int
	rows, _ := db.Query("SELECT password,id from users where username=?", username)
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&password, &id)

	}
	return password, id
}

//获取用户等级
func getLevel(username string) int {
	// ret 等级
	var level int
	rows, _ := db.Query("SELECT level from users where username=?", username)
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&level)
	}

	return level
}

func getChallenges() []Challenge {

	var Challenges []Challenge
	var challenge Challenge

	rows, _ := db.Query("SELECT * from challenge")
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&challenge.Id, &challenge.Name, &challenge.Img, &challenge.Description, &challenge.Type)
		challenge.Key = challenge.Id
		Challenges = append(Challenges, challenge)
	}
	return Challenges
}

func getChallengesStatus(uId int) []Challenge {
	var Challenges []Challenge
	var challenge Challenge

	rows, _ := db.Query("SELECT challengeId,start,userid,url from tasks where userid=?", uId)
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&challenge.Id, &challenge.StartTime, &challenge.Uid, &challenge.Url) //获取当前用户开启的实验信息
		Challenges = append(Challenges, challenge)
	}
	return Challenges
}

//停止对应的容器
func stopChallenges(cId, uId int, Username string) int {
	// ret 0,1,2
	// 0 失败
	// 1 成功
	// 2 该任务不存在
	var task Tasks
	rows, _ := db.Query("SELECT id,url,containerId from tasks where userid=? and challengeId=?", uId, cId)
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&task.Id, &task.Url, &task.ContainerId) //获取当前用户开启的实验信息
		if err != nil {
			return 2
		}
	}
	if task.ContainerId == "" {
		return 2
	}
	tmpPort := strings.Split(task.Url, ":")[2]
	port, _ := strconv.Atoi(tmpPort)
	rows, err := db.Query("delete from ports where port=?", port)
	defer rows.Close()
	if err != nil {
		return 0
	}
	rows, err = db.Query("delete from tasks where userid=? and challengeId=?", uId, cId)
	defer rows.Close()
	if err != nil {
		return 0
	}

	if engine.Ctr_StopContainer(task.ContainerId) == task.ContainerId {
		return 1
	} else {
		return 0
	}

}

func startChallenges(cId, uId int, Username string) int {
	// ret 0,1,2
	// 0 失败
	// 1 成功
	// 2 已存在
	var tasks []Tasks
	var task Tasks
	//启动后 返回对应的url

	//检测是否已经有开启了的容器
	rows, _ := db.Query("SELECT * from tasks where userid=? and challengeId=?", uId, cId)
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&task.Id, &task.ChallengeId, &task.Start, &task.Userid, &task.Url) //获取当前用户开启的实验信息
		tasks = append(tasks, task)
	}
	if len(tasks) > 0 {
		return 2
	}

	//获取实验的镜像名称
	var image string
	var inPort int //内部端口
	rows, _ = db.Query("SELECT image,inport from images where challengeId=?", cId)
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&image, &inPort) //获取镜像设置的内部端口
	}
	fmt.Println("要启动一个容器", image)
	//获取一个可用的端口号，不能重复

	var outPort int //映射后的端口
	for true {
		outPort = RandInt64(20000, 30000)

		rows, _ := db.Query("SELECT port from ports where port=?", outPort)
		defer rows.Close()
		if len(tasks) > 0 {
			continue
		} else {
			rows, err := db.Query("INSERT INTO ports(port) VALUES (?)", int(outPort))
			defer rows.Close()
			if err != nil {
				return 0
			} else {
				break
			}
			break
		}
	}
	//启动一个容器	image,inPort,outPort
	containerID := engine.Ctr_CreateContainer(image, inPort, outPort, Username, cId)
	if containerID == "" {
		return 0
	}

	//启动镜像，获取容器的端口号
	url := "http://" + config.HubHost + ":" + strconv.Itoa(outPort)
	//将容器的url存入数据库
	rows, err := db.Query("INSERT INTO tasks(challengeId,start,userid,url,containerId) VALUES (?,?,?,?,?)", int(cId), int(time.Now().Unix()), int(uId), url, containerID)
	defer rows.Close()
	if err != nil {
		return 0
	} else {
		return 1
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
