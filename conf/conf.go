package conf

import (
    "bufio"
    "strings"
    "fmt"
    "log"
    "os"
    "path"
    "path/filepath"
    "runtime"
    "reflect"
    "strconv"
    "io/ioutil"
)

type SshCfg struct {
    User    string      `ini:"user"`
    Local   string      `ini:"local"`
    Server  string      `ini:"server"`
}

type DatabaseCfg struct {
    Host     string     `ini:"host"`
    Port     int        `ini:"port"`
    Username string     `ini:"username"`
    Password string     `ini:"password"`
}

type WebCfg struct {
    Url     string      `ini:"url"`
    Symbol  string      `ini:"symbol"`
    Period  string      `ini:"period"`
}

var (
    Ver      string
    Ssh      SshCfg
    Database DatabaseCfg
    Web      WebCfg
)

func getCurrentPath() string{

    var absPath string
    
    // 读取临时目录
    
    dir := os.Getenv("TEMP")
	if dir == "" {
		dir = os.Getenv("TMP")
	}
	tp, _ := filepath.EvalSymlinks(dir)
    
    // 读取程序所在目录
    
    exePath, _ := os.Executable()
    ep, _ := filepath.EvalSymlinks(filepath.Dir(exePath))

    // 设定目录
    
    absPath = ep
    
	if strings.Contains(ep, tp)  {
        if _, fn, _, ok := runtime.Caller(0); ok {
            absPath = path.Dir(path.Dir(fn))
        }
	}
    if runtime.GOOS == "windows" {
        return strings.ReplaceAll(absPath, "/", "\\") + string(os.PathSeparator)
    }
    return strings.ReplaceAll(absPath, "\\", "/") + string(os.PathSeparator)
}

func Init() {

    // 初始化配置文件
    
    Ssh = SshCfg{
        User: "root", Local: "127.0.0.1:2080", Server: "172.104.162.206:22",
    }
    
    Database = DatabaseCfg{
        Host: "162.14.131.88", Port: 3306, Username: "fund", Password: "jzzG123",
    }
    
    Web = WebCfg{
        Url: "https://query1.finance.yahoo.com/v8/finance/chart/%s", 
        Symbol: "^SP500TR:SP0054,^FTSE:W00032",
        Period: "8,9,10",
    }
    
    // 配置文件

    cfgPath := getCurrentPath() + "id_rsa"
    iniPath := getCurrentPath() + "cfg.ini"
 
    // 读取私钥

    fmt.Printf("\n加载私钥文件 : %s\n", cfgPath)
    
    if _, err := os.Stat(cfgPath); err != nil {
        if os.IsNotExist(err) {
            fmt.Printf("私钥文件不存在 : %s\n", cfgPath)
            return
        }
    }
    
    rsa, err := os.Open(cfgPath)
    if err != nil {
        return
    }
    defer rsa.Close()
    
    ver, _ := ioutil.ReadAll(rsa)
    Ver = string(ver)

    // 读取 ini 文件
    
    fmt.Printf("加载配置文件 : %s\n\n", iniPath)
    
    // 判断文件是否存在

    if _, err := os.Stat(iniPath); err != nil {
        if os.IsNotExist(err) {
            fmt.Printf("配置文件不存在 : %s\n", iniPath)
            return
        }
    }
        
    // 读取配置文件
        
    fh, err := os.Open(iniPath)
    
    if err != nil {
        log.Fatal(err)
    }
    defer fh.Close()

    // 按行读取文件
    
    p1 := reflect.ValueOf(&Ssh).Elem()
    p2 := reflect.ValueOf(&Database).Elem()
    p3 := reflect.ValueOf(&Web).Elem()
    
    line := bufio.NewScanner(fh)

    var k string            // 结构体第一层 key
    var p reflect.Value     // 当前结构体反射

    for line.Scan() {
    
        str := line.Text()   

        if len(strings.TrimSpace(str)) == 0 {
            continue
        }        
        if j := strings.IndexAny(str, "["); j != -1 {
            k = strings.Trim(str, "[]")
            continue
        }
        
        switch k {
            case "ssh":
               p = p1 
            case "database":
               p = p2
            case "web":
               p = p3
        }
        
        // k 为 ini 的 section
        // v 为 ini 的 键值对
        
        v := strings.Split(str, "=")
        
        // 通过反射修改结构体值

        for i := 0; i < p.NumField(); i++ {
        
            if p.Type().Field(i).Tag.Get("ini") != strings.TrimSpace(v[0]) {
                continue
            }
            if p.Field(i).Kind() == reflect.String {
                p.Field(i).SetString(strings.TrimSpace(v[1]))
            }
            
            if p.Field(i).Kind() == reflect.Int {
                x, _ := strconv.ParseInt(strings.TrimSpace(v[1]), 10, 64)
                p.Field(i).SetInt(x)
            }
        }

        /*
        // 嵌套结构体
        
        for i := 0; i < p.NumField(); i++ {
        
            if p.Field(i).Type().Kind() != reflect.Struct {
                continue
            }
            
            sub := p.Field(i)

            for j :=0 ; j< sub.Type().NumField(); j++ {
            
                if k == p.Type().Field(i).Tag.Get("ini") && strings.Title(strings.TrimSpace(v[0])) == sub.Type().Field(j).Name {

                    if sub.Field(j).Kind() == reflect.String {
                        sub.Field(j).SetString(strings.TrimSpace(v[1]))
                    }
                    
                    if sub.Field(j).Kind() == reflect.Int {
                        x, _ := strconv.ParseInt(strings.TrimSpace(v[1]), 10, 64)
                        sub.Field(j).SetInt(x)
                    }
                }
            }
        }
        */
    }
   
    if err := line.Err(); err != nil {
        log.Fatal(err)
    }    
}