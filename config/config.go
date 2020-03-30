package config

import (
	"fmt"
)

var (
	// RootDir of your app
	RootDir = "./web"

	// SecretKey computes sg_token
	SecretKey string

	// DB addr and passwd.
	DB string

	DockerHost = "tcp://127.0.0.1:2376"
	HubHost    = "192.168.104.233"
)

func init() {
	//DB = fmt.Sprintf("%v:%v@tcp(mariadb:3306)/%v?charset=utf8",
	DB = fmt.Sprintf("%v:%v@tcp(127.0.0.1:3306)/%v?charset=utf8",
		//os.Getenv("DB_User"),
		//os.Getenv("DB_Passwd"),
		//os.Getenv("DB_Db"))
		"root",
		"root",
		"webtest")
	//SecretKey = os.Getenv("SecretKey")
	SecretKey = "biubiubiu"
}
