// 用于数据通讯，性能指标等高性能通讯

package yokaman

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

const MaxDelay = 100 //ms
// 协议序号
const ProtoRequestMetrics = 1

type Protohead struct {
	Len        int8 //
	Protocolid int8 //协议id
}

type RequestMetrics struct {
	Transid int8 //后面扩展
	Start   int64
	End     int64 //微秒
	Code    int8
	Count   int16
}

type RecodeMetrics struct {
	Protohead
	Testid int8 //测试id
	Nodeid int8 //机器人id
	RequestMetrics
}

var Metrics2send chan RequestMetrics

type MetricsNetCli struct {
	conn net.Conn

	buf        *bytes.Buffer
	lastPkgIn  time.Time
	stackedPkg uint32

	totalSendPkg    uint32 //已发送的数据包
	IntervalSendPkg uint32 //已发送的数据包

	metrics_chan chan RecodeMetrics
	// 创建消费端串行处理的Lockfree
	//ring lockfree.Lockfree
}

func NewMetricsNetCli() *MetricsNetCli {
	m := MetricsNetCli{}
	dialer := net.Dialer{}
	c, err := dialer.Dial("tcp", "127.0.0.1:2380")
	if err != nil {
		fmt.Printf("connect 10.225.21.227 failed\n")
		return nil
	}
	m.conn = c
	m.lastPkgIn = time.Unix(32500886400, 0) //默认设置一个极大的年份

	m.metrics_chan = make(chan RecodeMetrics, 1024*1024)

	m.buf = new(bytes.Buffer)
	m.stackedPkg = 0
	//defer m.conn.Close()
	return &m
}

func init() {
	Metrics2send = make(chan RequestMetrics, 1024*1024)
}

func GetNetStructSize(data interface{}) int {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, data)
	return len(buf.Bytes())
}

// 这里的判断逻辑有2个，
// 1. 如果这个包已经收到超过100ms了，那么发送
// 2. 如果缓冲区已经满的差不多了，即发送. 这里需要根据并发来评估
func (m *MetricsNetCli) Run1() {
	/*ticker := time.NewTicker((MaxDelay / 2) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 执行需要重复执行的代码
			m.SendPackets()
		}
	}*/
}

// 客户端循环发包，要注意buf 和socket 是否是线程安全的 。 待测
func (m *MetricsNetCli) Run( /*metrics_chan chan []RequestMetrics */ ) {
	//
	//count := 0
	for true {
		select {
		case v, ok := <-m.metrics_chan:
			if !ok {
				fmt.Println("channel closed")
				return
			}
			err := binary.Write(m.buf, binary.LittleEndian, v)
			if err != nil {
				fmt.Println(err)
			}
			m.stackedPkg++
			m.totalSendPkg++

			//fmt.Println("send pkg %d", m.stackedPkg)
			if m.stackedPkg == 1 {
				m.lastPkgIn = time.Now()
			}
			if m.stackedPkg >= 60 {
				_, err := m.conn.Write(m.buf.Bytes())
				if err != nil {
					fmt.Println(err)
					return
				}
				m.stackedPkg = 0
				m.lastPkgIn = time.Unix(32500886400, 0) //默认设置一个极大的年份
				m.buf.Reset()
			}
			break
		default:
			//fmt.Println(m.lastPkgIn, m.stackedPkg)
			if time.Since(m.lastPkgIn) > time.Duration(time.Millisecond*MaxDelay) {
				m.lastPkgIn = time.Unix(32500886400, 0) //默认设置一个极大的年份
				_, err := m.conn.Write(m.buf.Bytes())
				if err != nil {
					fmt.Println(err)
					return
				}
				m.stackedPkg = 0
				m.buf.Reset()
			}
			break
		}

	}
}

func (m *MetricsNetCli) UploadStatics(metrics RecodeMetrics) {
	m.metrics_chan <- metrics
}

func (m *MetricsNetCli) Exit() {
	m.conn.Close()
}
