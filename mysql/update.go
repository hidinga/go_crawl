package db

import (
    "fmt"
    "time"
    "database/sql"
     _ "github.com/go-sql-driver/mysql"
    "encoding/json"
    "strings"
    "strconv"
    "gogo/comm"
    "gogo/conf"
)

// 定义全局 DB

var (
    DB      *sql.DB
    cache   map[string]map[string]float64
    symbol  map[string]string
    config  comm.DateTime
)

// 初始化数据库

func Init(d comm.DateTime) {
    
    config = d
    
    // 读取配置内映射
    
    symbol = make(map[string]string, 2)
    
    for _, v := range strings.Split(conf.Web.Symbol, ",") {
        i := strings.IndexAny(v,":")
        symbol[ v[0:i] ] = strings.TrimLeft(v[i:],":")
    }

    var err error
    
	DB, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/test?charset=utf8mb4&parseTime=True", 
        conf.Database.Username, 
        conf.Database.Password, 
        conf.Database.Host, 
        conf.Database.Port,
    ))
    
	if err != nil {
        panic(err)
	}
    DB.SetConnMaxLifetime(100)          // 设置数据库最大连接数
    DB.SetMaxIdleConns(10)              // 设置上数据库最大闲置连接数
    
    // 验证连接
    
    if err := DB.Ping(); err != nil {
        fmt.Println("Open Database Fail")
        return
    }
}

func LoadDataCache() {

    // 执行多条查询, 使用 Prepare 可防止SQL注入问题

    startT := time.Now()
  
    stmt, _ := DB.Prepare("SELECT TICKER,TRADING_DATE,CLOSE from fin_base.`fin_inx_dailyquote` where TICKER in ('SP0054','W00032') AND TRADING_DATE >= ? AND TRADING_DATE <= ?")
	defer stmt.Close()

	row, _ := stmt.Query(config.StartDate, config.EndDate)
	defer row.Close()   

    // 缓存数据
    
    cache = make(map[string]map[string]float64)
    var total = 0
    var item comm.DataItem

    for row.Next() {
    
        row.Scan(&item.Ticker, &item.Date, &item.Point)
        
        if _, ok := cache[item.Ticker]; !ok {
            cache[item.Ticker] = make(map[string]float64)
        }
        // cache[item.Ticker][fmt.Sprintf("%s", item.Date.Format(time.DateOnly))] = []float64{item.Point, item.Lpoint}
        cache[item.Ticker][comm.TimeToDate(item.Date)] = item.Point        
        total++
    }
    tc := time.Since(startT)	                                // 计算耗时
    
    fmt.Printf("加载已有数据 : %d 条, 耗时 %v\n\n", total, tc)
}

