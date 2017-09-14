package main
import (
	"log"
	"time"
	"fmt"
)




type TradingUser struct {
	id int64
	currentMoney int32				//当前拥有的钱
	allBorrow map[int64] int32		//记录向那些ID借过钱
	allLoan map[int64] int32        //记录向那些ID借出过钱
	ch chan interface{}
	lock bool
}


func NewTradingUser(id int64, initMoney int32) *TradingUser {
	return &TradingUser{
		id: id,
		currentMoney: initMoney,
		allBorrow: make(map[int64] int32),
		allLoan: make(map[int64] int32),
		ch : make(chan interface{}),
		lock : false,
	}
}


func (self *TradingUser)Loop() {
	for {
		select {
		case req := <- self.ch:
			switch v := req.(type)  {
			case *Cmd_AddLoan:
				self.handle_AddLoan(v)
			case *Cmd_GiveMoneyToOther:
				self.handle_GiveMoneyToOther(v)
			case *Cmd_AddBorrow:
				self.handle_AddBorrow(v)
			case *Cmd_GaveMeMoney:
				self.handle_GaveMeMoney(v)
			}
		}
	}
}

func (self *TradingUser)updateSerialNumberState(serialNumber int64) error {
	updateSql := fmt.Sprintf("update active_trading_data set state=1 where serial_number=%d", serialNumber)
	_, err := GetDbMgr().GetDbConnect().Exec(updateSql)
	if err != nil {
		return err
	}

	return nil
}

func (self *TradingUser)CheckSerialNumber(serialNumber int64) (bool ,error) {
	selectSql := fmt.Sprintf("select source_id from passive_trading_data where `serial_number`=%d", serialNumber)
	rows, err := GetDbMgr().GetDbConnect().Query(selectSql)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	if rows.Next() {
		return true, nil
	} else {
		return false, nil
	}
}


func (self *TradingUser)InsertActiveTradingData(serialNumber int64, target_id int64, amount int32, oper_type int32, state int32) error {
	insertSql := fmt.Sprintf(
		"replace into active_trading_data (`serial_number`, `source_id`, `dest_id`, `amount`, `oper_type`, `state`) values(%d, %d, %d, %d, %d, %d)",
		serialNumber,
		self.id,
		target_id,
		amount,
		oper_type,
		state)
	_, err := GetDbMgr().GetDbConnect().Exec(insertSql)
	if err != nil {
		return err
	}

	return nil
}

