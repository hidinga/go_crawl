package yahoo

import (
    "io/ioutil"
    "os"
    "log"
    "net/http"
    "net/url"
    "golang.org/x/net/proxy"
    "fmt"
    "time"
    "strconv"
    "strings"
    "crawl/comm"
    "crawl/conf"
)

func GetData(t string ,d comm.DateTime) []byte {

	// 创建 SOCKS5 代理
    
	dialer, err := proxy.SOCKS5("tcp", conf.Ssh.Local, nil, proxy.Direct)
	if err != nil {
		log.Println(os.Stderr, "can't connect to the proxy:", err)
		os.Exit(1)
	}
    
	// 设置代理

    timeout := time.Duration(5 * time.Second)
    http.DefaultTransport = &http.Transport{Dial: dialer.Dial, ResponseHeaderTimeout:timeout}

    // 配置文件处理
    
    var ticker string = t
    
    if strings.Contains(t,":") {
        ticker = t[0: strings.Index(t, ":")]
    }

	// 请求地址
    
    params := url.Values{}
    params.Set("formatted", "true")
    params.Set("includeAdjustedClose", "true")
    params.Set("useYfid", "true")
    params.Set("interval", "1d")
    params.Set("period1", strconv.Itoa(int(d.StartTime)))
    params.Set("period2", strconv.Itoa(int(d.EndTime)))
    
    // 请求数据
    
    myURL, _ := url.Parse(fmt.Sprintf(conf.Web.Url, ticker))
    myURL.RawQuery = params.Encode()
    urlPath := myURL.String()
 
    log.Printf("%v\n", urlPath)
    
	if resp, err := http.Get(urlPath); err != nil {
        return []byte{}
	} else {
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
        return []byte(body)
	} 
}
