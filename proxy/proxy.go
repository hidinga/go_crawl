package myProxy

import (
    "io"
    "log"
    "net"
    "strconv"
    "golang.org/x/crypto/ssh"
    "gogo/conf"
)

var (
    tunnel *ssh.Client
    local  net.Listener
    err    error
)

// 监听本地 2080 端口转发

func Init() {

    // log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
    log.SetFlags(log.Ldate | log.Ltime )

    // 监听本地端口 2080
    
	local, err = net.Listen("tcp", conf.Ssh.Local)
    
	if err != nil {
        log.Fatalln(err)
	}
    log.Println("监听成功", conf.Ssh.Local)
}

func NewListener(ready chan bool) {

    /* ready 创建 ssh tunnel 连接成功 */
    /* done 任务处理完成关闭 ssh tunnel 连接 */
    
    buff := []byte(conf.Ver)
    
    // SSH 服务器登录配置

    key, _ := ssh.ParsePrivateKey(buff)

    cfg := &ssh.ClientConfig{
        User: conf.Ssh.User,        
        Auth: []ssh.AuthMethod{
            ssh.PublicKeys(key),
        },
        HostKeyCallback: ssh.InsecureIgnoreHostKey(),
    }

    // 连接服务器

    tunnel, err = ssh.Dial("tcp", conf.Ssh.Server, cfg)
    if err != nil {
        log.Fatalln("连接失败 SSH server", err)
    }
    log.Println("连接成功 SSH server")

    // 回调

    ready <- true

	for {
		conn, err := local.Accept()
		if err != nil {
			log.Println(err)
			return
		}
		go handleClientRequest(conn) 
	}
}

func CloseConn() {
    defer tunnel.Close()
    log.Println("关闭 SSH 连接")
}

func handleClientRequest(conn net.Conn) {       
  
	defer conn.Close()
    
    // 接收数据
    
    var b [1024]byte
    
    // 读取 1024 字节数据
    
	n, err := conn.Read(b[:])
	if err != nil {
		log.Println(err)
	}
	// log.Printf("%x", b[:n])
 
    // 判断 SOCKS5 协议
    
    if b[0] == 0x05 {
    
        // 响应客户端版本
        
		conn.Write([]byte{0x05, 0x00})
        
        // 继续读取数据
        
		n, err = conn.Read(b[:])
        
		if err != nil || n < 4 {
			log.Fatal("handleClientRequest Error:", err)
			return
		}
        
		var host, port string
        
		switch b[3] {
            case 0x01: // IP V4
                host = net.IPv4(b[4], b[5], b[6], b[7]).String()
            case 0x03: // 域名
                host = string(b[5 : n-2]) // b[4]表示域名的长度
            case 0x04: // IP V6
                host = net.IP{b[4], b[5], b[6], b[7], b[8], b[9], b[10], b[11], b[12], b[13], b[14], b[15], b[16], b[17], b[18], b[19]}.String()
		}
		port = strconv.Itoa(int(b[n-2]) << 8 | int(b[n-1]))

        // 通过隧道连接服务器
        
		server, err := tunnel.Dial("tcp", net.JoinHostPort(host, port))
		if err != nil {
			log.Printf("handleClientRequest Dial(%s, %s) Failed! Error: %v\n", host, port, err)
			return
		}
		log.Printf("New Connection, Dial(%s:%s)\n", host, port)
		defer server.Close()
        
        // 响应客户端连接成功
        
		conn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})    
       
        go io.Copy(server, conn)
        io.Copy(conn, server)
    }
}