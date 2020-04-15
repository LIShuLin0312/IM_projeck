# IM_projeck
这是一个基于websocket为核心的webapp项目。参考慕课


#5.4 实现发送文字、表情包等

前端user1拼接好数据对象Message
msg={id:1,userid:2,dstid:3,cmd:10,media:1,content:txt}
转化成json字符串jsonstr
jsonstr = JSON.stringify(msg)
通过websocket.send(jsonstr)发送
后端S在recvproc中接收收数据data
并做相应的逻辑处理dispatch(data)-转发给user2
user2通过websocket.onmessage收到消息后做解析并显示


###5.4.1 前端处理核心方法
前端所有的操作都在拼接数据
如何拼接?
```javascript
sendtxtmsg:function(txt){
//{id:1,userid:2,dstid:3,cmd:10,media:1,content:txt}
var msg =this.createmsgcontext();
//msg={"dstid":dstid,"cmd":cmd,"userid":userId()}
//选择某个好友/群的时候对dstid,cmd进行赋值
//userId()返回用户自己的id ,
// 从/chat/index.shtml?id=xx&token=yy中获得
//1文本类型
msg.media=1;msg.content=txt;
this.showmsg(userInfo(),msg);//显示自己发的文字
this.webSocket.send(JSON.stringify(msg))//发送
}

sendpicmsg:function(picurl){
    //{id:1,userid:2,dstid:3,cmd:10,media:4,
    // url:"http://www.baidu.com/a/log,jpg"}
    var msg =this.createmsgcontext();
    msg.media=4;
    msg.url=picurl;
    this.showmsg(userInfo(),msg)
    this.webSocket.send(JSON.stringify(msg))
}
sendaudiomsg:function(url,num){
    //{id:1,userid:2,dstid:3,cmd:10,media:3,url:"http://www.a,com/dsturl.mp3",anount:40}
    var msg =this.createmsgcontext();
    msg.media=3;
    msg.url=url;
    msg.amount = num;
    this.showmsg(userInfo(),msg)
    console.log("sendaudiomsg",this.msglist);
    this.webSocket.send(JSON.stringify(msg))
}
```

##5.4.2 后端逻辑处理函数 func dispatch(data[]byte)
```cgo
func dispatch(data[]byte){
    //todo 解析data为message
    
    //todo根据message的cmd属性做相应的处理
    
}
func recvproc(node *Node) {
	for{
		_,data,err := node.Conn.ReadMessage()
		if err!=nil{
			log.Println(err.Error())
			return
		}
		//todo 对data进一步处理
		dispatch(data)
		fmt.Printf("recv<=%s",data)
	}
}
```
###5.4.3 对端接收到消息后处理函数
```js
//初始化websocket的时候进行回调配置
this.webSocket.onmessage = function(evt){
     //{"data":"}",...}
     if(evt.data.indexOf("}")>-1){
         this.onmessage(JSON.parse(evt.data));
     }else{
         console.log("recv<=="+evt.data)
     }
 }.bind(this)
onmessage:function(data){
     this.loaduserinfo(data.userid,function(user){
         this.showmsg(user,data)
     }.bind(this))
 }

 //消息显示函数
showmsg:function(user,msg){
    var data={}
    data.ismine = userId()==msg.userid;
    //console.log(data.ismine,userId(),msg.userid)
    data.user = user;
    data.msg = msg;
    //vue 只需要修改数据结构即可完成页面渲染
    this.msglist = this.msglist.concat(data)
    //面板重置
    this.reset();
    var that =this;
    //滚动到新消息处
    that.timer = setTimeout(function(){
        window.scrollTo(0, document.getElementById("convo").offsetHeight);
        clearTimeout(that.timer)
    },100)
 }
```
###5.4.4 表情包简单逻辑
弹出一个窗口,
选择图片获得一个连接地址
调用sendpicmsg方法开始发送流程

