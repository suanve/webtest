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

// 获取所有的实验
func getChallenges() []Challenge {

	var Challenges []Challenge
	var challenge Challenge

	rows, _ := db.Query("SELECT id,Name,Img,Description,Type,Image from challenges")
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&challenge.Id, &challenge.Name, &challenge.Img, &challenge.Description, &challenge.Type, &challenge.Image)
		challenge.Key = challenge.Id
		Challenges = append(Challenges, challenge)
	}
	return Challenges
}

// 获取指定实验的信息
func getChallenge(cId int) []Challenge {

	var Challenges []Challenge
	var challenge Challenge

	rows, _ := db.Query("SELECT * from challenges where id=?", cId)
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&challenge.Id, &challenge.Name, &challenge.Img, &challenge.Description, &challenge.Type, &challenge.Image, &challenge.Inport)
		// challenge.Image = getChallengeToImageName(cId)
		Challenges = append(Challenges, challenge)
	}
	return Challenges
}

// 获取实验id对应的镜像名称
func getChallengeToImageName(cId int) string {

	var challengeId string
	rows, _ := db.Query("SELECT image from images where challengeId=?", cId)
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&challengeId)
	}
	return challengeId
}

//获取用户开启的实验信息
func getChallengesStatus(uId int) []Challenge {
	var Challenges []Challenge
	var challenge Challenge

	rows, _ := db.Query("select id,Name,Img,Description,Type from challenges")
	defer rows.Close()
	for rows.Next() {
		challenge = Challenge{}
		rows.Scan(&challenge.Id, &challenge.Name, &challenge.Img, &challenge.Description, &challenge.Type) //获取当前用户开启的实验信息

		rowsTask, _ := db.Query("select url from tasks where userid=? and challengeId=?", uId, challenge.Id)
		defer rowsTask.Close()
		for rowsTask.Next() {
			rowsTask.Scan(&challenge.Description) //获取当前用户开启的实验信息

		}

		fmt.Println(challenge)
		Challenges = append(Challenges, challenge)
	}
	return Challenges
}

// 后台停止容器
func stopContainer(challenge Challenge) int {
	// ret 0,1,2
	// 0 失败
	// 1 成功
	// 2 该任务不存在
	var task Tasks
	rows, _ := db.Query("SELECT challengeId,userid from tasks where id=?", challenge.Id)
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&task.ChallengeId, &task.Userid)
		if err != nil {
			return 2
		}
	}
	fmt.Println(task.ChallengeId, task.Userid, challenge.Username)
	return stopChallenge(task.ChallengeId, task.Userid, challenge.Username)

}

//停止对应的容器
func stopChallenge(cId, uId int, Username string) int {
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

func startChallenge(cId, uId int, Username string) int {
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
	rows, _ = db.Query("SELECT Image,Inport from challenges where id=?", cId)
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

// 添加实验信息
func addChallenge(challenge Challenge) int {
	// 插入实验
	fmt.Println("INSERT INTO challenge(Name,Img,Description,Type,Image,Inport) VALUES(?,?,?,?,?,?)", challenge.Name, challenge.Img, challenge.Description, challenge.Type, challenge.Image, challenge.Inport)
	rows, err := db.Query("INSERT INTO challenge(Name,Img,Description,Type,Image,Inport) VALUES(?,?,?,?,?,?)", challenge.Name, challenge.Img, challenge.Description, challenge.Type, challenge.Image, challenge.Inport)
	defer rows.Close()
	if err != nil {
		return 0
	}
	return 1
}

// 更新实验信息
func updateChallenge(challenge Challenge) int {
	//更新实验表
	rows, err := db.Query("UPDATE challenge set Name=?,Img=?,Description=?,Type=?,Image=?,Inport=? where id=?", challenge.Name, challenge.Img, challenge.Description, challenge.Type, challenge.Image, challenge.Inport, challenge.Id)
	defer rows.Close()
	if err != nil {
		return 0
	}
	return 1
}

// 删除实验信息
func delChallenge(cId int) int {
	fmt.Println("delete from challenges where id=?", cId)
	rows, err := db.Query("delete from challenges where id=?", cId)
	defer rows.Close()
	if err != nil {
		return 0
	} else {
		return 1
	}
}

//获取用户开启的容器
func getContainers() []Challenge {
	var Challenges []Challenge
	var challenge Challenge

	rows, _ := db.Query("SELECT id,challengeId,start,userid,url from tasks")
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&challenge.Key, &challenge.Id, &challenge.StartTime, &challenge.Uid, &challenge.Url) //获取已经开启的容器

		rowsImage, _ := db.Query("SELECT Image,Name,Type from challenges where id=?", challenge.Id) //获取对应容器的名称与镜像名称
		defer rowsImage.Close()
		for rowsImage.Next() {
			rowsImage.Scan(&challenge.Image, &challenge.Name, &challenge.Type)
		}
		rowsUser, _ := db.Query("SELECT username from users where id=?", challenge.Uid) //获取对应的用户
		defer rowsUser.Close()
		for rowsUser.Next() {
			rowsUser.Scan(&challenge.Username)
		}

		Challenges = append(Challenges, challenge)
	}
	return Challenges
}

// 获取所有的用户
func getUsers() []User {

	var users []User
	var user User

	rows, _ := db.Query("SELECT id,username,level from users")
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&user.Id, &user.Username, &user.Level)
		users = append(users, user)
	}
	return users
}

// 获取指定的用户信息
func getUser(uId int) []User {

	var users []User
	var user User

	rows, _ := db.Query("SELECT id,username,password,level from users where id=?", uId)
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&user.Id, &user.Username, &user.Password, &user.Level)
		users = append(users, user)
	}
	return users
}

// 添加用户
func addUser(user User) int {

	fmt.Println("INSERT INTO users(username,password,level) VALUES(?,?,?)", user.Username, user.Password, user.Level)
	rows, err := db.Query("INSERT INTO users(username,password,level) VALUES(?,?,?)", user.Username, user.Password, user.Level)
	defer rows.Close()
	if err != nil {
		return 0
	}
	return 1
}

// 删除用户
func delUser(uId int) int {
	fmt.Println("delete from users where id=?", uId)
	rows, err := db.Query("delete from users where id=?", uId)
	defer rows.Close()
	if err != nil {
		return 0
	} else {
		return 1
	}
}

// 更新用户信息
func updateUser(user User) int {

	rows, err := db.Query("UPDATE users set username=?,password=?,level=? where id=?", user.Username, user.Password, user.Level, user.Id)
	defer rows.Close()
	if err != nil {
		return 0
	}
	return 1
}
