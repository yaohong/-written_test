package main


//描述一个交易订单
type TradingData struct {
	serialNumber int64
	source int64
	target int64
	money int32
	operType int32
}


func NewTradomgData(serialNumber int64, source int64, target int64, money int32, operType int32) *TradingData {
	return &TradingData {
		serialNumber: serialNumber,
		source: source,
		target: target,
		money: money,
		operType: operType,
	}
}