##5.5 发送图片/拍照
弹出一个窗口,
选择图片,上传到服务器
获得一个链接地址
调用sendpicmsg方法开始发送流程
###5.5.1 界面处理技巧
```html
<input 
accept="image/gif,image/jpeg,,image/png" 
type="file" 
onchange="upload(this)" 
class='upload'/>
```
sendpicmsg方法开始发送流程
###5.5.2 upload前端实现
```javascript
function upload(dom){
        uploadfile("attach/upload",dom,function(res){
            if(res.code==0){//成功以后调用sendpicmsg
                vm.sendpicmsg(res.data)
            }
        })
    }
    
function uploadfile(uri,dom,callback){
    //H5新特性
    var formdata = new FormData();
    //获得一个文件dom.files[0]
    formdata.append("file",dom.files[0])
    //formdata.append("filetype",".png")//.mp3指定后缀
    
    var xhr = new XMLHttpRequest();//ajax初始化
    var url = "http://"+location.host+"/"+uri;
    //"http://127.0.0.1/attach/upload"
    xhr.open("POST",url, true);
    //成功时候回调
    xhr.onreadystatechange = function() {
        if (xhr.readyState == 4 && 
        xhr.status == 200) {
            //fn.call(this, JSON.parse(xhr.responseText));
            callback(JSON.parse(xhr.responseText))
        }
    };
    xhr.send(formdata);
}    
```
###5.5.2 upload后端实现
####存储到本地
```
func UploadLocal(writer http.ResponseWriter,
	request * http.Request){
	}
```
###存储到alioss
```
func UploadLocal(writer http.ResponseWriter,
	 request * http.Request){
}
如何安装 golang.org/x/time/rate
>cd $GOPATH/src/golang.org/x/
>git clone https://github.com/golang/time.git time

``` 
###5.6 发送语音
####5.6.1 采集语音
```javascript
navigator.mediaDevices.getUserMedia(
    {audio: true, video: true}
    ).then(successfunc).catch(errfunc);


navigator.mediaDevices.getUserMedia(
    {audio: true, video: false}
    ).then(function(stream)  {
              //请求成功
              this.recorder = new MediaRecorder(stream);
              this.recorder.start();
              this.recorder.ondataavailable = (event) => {
                  uploadblob("attach/upload",event.data,".mp3",res=>{
                      var duration = Math.ceil((new Date().getTime()-this.duration)/1000);
                      this.sendaudiomsg(res.data,duration);
                  })

                  stream.getTracks().forEach(function (track) {
                      track.stop();
                  });
                  this.showprocess = false
              }
              
          }.bind(this)).catch(function(err){
                mui.toast(err.msg)
                this.showprocess = false
            }.bind(this));
```
####5.6.2 上传语音
```javascript
function uploadblob(uri,blob,filetype,fn){
       var xhr = new XMLHttpRequest();
       xhr.open("POST","//"+location.host+"/"+uri, true);
       // 添加http头，发送信息至服务器时内容编码类型
       xhr.onreadystatechange = function() {
           if (xhr.readyState == 4 && (xhr.status == 200 || xhr.status == 304)) {
               fn.call(this, JSON.parse(xhr.responseText));
           }
       };
       var _data=[];
       var formdata = new FormData();
       formdata.append("filetype",filetype);
       formdata.append("file",blob)
       xhr.send(formdata);
   }
```

###5.7 实现群聊

####5.7.1 原理
分析群id,找到加了这个群的用户,把消息发送过去
方案一、
map<userid><qunid1,qunid2,qunid3>
优势是锁的频次低
劣势是要轮训全部map
```cgo
type Node struct {
	Conn *websocket.Conn
	//并行转串行,
	DataQueue chan []byte
	GroupSets set.Interface
}
//映射关系表
var clientMap map[int64]*Node = make(map[int64]*Node,0)
```
方案二、
map<群id><userid1,userid2,userid3>
优势是找用户ID非常快
劣势是发送信息时需要根据userid获取node,锁的频次太高
```cgo
type Node struct {
	Conn *websocket.Conn
	//并行转串行,
	DataQueue chan []byte
}
//映射关系表
var clientMap map[int64]*Node = make(map[int64]*Node,0)
var comMap map[int64]set.Interface= make(map[int64]set.Interface,0)

```
####5.7.2 需要处理的问题
```javascript
1、当用户接入的时候初始化groupset
2、当用户加入群的时候刷新groupset
3、完成信息分发
```
###5.8 性能优化
1 锁的频次
2 json编码次数
3 静态资源分离
```cgo
###存储到alioss
```
func UploadOss(writer http.ResponseWriter,
	 request * http.Request){
}
如何安装
>go get github.com/aliyun/aliyun-oss-go-sdk/oss
 >golang.org/x/time/rate
>cd $GOPATH/src/golang.org/x/
>git clone https://github.com/golang/time.git time

``` 

```