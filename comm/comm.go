package comm

import (
    "time"
    // "strconv"
    // "fmt"
)

type DateTime struct {
    StartTime   int64
    EndTime     int64 
    StartDate   string 
    EndDate     string
    StartFull   string 
    EndFull     string
}

type Yahoo struct {
    Chart struct {
        Result []struct {
            Meta struct {
                Symbol         string    `json:"symbol"`
                Gmtoffset      int64       `json:"gmtoffset"`
                ExchangeName   string    `json:"exchangeName"`
            } `json:"meta"`
            Timestamp          []int64     `json:"timestamp"`
            Indicators struct {
                Quote []struct {
                     High     []float64  `json:"high"`
                     Volume   []float64  `json:"volume"`
                     Open     []float64  `json:"open"`
                     Low      []float64  `json:"low"`
                     Close    []float64  `json:"close"`
                } `json:"quote"`
            } `json:"indicators"`
        } `json:"result"`
    } `json:"chart"`
}

type DataItem struct {
    Ticker  string
    Date    time.Time
    Point   float64
}

type QuoteItem struct {
    Ticker        string    `sql:"TICKER"`
    Trading_date  string    `sql:"TRADING_DATE"`
    Open          float64   `sql:"OPEN"`
    Highest       float64   `sql:"HIGHEST"`
    Lowest        float64   `sql:"LOWEST"`
    Lclose        float64   `sql:"LCLOSE"`
    Close         float64   `sql:"CLOSE"`
    Inx_dr        float64   `sql:"INX_DR"`
    Volume        float64   `sql:"VOLUME"`
}

func TimeToDate(n interface{}) string {
    switch n.(type) {
        case int64:
            i := n.(int64)
            return time.Unix(i, 0).Format(time.DateOnly)
        case time.Time:
            j := n.(time.Time)
            return j.Format(time.DateOnly)
        default:
            return ""
    }
    return ""
}

func QuoteData() []byte {
    return []byte(`{"chart":{"result":[{"meta":{"currency":"USD","symbol":"^SP500TR","exchangeName":"SNP","instrumentType":"INDEX","firstTradeDate":568305000,"regularMarketTime":1689369274,"gmtoffset":-14400,"timezone":"EDT","exchangeTimezoneName":"America/New_York","regularMarketPrice":9683.69,"chartPreviousClose":9468.81,"priceHint":2,"currentTradingPeriod":{"pre":{"timezone":"EDT","end":1689341400,"start":1689321600,"gmtoffset":-14400},"regular":{"timezone":"EDT","end":1689364800,"start":1689341400,"gmtoffset":-14400},"post":{"timezone":"EDT","end":1689379200,"start":1689364800,"gmtoffset":-14400}},"dataGranularity":"1d","range":"","validRanges":["1d","5d","1mo","3mo","6mo","1y","2y","5y","10y","ytd","max"]},"timestamp":[1687267800,1687354200,1687440600,1687527000,1687786200,1687872600,1687959000,1688045400,1688131800,1688391000,1688563800,1688650200,1688736600,1688995800,1689082200,1689168600,1689255000,1689341400],"indicators":{"quote":[{"high":[9448.51953125,9419.1796875,9411.5595703125,9377.849609375,9368.509765625,9416.5302734375,9429.0703125,9447.5400390625,9577.0703125,9572.9501953125,9569.240234375,9502.419921875,9542.099609375,9482.2998046875,9549.099609375,9645.2900390625,9708.9296875,9731.66015625],"close":[9424.01953125,9375.2099609375,9410.830078125,9338.83984375,9297.1201171875,9403.6201171875,9400.2998046875,9443.4599609375,9559.669921875,9571.349609375,9553.6904296875,9478.73046875,9453.099609375,9475.900390625,9539.9501953125,9610.8701171875,9693.2998046875,9683.6904296875],"low":[9378.6298828125,9363.2900390625,9346.650390625,9324.08984375,9295.73046875,9310.400390625,9364.9501953125,9391.259765625,9499.66015625,9542.759765625,9532.73046875,9421.740234375,9449.76953125,9434.01953125,9474.1103515625,9591.990234375,9648.9404296875,9671.3203125],"open":[9439.91015625,9405.9599609375,9353.9501953125,9351.3701171875,9331.51953125,9315.4697265625,9380.1396484375,9397.2900390625,9499.66015625,9560.3701171875,9543.41015625,9502.419921875,9465.1103515625,9443.0400390625,9489.0,9601.26953125,9653.4501953125,9703.4404296875],"volume":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]}],"adjclose":[{"adjclose":[9424.01953125,9375.2099609375,9410.830078125,9338.83984375,9297.1201171875,9403.6201171875,9400.2998046875,9443.4599609375,9559.669921875,9571.349609375,9553.6904296875,9478.73046875,9453.099609375,9475.900390625,9539.9501953125,9610.8701171875,9693.2998046875,9683.6904296875]}]}}],"error":null}}`)
}

func PrivateKey() []byte {
    return []byte(`-----BEGIN RSA PRIVATE KEY-----
-----END RSA PRIVATE KEY-----`)
}