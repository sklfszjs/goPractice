package main
import(
  "net"
  "sync"
  "fmt"
  "io"
  "time"
  "runtime"
)

type Server struct{
  Ip string
  Port int
  OnlineMap map[string]*User
  mapLock sync.RWMutex
  Message chan string
}

func NewServer (ip string ,port int) *Server{
  server:=&Server{
    Ip:ip,
    Port:port,

    OnlineMap: make(map[string]*User),
    Message: make(chan string), 
  }
  return server
}

func(this *Server)DeleteUser(user *User){
  this.mapLock.Lock()
  delete(this.OnlineMap,user.Name)
  this.mapLock.Unlock()
  close(user.C)
  user.conn.Close()
}

func(this *Server)BroadCast(user *User,msg string){
  sendMsg:=msg
  this.Message<-sendMsg
}

func(this *Server)MessageListener(){
  for{
    msg:=<-this.Message
    this.mapLock.Lock()
    for name,cli:=range this.OnlineMap{
      cli.C<-msg
      fmt.Println(name)
    }
    this.mapLock.Unlock()
  }
}

func (this *Server) handler (conn net.Conn){
  user:=NewUser(conn,this)
  user.Online()
  isalive:=make(chan bool)
  go func(){
    buf:=make([]byte,4096)
    for{
      n,err:=conn.Read(buf)
      if n==0{
        user.Offline()
        return
      }
      if err!=nil && err!=io.EOF{
        fmt.Println("Conn read err",err)
      }
      msg:=string(buf[:n-1])
      user.DoMessage(msg)
      isalive<-true
    }
  }()
  for{

    select{
    case <-isalive:
    case <-time.After(time.Second*1000):
      user.SendMessage("你被踢了")
      this.DeleteUser(user)
      runtime.Goexit()
    }
  }

}

func(this *Server) Start(){
  listener,err:=net.Listen("tcp",fmt.Sprintf("%s:%d",this.Ip,this.Port))
  if err!=nil{
    fmt.Println("net.Listen error",err)
    return
  }
  defer listener.Close()
  go this.MessageListener()
  for{
    conn,err:=listener.Accept()
    if err!=nil {
      fmt.Println("accept error",err)
      continue
    }

    
    go  this.handler(conn)


  }
  
}

func main(){
  server:=NewServer("127.0.0.1",8888)
  server.Start()
}
