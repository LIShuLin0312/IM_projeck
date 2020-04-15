package ctrl

import (
	"../model"
	"../service"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"gopkg.in/fatih/set.v0"
	"log"
	"net/http"
	"strconv"
	"sync"
)

const (
	CMD_SINGLE_MSG = 10
	CMD_ROOM_MSG   = 11
	CMD_HEART      = 0
)
type Message struct {
	Id      int64  `json:"id,omitempty" form:"id"` //消息ID
	Userid  int64  `json:"userid,omitempty" form:"userid"` //谁发的
	Cmd     int    `json:"cmd,omitempty" form:"cmd"` //群聊还是私聊
	Dstid   int64  `json:"dstid,omitempty" form:"dstid"`//对端用户ID/群ID
	Media   int    `json:"media,omitempty" form:"media"` //消息按照什么样式展示
	Content string `json:"content,omitempty" form:"content"` //消息的内容
	Pic     string `json:"pic,omitempty" form:"pic"` //预览图片
	Url     string `json:"url,omitempty" form:"url"` //服务的URL
	Memo    string `json:"memo,omitempty" form:"memo"` //简单描述
	Amount  int    `json:"amount,omitempty" form:"amount"` //其他和数字相关的
}
/**
消息发送结构体
1、MEDIA_TYPE_TEXT
{id:1,userid:2,dstid:3,cmd:10,media:1,content:"hello"}
2、MEDIA_TYPE_News
{id:1,userid:2,dstid:3,cmd:10,media:2,content:"标题",pic:"http://www.baidu.com/a/log,jpg",url:"http://www.a,com/dsturl","memo":"这是描述"}
3、MEDIA_TYPE_VOICE，amount单位秒
{id:1,userid:2,dstid:3,cmd:10,media:3,url:"http://www.a,com/dsturl.mp3",anount:40}
4、MEDIA_TYPE_IMG
{id:1,userid:2,dstid:3,cmd:10,media:4,url:"http://www.baidu.com/a/log,jpg"}
5、MEDIA_TYPE_REDPACKAGR //红包amount 单位分
{id:1,userid:2,dstid:3,cmd:10,media:5,url:"http://www.baidu.com/a/b/c/redpackageaddress?id=100000","amount":300,"memo":"恭喜发财"}
6、MEDIA_TYPE_EMOJ 6
{id:1,userid:2,dstid:3,cmd:10,media:6,"content":"cry"}
7、MEDIA_TYPE_Link 6
{id:1,userid:2,dstid:3,cmd:10,media:7,"url":"http://www.a,com/dsturl.html"}

7、MEDIA_TYPE_Link 6
{id:1,userid:2,dstid:3,cmd:10,media:7,"url":"http://www.a,com/dsturl.html"}

8、MEDIA_TYPE_VIDEO 8
{id:1,userid:2,dstid:3,cmd:10,media:8,pic:"http://www.baidu.com/a/log,jpg",url:"http://www.a,com/a.mp4"}

9、MEDIA_TYPE_CONTACT 9
{id:1,userid:2,dstid:3,cmd:10,media:9,"content":"10086","pic":"http://www.baidu.com/a/avatar,jpg","memo":"胡大力"}

*/


//本核心在于形成userid和Node的映射关系
type Node struct {
	Conn *websocket.Conn
	//并行转串行,
	DataQueue chan []byte
	GroupSets set.Interface
}
//映射关系表
var clientMap map[int64]*Node = make(map[int64]*Node,0)
//读写锁
var rwlocker sync.RWMutex


