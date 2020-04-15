package service

import (
	"../model"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/go-xorm/xorm"
	"log"
)

var (
	DbEngin *xorm.Engine
	err error
	RedisConn *redis.Client
	RedisDatabases int
)
func  init()  {
	drivename :="mysql"
	DsName := "root:@(127.0.0.1:3306)/chat?charset=utf8"
	DbEngin,err = xorm.NewEngine(drivename,DsName)
	if nil!=err && ""!=err.Error() {
		log.Fatal(err.Error())
	}
	//是否显示SQL语句
	//DbEngin.ShowSQL(true)
	//数据库最大打开的连接数
	DbEngin.SetMaxOpenConns(2)

	//自动User
	err := DbEngin.Sync2(new(model.User),
		new(model.Contact),
		new(model.Community))
	if err != nil {
		log.Fatal(err.Error())
	}
	//DbEngin = dbengin
	fmt.Println("init MySQL conn ok")
}

func init()  {
	RedisConn = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "", // no password set
		DB:       RedisDatabases,  // use default DB
	})
	//Redis连接检测
	_, err := RedisConn.Ping().Result()
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	fmt.Println("init Redis conn ok")
}