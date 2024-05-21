// Package yokaman
// yoka机器人事务统计客户端， 用于和服务端通讯
package yokaman

import (
	"fmt"
	"sync"
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
	testid int8 //测试id
	nodeid int8 //机器

	cache *Cache

	DataCli      *MetricsNetCli
	CmdCli       *MetricsCmdCli
	storeCli     *Localstorage
	totalpkg     int64
	totalpkgrecv int64
	buffer       chan RequestMetrics
	enableBackup bool
}

var instance *YoKaMan
var once sync.Once

func YoKaManCli() *YoKaMan {
	once.Do(func() {
		instance = &YoKaMan{
			cache:   NewCache(),
			DataCli: NewMetricsNetCli(),
			CmdCli:  NewMetricsCmdCli(),
			//storeCli: 	  NewStorage(),
			totalpkg:     0,
			totalpkgrecv: 0,
			enableBackup: true,
			buffer:       make(chan RequestMetrics, 1024*1024),
		}
	})
	return instance
}

// SetTestInfo
// 如果执行多次测试，需要提前设置好测试id，以便区分不同测试场景
// testid:  每次测试的唯一id
// nodeid:  每个机器人的id
func (m *YoKaMan) SetTestInfo(testid uint, nodeid ...uint) {
	m.testid = int8(testid) //暂时不会超过256
	if len(nodeid) > 0 {
		m.nodeid = int8(nodeid[0]) //暂时不会超过256
	}

	m.storeCli = NewStorage(int(testid))
	//m.storeCli.WriteMetrics()
	m.storeCli.WriteHeader()
	//fmt.Println(m.nodeid)
}

func (m *YoKaMan) SetMetricsSvrAddr(addr string) {
	m.DataCli.metricaddr = addr
	m.CmdCli.metricaddr = addr
}

// SetAggregation 预留，用于在性能不足时，做数据聚合
// second:  表示多少秒内的数据做一次聚合。 注意如果做了数据聚合，StatMetrics中的robotid可能就不准确了
func (m *YoKaMan) SetAggregation(second uint) {

}

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

	metrics := RequestMetrics{
		Transid: int8(id),
		Start:   m.Reqtime,
		End:     m.Resptime,
		Code:    int8(m.Code),
		Count:   1,
	}

	if cli.enableBackup {
		cli.storeCli.WriteMetris(m)
	}

	cli.buffer <- metrics
	return nil
}

// Start
// 启动上传线程
func (cli *YoKaMan) Start() error {
	//链接metricssvr， 连不上则报错
	err := cli.DataCli.ConnectSvr()
	if err != nil {
		fmt.Println("connect server err ", err)
		return err
	}

	cli.DataCli.ThreadSend()
	if cli.enableBackup {
		cli.storeCli.ThreadWrite()
	}

	metrictBuffSize := GetNetStructSize(RecodeMetrics{}) - GetNetStructSize(Protohead{})
	go func() {
		for {
			select {
			case v, ok := <-cli.buffer:
				if !ok {
					fmt.Println("Request buffer closed")
					return
				}
				cli.totalpkg++
				if cli.totalpkg%1000000 == 0 {
					fmt.Printf("【%v】hanlde %d metrics\n", time.Now().Format("2006-01-02 15:04:05.00"), cli.totalpkg)
				}

				pkg2send := RecodeMetrics{
					Protohead: Protohead{
						Len:        int8(metrictBuffSize),
						Protocolid: ProtoRequestMetrics,
					},
					Testid:         cli.testid,
					Nodeid:         cli.nodeid,
					RequestMetrics: v,
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
	//cli.DataCli.Exit()
	//cli.storeCli.Close()
	return nil
}
