// 用于数据通讯，性能指标等高性能通讯

package yokaman

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"go.uber.org/zap"
	"math"
	"net"
	"sync/atomic"
	"time"
)

// 最大延迟发送时间
const MaxLatency = 100 //100毫秒

const MTU = 1460

// 最大延迟发送包量
var MaxLatencyPkg uint32

func init() {
	MaxLatencyPkg = uint32(math.Floor(float64(MTU / binary.Size((NetMetrics{})))))
}

// 待发送队列最大值
const MAX_CHAN_SIZE = (1024 * 1024)

type MetricsNetCli struct {
	conn         net.Conn      //tcp链接句柄
	metricaddr   string        //metric svr 地址
	buf          *bytes.Buffer //发送用的字节缓冲区
	lastPkgIn    time.Time     //最近一次收到待发送业务包的时间戳，当时间超过MaxDelay立即发送
	stackedPkg   uint32        //堆积的业务包量
	totalSendPkg uint32        //已发送的数据包
	//IntervalSendPkg uint32 //已发送的数据包
	metrics2send  chan NetMetrics //待发送缓冲区
	pkg2SendQSize uint64          //待发送的网络包
}

func NewMetricsNetCli() *MetricsNetCli {
	m := MetricsNetCli{}
	m.metricaddr = "127.0.0.1"
	m.lastPkgIn = time.Unix(32500886400, 0) //默认设置为3000年，仅当收到第一个待发送业务包时，开始计时
	m.metrics2send = make(chan NetMetrics, MAX_CHAN_SIZE)
	m.buf = new(bytes.Buffer)
	m.stackedPkg = 0
	//defer m.conn.Close()
	return &m
}

func (m *MetricsNetCli) ConnectSvr() error {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:2380", m.metricaddr), 5*time.Second)
	if err != nil {
		zap.L().Error(fmt.Sprintf("connect metrics svr (%s) failed", m.metricaddr))
		return err
	}
	m.conn = conn
	return nil
}

func (m *MetricsNetCli) reset() {
	m.stackedPkg = 0
	m.lastPkgIn = time.Unix(32500886400, 0) //默认设置一个极大的年份
}

func (m *MetricsNetCli) realSend() {
	_, err := m.conn.Write(m.buf.Bytes()) //这里可能卡一下
	if err != nil {
		fmt.Println(err)
		return
	}
	m.buf.Reset()
}
func (m *MetricsNetCli) sendPkg(v NetMetrics) {
	err := binary.Write(m.buf, binary.LittleEndian, v)
	if err != nil {
		fmt.Println(err)
	}
	m.stackedPkg++
	m.totalSendPkg++
	if m.stackedPkg == 1 {
		m.lastPkgIn = time.Now()
	}
	if m.stackedPkg >= MaxLatencyPkg<<2 {
		m.realSend()
		m.reset()
	}
}

// 客户端循环发包，要注意buf 和socket 是否是线程安全的 。 待测
func (m *MetricsNetCli) ThreadSend( /*metrics2send chan []NetReqMetrics */ ) {
	//
	now := time.Now()
	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				fmt.Printf("【%v】 stackedPkg = %d, send buff size %d, totoal send %d , qps = %.2f\n", time.Now().Format("2006-01-02 15:04:05.00"),
					m.stackedPkg, m.pkg2SendQSize, m.totalSendPkg, float64(m.totalSendPkg)/time.Since(now).Seconds())
				break
				// 在这里写入打印数据的逻辑
			}
		}
	}()

	go func() {
		for {
			select {
			case v, ok := <-m.metrics2send:
				if !ok {

					fmt.Println("net message channel closed")
					return
				}
				atomic.AddUint64(&m.pkg2SendQSize, ^uint64(0))
				m.sendPkg(v)
				break
			default:
				if time.Since(m.lastPkgIn) > time.Duration(time.Millisecond) {
					m.realSend()
					m.reset()
					//	fmt.Println("有数据，但是缓冲区未满，定时发送")

				} else {
					//	fmt.Println("长期无数据发送，sleep一下")
					time.Sleep(time.Millisecond) //长期无数据，释放cpu
				}
				break
			}
		}
	}()
}

func (m *MetricsNetCli) UploadStatics(metrics NetMetrics) {
	atomic.AddUint64(&m.pkg2SendQSize, 1)
	m.metrics2send <- metrics
}

func (m *MetricsNetCli) Exit() {
	err := m.conn.Close()
	if err != nil {
		fmt.Printf("%s connect err, %s", m.conn.RemoteAddr().String(), err)
		return
	}
}