func UpdateData(b []byte) {

    var obj comm.Yahoo
    
    err := json.Unmarshal(b, &obj)
    
    if(err != nil){
        fmt.Printf("反序列化错误 err = %v\n", err)
        return
    }
    if len(obj.Chart.Result) == 0 {
        return
    }
    if len(obj.Chart.Result[0].Timestamp) == 0 {
        return
    }
    if len(obj.Chart.Result[0].Indicators.Quote) == 0 {
        return
    }
    
    fmt.Println()
    fmt.Println(string(b))
    
    // 处理数据
 
    var jswTicker = symbol[obj.Chart.Result[0].Meta.Symbol]
    var timesObj = obj.Chart.Result[0].Timestamp                // 日期
    var quoteObj = obj.Chart.Result[0].Indicators.Quote[0]      // 行情
    var lastPoint float64                                       // 前一日点位
    var quoteItem comm.QuoteItem                                // 单日数据
    var info string

    fmt.Println()

    for k, v := range timesObj {
    
        if k == 0 {
            continue
        }
        
        // 单日数据拼装
        
        lastPoint = quoteObj.Close[k-1]
        
        quoteItem = comm.QuoteItem{
            Ticker: jswTicker,
            Trading_date: comm.TimeToDate(v),
            Open: quoteObj.Open[k],
            Highest: quoteObj.High[k],
            Lowest: quoteObj.Low[k],
            Lclose: lastPoint,
            Close: quoteObj.Close[k],
            Inx_dr: (quoteObj.Close[k]-lastPoint)/lastPoint,
            Volume: quoteObj.Volume[k],
        }
        
        fmt.Printf(
            "%10s : %s {\"close\":%.4f, \"open\":%.4f, \"high\":%.4f, \"low\":%.4f, \"lclose\":%0.8f}\n",
            quoteItem.Trading_date,
            jswTicker,
            quoteObj.Close[k],    
            quoteObj.Open[k],
            quoteObj.High[k],
            quoteObj.Low[k],
            lastPoint,
        )
        
        // 修复 BUG : 数据日期早于交易实际时间
        
        if quoteItem.Trading_date == config.EndDate {
            continue
        }
        
        if _, ok := cache[jswTicker][quoteItem.Trading_date]; !ok {
            info += fmt.Sprintf("新增 %s %d : %s %.4f\n", jswTicker, v, quoteItem.Trading_date, quoteItem.Close)
            insertQuote(quoteItem)
            continue
        }
        
        if fmt.Sprintf("%.4f", quoteItem.Close) != fmt.Sprintf("%.4f", cache[jswTicker][quoteItem.Trading_date]) {
            info += fmt.Sprintf("更新 %s %d : %s %.4f => %.4f\n", jswTicker, v, quoteItem.Trading_date, cache[jswTicker][quoteItem.Trading_date], quoteItem.Close)
            updateQuote(quoteItem)
            continue
        }
    }
    
    fmt.Println()
    
    if info != "" {
        fmt.Printf("%s\n", info)
    }
}

func insertQuote(item comm.QuoteItem) {

    stmt, err := DB.Prepare("insert into fin_base.fin_inx_dailyquote(TICKER,TRADING_DATE,OPEN,HIGHEST,LOWEST,LCLOSE,CLOSE,INX_DR,VOLUME) VALUES (?,?,?,?,?,?,?,?,?)")  
    defer stmt.Close()
    
    if err != nil {  
        fmt.Println("新增数据错误", err)  
        return  
    }  
	row, _ := stmt.Exec(
        item.Ticker, 
        item.Trading_date, 
        fmt.Sprintf("%.4f", item.Open),
        fmt.Sprintf("%.4f", item.Highest),
        fmt.Sprintf("%.4f", item.Lowest),
        fmt.Sprintf("%.8f", item.Lclose),
        fmt.Sprintf("%.4f", item.Close),
        fmt.Sprintf("%.6f", item.Inx_dr),
        fmt.Sprintf("%.6f", item.Volume),
    )
    newID, _ := row.LastInsertId()                               // 新增数据的ID  
    fmt.Printf("\n新增数据 : %d\n", newID)
    
    // update cache
    
    newPoint, err := strconv.ParseFloat(fmt.Sprintf("%.4f", item.Close), 64)
    cache[item.Ticker][item.Trading_date] = newPoint
}

func updateQuote(item comm.QuoteItem) {

    stmt, err := DB.Prepare("update fin_base.fin_inx_dailyquote set OPEN=?,HIGHEST=?,LOWEST=?,LCLOSE=?,CLOSE=?,INX_DR=?,VOLUME=? where TICKER=? AND TRADING_DATE=?")  
    defer stmt.Close()
    
    if err != nil {  
        fmt.Println("修改数据错误", err)  
        return  
    }  
	row, _ := stmt.Exec(
        fmt.Sprintf("%.4f", item.Open),
        fmt.Sprintf("%.4f", item.Highest),
        fmt.Sprintf("%.4f", item.Lowest),
        fmt.Sprintf("%.8f", item.Lclose),
        fmt.Sprintf("%.4f", item.Close),
        fmt.Sprintf("%.6f", item.Inx_dr),
        fmt.Sprintf("%.6f", item.Volume),
        item.Ticker, 
        item.Trading_date, 
    )
    total, _ := row.RowsAffected()
    fmt.Printf("\n影响行数 : %d \n", total) 
    
    // update cache
    
    newPoint, err := strconv.ParseFloat(fmt.Sprintf("%.4f", item.Close), 64)
    cache[item.Ticker][item.Trading_date] = newPoint
}