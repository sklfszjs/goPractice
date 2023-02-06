package main

import(
  "net"
  "strings"
)

type User struct {
  Name string
  Addr string 
  C chan string
  conn net.Conn
  server *Server
}

func NewUser(conn net.Conn,server *Server)  *User {
  userAddr:=conn.RemoteAddr().String()
  user:=&User{
    Name:userAddr,
    Addr:userAddr,
    C: make(chan string),
    conn:conn,
    server:server,
  }
  go user.ListenMessager()
  return user
}

func (this *User)DoMessage(msg string){
  if msg=="who"{
    this.server.mapLock.Lock()
    for _,user :=range this.server.OnlineMap{
      onlineMessage:= "["+user.Addr+"]"+user.Name+"在线"
      this.SendMessage(onlineMessage)
    }
    this.server.mapLock.Unlock()
  }else if len(msg)>6 && msg[:7]=="rename|"{
    newName:=strings.Split(msg,"|")[1]
    _,ok:=this.server.OnlineMap[newName]
    if ok{
      this.SendMessage("当前用户名已经被使用\n")
    }else{
      this.server.mapLock.Lock()
      delete(this.server.OnlineMap,this.Name)
      this.server.OnlineMap[newName]=this
      this.server.mapLock.Unlock()
      this.Name=newName
      this.SendMessage("您已经重置用户名")
    }
  }else if len(msg)>4 && msg[:3]=="to|"{
    targetusername:=strings.Split(msg,"|")[1]

    if targetusername == ""||len(strings.Split(msg,"|"))==2{
      this.SendMessage("格式不正确")
      return
      
    }
    msg=strings.Split(msg,"|")[2]
    targetuser,ok:=this.server.OnlineMap[targetusername]
    if !ok {
      this.SendMessage("用户不存在")

    }else{
      targetuser.SendMessage(this.Name+"对您说"+msg)
    }

  } else{
   this.server.BroadCast(this, msg)
  }
}

func (this *User)Online(){
  this.server.mapLock.Lock()
  this.server.OnlineMap[this.Name]=this
  this.server.mapLock.Unlock()
  this.server.BroadCast(this,"已经上线")

}

func (this *User)Offline(){
  this.server.mapLock.Lock()
  delete(this.server.OnlineMap,this.Name)
  this.server.mapLock.Unlock()
  this.server.BroadCast(this,"已经下线")

}

func(this *User)SendMessage(msg string){
  this.conn.Write([]byte(msg))
}


func (this *User)ListenMessager(){
  for{
    msg:=<-this.C
    this.conn.Write([]byte(msg+"\n"))
  }
} 
