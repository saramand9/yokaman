// yoka机器人事务统计客户端， 用于和服务端通讯
package yokaman

import (
	"fmt"
	"sync"
	"time"
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

	metricsNetCli *MetricsNetCli
	totalpkg      int64
	totalpkgrecv  int64
}

var instance *YoKaMan
var once sync.Once

func YoKaManCli() *YoKaMan {
	once.Do(func() {
		instance = &YoKaMan{
			cache:         NewCache(),
			metricsNetCli: NewMetricsNetCli(),
			totalpkg:      0,
			totalpkgrecv:  0,
		}
	})
	return instance
}

// 如果执行多次测试，需要提前设置好测试id，以便区分不同测试场景
// testid:  每次测试的唯一id
// nodeid:  每个机器人的id
func (m *YoKaMan) SetTestInfo(testid uint, nodeid uint) {
	m.testid = int8(testid) //暂时不会超过256
	m.nodeid = int8(nodeid) //暂时不会超过256
}

// 预留，用于在性能不足时，做数据聚合
// second:  表示多少秒内的数据做一次聚合。 注意如果做了数据聚合，StatMetrics中的robotid可能就不准确了
func (m *YoKaMan) SetAggregation(second uint) {

}

// 用于统计事务qps，成功率等
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
		id, err = Register(m.Trans, uint32(cli.testid))
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
	Metrics2send <- metrics
	return nil
}

// 启动上传线程
func (cli *YoKaMan) Start() error {

	go cli.metricsNetCli.Run()
	metrictBuffSize := GetNetStructSize(RecodeMetrics{}) - GetNetStructSize(Protohead{})

	for v := range Metrics2send {
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
		cli.metricsNetCli.UploadStatics(pkg2send)
	}
	return nil
}

// 关闭上传线程
func (cli *YoKaMan) Stop() error {
	cli.metricsNetCli.Exit()
	return nil
}
