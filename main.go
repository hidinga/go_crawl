package main

import (
    "cawal/proxy"
    "cawal/yahoo"
    "cawal/conf"
    "cawal/mysql"
    "cawal/comm"
    "github.com/robfig/cron/v3"
    "fmt"
    "time"
    "strings"
)

var (
    d comm.DateTime
)

func main() {
    
    // 加载配置文件
    
    conf.Init()
    myProxy.Init()

    // 当前任务
    
    worker()
    
    // 创建新定时器
    
    c := cron.New()
    
    c.AddFunc(fmt.Sprintf("0 %s * * *", conf.Web.Period), func(){
        worker()
    })
    c.Start()
    
    for {
        time.Sleep(time.Second)
    }
}

func worker() {

    d = comm.DateTime{
        StartTime: time.Now().AddDate(0,-1,0).Unix(),
          EndTime: time.Now().Unix(),
        StartDate: time.Now().AddDate(0,-1,0).Format(time.DateOnly),
          EndDate: time.Now().Format(time.DateOnly),
        StartFull: time.Now().AddDate(0,-1,0).Format(time.DateTime),
          EndFull: time.Now().Format(time.DateTime),
    }
    
    // 连接代理

    var ready chan bool = make(chan bool, 1)     // 确认连接成功

    go myProxy.NewListener(ready)
    
    ref, ok := <-ready 
    
    if !ok {
        fmt.Printf("\n代理连接失败 : %#v\n\n", ref)
        return
    }
    fmt.Printf("\n代理连接成功 : %#v\n", ref)
    
    // 更新数据

    db.Init(d)
    db.LoadDataCache()

    var quoteData []byte
    
    for _, v := range strings.Split(conf.Web.Symbol, ",") {
    
        fmt.Printf("获取行情 : %v\n", v)
        fmt.Printf("时间区间 : %v %v\n\n", d.StartFull, d.EndFull)
        
        quoteData = yahoo.GetData(v, d)
        db.UpdateData(quoteData)
    }
    
    // 任务结束

    myProxy.CloseConn()
}