//
// ws://127.0.0.1/chat?id=1&token=xxxx
func Chat(writer http.ResponseWriter,
	request *http.Request) {

	//todo 检验接入是否合法
    //checkToken(userId int64,token string)
    query := request.URL.Query()
    id := query.Get("id")
    token := query.Get("token")
    userId ,_ := strconv.ParseInt(id,10,64)
	isvalida := checkToken(userId,token)
	//如果isvalida=true
	//isvalida=false

	conn,err :=(&websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return isvalida
		},
	}).Upgrade(writer,request,nil)
	if err!=nil{
		log.Println(err.Error())
		return
	}
	//todo 获得conn
	node := &Node{
		Conn:conn,//Websocket连接节点
		DataQueue:make(chan []byte,50),//并行转串行数据.保证数据的发送顺序
		GroupSets:set.New(set.ThreadSafe),
	}
	//todo 获取用户全部群Id
	comIds := contactService.SearchComunityIds(userId)
	for _,v:=range comIds{
		node.GroupSets.Add(v)
	}
	//todo userid和node形成绑定关系
	rwlocker.Lock()
	clientMap[userId]=node //为了保证绑定关系的安全,需要读写锁控制
	rwlocker.Unlock()
	//todo 完成发送逻辑,con
	go sendproc(node,userId)
	//todo 完成接收逻辑
	go recvproc(node)
}

//发送协程
func sendproc(node *Node,userId int64) {
	lLen := service.RedisConn.LLen(fmt.Sprintf("%d", userId))//查询是否有离线缓存消息
	if lLen.Val()>0{
		//获取离线缓存的全部数据
		for i:=int64(0);i<lLen.Val();i++{
			data := service.RedisConn.LPop(fmt.Sprintf("%d",userId))
			err := node.Conn.WriteMessage(websocket.TextMessage,[]byte(data.Val()))
			if err!=nil{
				log.Println(err.Error())
				return
			}
		}
	}
	for {
		select {
			case data:= <-node.DataQueue:
				err := node.Conn.WriteMessage(websocket.TextMessage,data)
			    if err!=nil{
			    	log.Println(err.Error())
			    	return
				}
		}
	}
}
//todo 添加新的群ID到用户的groupset中
func AddGroupId(userId,gid int64){
	//取得node
	rwlocker.Lock()
	node,ok := clientMap[userId]
	if ok{
		node.GroupSets.Add(gid)
	}
	//clientMap[userId] = node
	rwlocker.Unlock()
	//添加gid到set
}
//接收协程
func recvproc(node *Node) {
	for{
		_,data,err := node.Conn.ReadMessage()//读取websocket数据返回3个变量.1:数据类型,2:数据本身,3:错误信息
		if err!=nil{
			log.Println(err.Error())
			return
		}
		//todo 对data进一步处理
		dispatch(data)
		//fmt.Printf("recv<=%s",data)
	}
}
//后端调度逻辑处理
func dispatch(data[]byte){
	//todo 解析data为message
	msg := Message{}
	err := json.Unmarshal(data,&msg)
	if err!=nil{
		log.Println(err.Error())
		return
	}
	//todo 根据cmd对逻辑进行处理
	switch msg.Cmd {
	case CMD_SINGLE_MSG:
		sendMsg(msg.Dstid,data)
	case CMD_ROOM_MSG:
		//todo 群聊转发逻辑
		for _,v:= range clientMap{
			if v.GroupSets.Has(msg.Dstid){
				v.DataQueue<-data
			}
		}
	case CMD_HEART:
		//todo 一般啥都不做
	}
}

//todo 发送消息
func sendMsg(userId int64,msg []byte) {
	rwlocker.RLock() //不直接用sync.LOCK是为了保证是同一把锁
	node,ok:=clientMap[userId] //读取用户websocket数据,保护存储用户的map安全,此处需要加锁
	rwlocker.RUnlock()
	if ok{
		node.DataQueue<- msg
	}else {
		userinfo := model.User{}
		exist, err := service.DbEngin.Where("id = ?", userId).Get(&userinfo)
		if err != nil {
			log.Println(err.Error())
		}
		if exist{
			service.RedisConn.RPush(fmt.Sprintf("%d", userinfo.Id),msg)
		}
	}
}
//检测是否有效
func checkToken(userId int64,token string)bool{
	//从数据库里面查询并比对
	user := userService.Find(userId)
	return user.Token==token
}