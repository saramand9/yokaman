// Package yokaman
// yoka机器人事务统计客户端， 用于和服务端通讯
package yokaman

import (
	"encoding/binary"
	"fmt"
	"go.uber.org/zap"
	"sync"
	"sync/atomic"
	"time"
)

const (
	SUCCESS    = 0
	FAILURE    = 1
	NORESPONSE = 2
)

type ReqMetrics struct {
	Trans    string
	Reqtime  int64
	Resptime int64
	Code     int8
	Robotid  uint
}

type YoKaMan struct {
	testid    int32 //测试id
	nodeid    int8  //机器
	projectid uint
	cache     *Cache

	DataCli         *MetricsNetCli
	CmdCli          *MetricsCmdCli
	WebCli          *WebCli
	storeCli        *Localstorage
	totalpkg        int64
	totalpkgrecv    int64
	buffer          chan NetReqMetrics
	enableBackup    bool
	moniterBuffSize int32
}

var instance *YoKaMan
var once sync.Once

func YoKaManCli() *YoKaMan {
	once.Do(func() {
		instance = &YoKaMan{
			cache:        NewCache(),
			DataCli:      NewMetricsNetCli(),
			CmdCli:       NewMetricsCmdCli(),
			WebCli:       NewWebCli(),
			totalpkg:     0,
			totalpkgrecv: 0,
			enableBackup: false,
			buffer:       make(chan NetReqMetrics, 1024*1024),
		}
		instance.initLog()
	})

	return instance
}

func (cli *YoKaMan) initLog() {
	lc := LogConfig{
		Level:      "debug",
		FileName:   fmt.Sprintf("./logs/metrics/%v.log", time.Now().Format("20060102")),
		MaxSize:    100,
		MaxBackups: 5,
		MaxAge:     30,
	}
	err := InitLogger(lc)
	if err != nil {
		fmt.Println(err)
	}
	log := zap.L()
	defer log.Sync()
}

// SetTestInfo
// 如果执行多次测试，需要提前设置好测试id，以便区分不同测试场景
// testid:  每次测试的唯一id
// nodeid:  每个机器人的id
func (m *YoKaMan) SetTestInfo(testid int32, nodeid ...uint) {
	m.testid = testid //暂时不会超过256
	if len(nodeid) > 0 {
		m.nodeid = int8(nodeid[0]) //暂时不会超过256
	}

	m.storeCli = NewStorage(testid)
	m.storeCli.WriteHeader()
}

func (m *YoKaMan) SetProjectInfo(projectid uint, nodeid ...uint) {
	m.projectid = projectid
	if len(nodeid) > 0 {
		m.nodeid = int8(nodeid[0]) //暂时不会超过256
	}
}

func (m *YoKaMan) SetMetricsSvrAddr(addr string) {
	m.DataCli.metricaddr = addr
	m.CmdCli.metricaddr = addr
}

// SetAggregation 预留，用于在性能不足时，做数据聚合
// second:  表示多少秒内的数据做一次聚合。 注意如果做了数据聚合，StatMetrics中的robotid可能就不准确了
func (m *YoKaMan) SetAggregation(second uint) {

}

// 释放要做数据备份
func (m *YoKaMan) EnableBackup(flag bool) {
	m.enableBackup = flag
}

func StatReqMetrics(m ReqMetrics) error {
	err := YoKaManCli().StatReqMetrics(m)
	return err
}

// StatReqMetrics 用于统计事务qps，成功率等
// trans :  	事务名字
// reqtime:  请求时间， 以	time.Now().UnixMilli() 毫秒传入
// resptime: 响应时间， 以	time.Now().UnixMilli() 毫秒传入
// code :	事务业务标记， 0表示成功， 1表示失败， 2表示无响应，   其他可以自己传。
// robotid:  机器人id， 预留，用于后期跟踪每个机器人
func (cli *YoKaMan) StatReqMetrics(m ReqMetrics) error {
	//目前是顺序执行，这就意味着当事务类型特别多时，最开始在获取每个事务映射时，可能会花去较多时间。
	//这块如果期望优化，可以考虑本地缓存，或者增加协程，同步获取多个事务映射
	//如果本地已经存在映射关系，取本地id
	cli.totalpkgrecv++
	if cli.totalpkgrecv%1000000 == 0 {
		fmt.Printf("【%v】 %d metrics to start\n", time.Now().Format("2006-01-02 15:04:05.00"), cli.totalpkgrecv)
	}

	id, err := cli.cache.Get(m.Trans)
	//如果本地没有映射关系，需要跟svr去注册， 本地没有映射关系，则从服务器同步
	if err != nil {
		id, err = cli.CmdCli.RegisterRequest(m.Trans, uint32(cli.testid))
		if err != nil {
			return err
		}
		cli.cache.Set(m.Trans, id)
	}

	metrics := NetReqMetrics{
		Transid: int8(id),
		Start:   m.Reqtime,
		End:     m.Resptime,
		Code:    int8(m.Code),
		//Count:   1,
	}

	if cli.enableBackup {
		cli.storeCli.WriteMetris(m)
	}
	atomic.AddInt32(&cli.moniterBuffSize, 1)
	cli.buffer <- metrics
	return nil
}

// Start
// 启动上传线程
func (cli *YoKaMan) Start() error {

	reportid, err := cli.WebCli.StartTest(cli.projectid, "1")
	cli.projectid = uint(reportid)
	cli.testid = int32(reportid)
	if cli.enableBackup {
		cli.storeCli = NewStorage(cli.testid)
		cli.storeCli.WriteHeader()
	}

	//链接metricssvr， 连不上则报错
	err = cli.DataCli.ConnectSvr()
	if err != nil {
		zap.L().Error("connect metrics svr failed")
		return err
	}

	cli.DataCli.ThreadSend()
	if cli.enableBackup {
		cli.storeCli.ThreadWrite()
	}

	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				zap.L().Debug(fmt.Sprintf("收包缓冲区长度 %d", cli.moniterBuffSize))
				break
			}
		}
	}()

	metricsBodySize := binary.Size(NetMetrics{}) - binary.Size(Protohead{})
	now := time.Now()
	start := time.Now()
	go func() {
		for {
			select {
			case v, ok := <-cli.buffer:
				if !ok {
					fmt.Println("Request buffer closed")
					return
				}
				atomic.AddInt32(&cli.moniterBuffSize, ^int32(0))
				atomic.AddInt64(&cli.totalpkg, 1)
				if time.Since(now).Seconds() > 10 {
					zap.L().Debug(fmt.Sprintf("收包处理qps: %.2f", float64(cli.totalpkg)/time.Since(start).Seconds()))
					now = time.Now()
				}

				pkg2send := NetMetrics{
					Protohead: Protohead{
						Len:        int8(metricsBodySize),
						Protocolid: ProtoRequestMetrics,
					},
					Testid: cli.testid,
					//Nodeid:        cli.nodeid,
					NetReqMetrics: v,
				}
				cli.DataCli.UploadStatics(pkg2send)
				break

			default:
				time.Sleep(time.Millisecond) //长期无数据，释放cpu
				break
			}
		}
	}()
	return nil
}

// Stop 关闭停止测试
func (cli *YoKaMan) Stop() error {
	cli.CmdCli.StopTest(uint32(cli.testid))
	_, err := cli.WebCli.StopTest()
	if err != nil {
		return err
	}
	cli.DataCli.Exit()
	if cli.enableBackup {
		cli.storeCli.Close()

	}
	return nil
}
