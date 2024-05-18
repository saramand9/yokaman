package yokaman

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