func (self *TradingUser)InsertPassiveTradingData(serialNumber int64, source_id int64, amount int32, oper_type int32) error {
	insertSql := fmt.Sprintf(
		"replace into passive_trading_data (`serial_number`, `source_id`, `dest_id`, `amount`, `oper_type`, `state`) values(%d, %d, %d, %d, %d, 1)",
		serialNumber,
		source_id,
		self.id,
		amount,
		ot_borrow)

	_, err := GetDbMgr().GetDbConnect().Exec(insertSql)
	if err != nil {
		return err
	}

	return nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func (self *TradingUser)handle_AddLoan(v *Cmd_AddLoan) {
	//我给对方借钱
	if self.lock {
		v.replyCh <- ecode_lock
		return
	}
	if self.currentMoney < v.money {
		log.Printf("handle_AddLoan, user_id[%d] currentMoney[%d] loanMoney[%d]\n", self.id, self.currentMoney, v.money)
		v.replyCh <- ecode_money_not_enough
		return
	}

	//获取流水号
	newSerialNumber, err := GetTradingMgr().GenerateSerialNumber()
	if err != nil {
		log.Printf("generate serialNumber failed, %s\n", err.Error())
		v.replyCh <- ecode_generate_serial_number_failed
		return
	}

	{
		//更新缓存
		self.currentMoney -= v.money
		old, ok := self.allLoan[v.targetId]
		if ok {
			self.allLoan[v.targetId] = old + v.money
		} else {
			self.allLoan[v.targetId] = v.money
		}
	}

	err = self.InsertActiveTradingData(newSerialNumber, v.targetId, v.money, ot_borrow, 0)
	if err != nil {
		//失败了?
		//真失败?检查线路故障，服务器重启FIX就好了
		//假失败?mysql成功处理请求,应答出错了, 缓存里是最新的数据，并且每次操作都是做过前置检测,在服务器重启之后会由tardingmgr尝试完成订单
		v.replyCh <- ecode_db_error
		return
	}

	//操作DB
	replyCode := ecode_success
	//看对方是在本机还是跨机器
	targetUser ,ok := GetTradingMgr().GetUser(v.targetId)
	if !ok {
		//这里就写伪代码了
		//根据ID或者tagetId所在的地址
		//发送一个http请求
		//replyCode = rpc.call(ip, targetUser.AddBorrow(newSerialNumber, self.id, v.money)
	} else {
		replyCode = targetUser.AddBorrow(newSerialNumber, self.id, v.money)
		if replyCode == ecode_success {
			//成功了,在数据库更新状态
			err := self.updateSerialNumberState(newSerialNumber)
			if err != nil {
				//更新流水状态失败，依然重试
				log.Printf("updateSerialNumberState failed, %s\n", err.Error())
				replyCode = ecode_db_error
			}
		}

	}


	if replyCode != ecode_success {
		//失败了,锁定
		self.lock = true

		//数据发给trading_mgr去重试操作,成功后解锁
		//伪代码
	}


	v.replyCh <- replyCode
}

func (self *TradingUser)handle_GiveMoneyToOther(v *Cmd_GiveMoneyToOther) {
	//给别人还钱
	if self.lock {
		v.replyCh <- ecode_lock
		return
	}

	//目标是否给自己借过钱
	borrowMoney, ok := self.allBorrow[v.targetId]
	if !ok || borrowMoney < v.money {
		//对方没有给自己借过钱
		v.replyCh <- ecode_20002
		return
	}

	//获取流水号
	newSerialNumber, err := GetTradingMgr().GenerateSerialNumber()
	if err != nil {
		log.Printf("generate serialNumber failed, %s\n", err.Error())
		v.replyCh <- ecode_generate_serial_number_failed
		return
	}

	{
		//更新缓存
		self.currentMoney -= v.money
		self.allBorrow[v.targetId] = borrowMoney - v.money
	}

	err = self.InsertActiveTradingData(newSerialNumber, v.targetId, v.money, ot_repay, 0)
	if err != nil {
		//失败了?
		v.replyCh <- ecode_db_error
		return
	}

	//操作DB
	replyCode := ecode_success
	//看对方是在本机还是跨机器
	targetUser ,ok := GetTradingMgr().GetUser(v.targetId)
	if !ok {
		//这里就写伪代码了
		//根据ID或者tagetId所在的地址
		//发送一个http请求
		//replyCode = rpc.call(ip, targetUser.GaveMeMoney(newSerialNumber, self.id, v.money)
	} else {
		replyCode = targetUser.GaveMeMoney(newSerialNumber, self.id, v.money)
		if replyCode == ecode_success {
			//成功了,在数据库更新状态
			err := self.updateSerialNumberState(newSerialNumber)
			if err != nil {
				//更新流水状态失败，依然重试
				log.Printf("updateSerialNumberState failed, %s\n", err.Error())
				replyCode = ecode_db_error
			}
		}

	}


	if replyCode != ecode_success {
		//失败了,锁定
		self.lock = true

		//数据发给trading_mgr去重试操作,成功后解锁
		//伪代码
	}


	v.replyCh <- replyCode
}

func (self *TradingUser)handle_AddBorrow(v *Cmd_AddBorrow) {
	//别人给我借钱了
	//检测流水号是否处理了
	exist ,err := self.CheckSerialNumber(v.serialNumber)
	if err != nil {
		log.Printf("CheckSerialNumber failed, %s\n", err.Error())
		v.replyCh <- ecode_db_error
		return
	}

	if exist {
		//已经处理过了
		v.replyCh <- ecode_success
		return
	}

	//往DB插入一条数据
	err = self.InsertPassiveTradingData(v.serialNumber, v.sourceId, v.money, ot_borrow)
	if err != nil {
		log.Printf("InsertPassiveTradingData failed, %s\n", err.Error())
		v.replyCh <- ecode_db_error
		return
	}
	//更新缓存
	{
		self.currentMoney += v.money
		old, ok := self.allBorrow[v.sourceId]
		if ok {
			self.allBorrow[v.sourceId] = old + v.money
		} else {
			self.allBorrow[v.sourceId] = v.money
		}
	}

	v.replyCh <- ecode_success

}

func (self *TradingUser)handle_GaveMeMoney(v *Cmd_GaveMeMoney) {

}

////////////////////////////////////////////////////////////////////////////////////

func (self *TradingUser)totalBorrow() int32 {
	var totalValue int32
	for _, v:= range self.allBorrow {
		totalValue += v
	}

	return totalValue
}

func (self *TradingUser)totalLoan() int32 {
	var totalValue int32
	for _, v:= range self.allLoan {
		totalValue += v
	}

	return totalValue
}

//DB的初始化接口
//添家一笔借入从DB
func (self *TradingUser)addBorrowFromDb(destId int64, money int32) {
	old, ok := self.allBorrow[destId]
	if ok {
		self.allBorrow[destId] = old + money
	} else {
		self.allBorrow[destId] = money
	}

	self.currentMoney += money
}

//添加一笔借出从DB
func (self *TradingUser)addLoanFromDb (destId int64, money int32)  {

	old, ok := self.allLoan[destId]
	if ok {
		self.allLoan[destId] = old + money
	} else {
		self.allLoan[destId] = money
	}

	self.currentMoney -= money
}

//别人给自己还钱从DB
func (self *TradingUser)addGaveMeMoneyFromDb(destId int64, money int32) {
	loanMoney, ok := self.allLoan[destId]
	if !ok {
		self.allLoan[destId] = 0 - money
	} else {
		self.allLoan[destId] = loanMoney - money
	}

	self.currentMoney += money
}

//给别人还钱
func (self *TradingUser)addGiveMoneyToOtherFromDb(destIn int64, money int32) {
	borrowMoney, ok := self.allBorrow[destIn]
	if !ok {
		self.allBorrow[destIn] = 0 - money
	} else {
		self.allBorrow[destIn] = borrowMoney - money
	}

	self.currentMoney -= money
}

////////////////////////////////////////////////////////////////////////////////////////////
//外部可以调用的接口
//别人给自己借钱
func (self *TradingUser)AddBorrow(serialNumber int64, sourceId int64, money int32) int32{
	req := NewCmdAddBorrow(serialNumber, sourceId, money)
	select {
	case self.ch <- req:
		return <-req.replyCh
	case <- time.After(normal_timeout):
		//超时了
		return ecode_timeout
	}

}

//我给别人借钱
func (self *TradingUser)AddLoan (targetId int64, money int32) int32 {
	req := NewCmdAddLoan(targetId, money)

	select {
	case self.ch <- req:
		return <-req.replyCh
	case <- time.After(normal_timeout):
		//超时了
		return ecode_timeout
	}

}

//别人给我还钱
func (self *TradingUser)GaveMeMoney(serialNumber int64, sourceId int64, money int32) int32 {
	req := NewCmdGaveMeMoney(serialNumber, sourceId, money)

	select {
	case self.ch <- req:
		return <-req.replyCh
	case <- time.After(normal_timeout):
		//超时了
		return ecode_timeout
	}

}

//我给别人还钱
func (self *TradingUser)GiveMoneyToOther(targetId int64, money int32) int32 {
	req := NewCmdGiveMoneyToOther(targetId, money)
	select {
	case self.ch <- req:
		return <-req.replyCh
	case <- time.After(normal_timeout):
		//超时了
		return ecode_timeout
	}
}


//func (self *TradingUser)GetCurrentMoney() int32 {
//	return self.currentMoney
//}
//
//
////检测给自己借过钱的人
//func (self *TradingUser)CheckBorrowUser(id int64) (int32, bool) {
//	v, ok := self.allBorrow[id]
//	return v, ok
//}
//
////检测向我借过钱的人
//func (self *TradingUser)CheckLoanUser(id int64) (int32, bool) {
//	v, ok := self.allLoan[id]
//	return v, ok
//}



