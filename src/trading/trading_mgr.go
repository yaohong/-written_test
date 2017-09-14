package main
import (
	"sync"
	"fmt"
)



var gTradingMgr *TradingMgr = nil


type TradingMgr struct {
	mutex sync.Mutex
	allTradingUser map[int64] *TradingUser
	notCompletedTrading map[int64] *TradingData
}

//在main函数里调用
func InitGlobalTradingMgr() {
	if gTradingMgr  == nil {
		gTradingMgr = &TradingMgr {
			allTradingUser: make(map[int64] *TradingUser),
			notCompletedTrading: make(map[int64] *TradingData),
		}
	}
}


func GetTradingMgr() *TradingMgr {
	return gTradingMgr
}


func (self *TradingMgr)loadTradingData(user *TradingUser, mode int32) error {
	var suffix string
	if mode == md_active {
		suffix = fmt.Sprintf("active_trading_data where `source_id`=%d", user.id)
	} else {
		suffix = fmt.Sprintf("passive_trading_data where `dest_id`=%d", user.id)
	}

	selectSql := fmt.Sprintf("select serial_number, source_id,target_id,amount,oper_type, state from %s", suffix)
	rows, err := GetDbMgr().GetDbConnect().Query(selectSql)
	if err != nil {
		return err
	}

	defer rows.Close()


	for rows.Next() {
		var serialNumber int64
		var sourceId int64
		var targetId int64
		var money int32
		var operType int32
		var state int32
		err = rows.Scan(&serialNumber, &sourceId, &targetId, &money, &operType, &state)
		if err != nil {
			return err
		}

		if operType != ot_borrow && operType != ot_repay {
			return CreateError("operType error ~p", operType)
		}

		if mode == md_active {
			//主角是发起方， 有可能没有完成
			if state == 0 {
				//没有完成的，由mgr继续发起交易
				self.notCompletedTrading[serialNumber] = NewTradomgData(serialNumber, sourceId, targetId, money, operType)
				user.lock = true
			} else {
				if operType == ot_borrow {
					//我向dest借钱
					user.addLoanFromDb(targetId, money)
				} else {
					//我向dest还钱
					user.addGiveMoneyToOtherFromDb(targetId, money)
				}
			}
		} else {
			//被动接受state肯定为1(插入时state就为1)

			if operType == ot_borrow {
				//source给我借钱
				user.addBorrowFromDb(sourceId, money)
			} else {
				//source给我还钱
				user.addGaveMeMoneyFromDb(sourceId, money)
			}
		}
	}

	return nil
}

func (self *TradingMgr)LoadAllByDb() error{

	{
		//加载玩家数据
		loadUserSql := fmt.Sprintf("select id, money from trading_user")
		rows1, err := GetDbMgr().GetDbConnect().Query(loadUserSql)
		if err != nil {
			return err
		}
		defer rows1.Close()

		for rows1.Next() {
			var id int64
			var money int32
			err = rows1.Scan(&id, &money)
			if err != nil {
				return err
			}

			self.allTradingUser[id]  = NewTradingUser(id, money)
		}

	}

	{
		for _, v := range self.allTradingUser {
			//加载交易数据
			self.loadTradingData(v, md_active)
			self.loadTradingData(v, md_passive)
			go v.Loop()
		}

	}
	return nil
}



func (self *TradingMgr)GetUser(userId int64) (*TradingUser, bool) {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	v, ok := self.allTradingUser[userId]
	return v, ok
}



func (self *TradingMgr)CreateUser(initMoney int32) (int64, error) {
	if initMoney < 0 {
		return 0, CreateError("initMoney[%d] exception", initMoney)
	}

	insertSql := fmt.Sprintf("insert into trading_user (`money`) values(%d)", initMoney)
	result, err := GetDbMgr().GetDbConnect().Exec(insertSql)
	if err != nil {
		return 0, err
	}

	newId, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	newUser := NewTradingUser(newId, initMoney)
	go newUser.Loop()

	self.mutex.Lock()
	self.allTradingUser[newId] = newUser
	self.mutex.Unlock()

	return newId, nil
}


func (self *TradingMgr)GenerateSerialNumber() (int64, error) {
	insertSql := fmt.Sprintf("insert into serial_number () values()")
	result, err := GetDbMgr().GetDbConnect().Exec(insertSql)
	if err != nil {
		return 0, err
	}

	newId, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return newId, nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////


func (self *TradingMgr)CreateBorrow(source int64, dest int64, money int32) error {
	if money < 0 {
		return CreateError("create borrow error, money[%d] exception", money)
	}
	//source向dest借钱 数额 [money]
	self.mutex.Lock()
	defer self.mutex.Unlock()

	sourceUser ,ok := self.allTradingUser[source]
	if !ok {
		return CreateError("create borrow error, find source[%d] faild", source)
	}

	destUser ,ok := self.allTradingUser[dest]
	if !ok {
		return CreateError("create borrow error, find dest[%d] faild", dest)
	}


	//检测source的钱是否够
	if sourceUser.GetCurrentMoney() < money {
		return CreateError("create borrow failed, source currentMoney[%d] borrowMoney[%d]", sourceUser.GetCurrentMoney(), money)
	}

	//更新数据库
	{
		insertBorrowSql := fmt.Sprintf("insert into trading_data (`source_id`, `dest_id`, `amount`, `oper_type`) values(%d, %d, %d, %d)", source, dest, money, ot_borrow)
		updateSql1 := fmt.Sprintf("update trading_user set money=money-%d where id=%d", money, source)
		updateSql2 := fmt.Sprintf("update trading_user set money=money+%d where id=%d", money, dest)
		connect := GetDbMgr().GetDbConnect()
		Tx, err := connect.Begin()
		if err != nil {
			return err
		}

		_, err = Tx.Exec(insertBorrowSql)
		if err != nil {
			Tx.Rollback()
			return err
		}

		_, err = Tx.Exec(updateSql1)
		if err != nil {
			Tx.Rollback()
			return err
		}

		_, err = Tx.Exec(updateSql2)
		if err != nil {
			Tx.Rollback()
			return err
		}

		err = Tx.Commit()
		if err != nil {
			return err
		}
	}

	//没有问题了
	sourceUser.AddLoan(dest, money)
	destUser.AddBorrow(source, money)



	return nil
}

//添加一笔还钱
func (self *TradingMgr)CreateRepay(source int64, dest int64, money int32) error {
	//source给dest还钱 数额 [money]
	if money < 0 {
		return CreateError("create repay error, money[%d] exception", money)
	}
	self.mutex.Lock()
	defer self.mutex.Unlock()

	sourceUser ,ok := self.allTradingUser[source]
	if !ok {
		return CreateError("create repay error, find source[%d] faild", source)
	}

	destUser ,ok := self.allTradingUser[dest]
	if !ok {
		return CreateError("create repay error, find dest[%d] faild", dest)
	}


	//source是否向dest借過這麽多钱
	sourceCurrentMoney := sourceUser.GetCurrentMoney()
	if sourceCurrentMoney < money {
		return CreateError("create repay error, source[%d] currentMoney[%d] repayMoney[%d]", source, sourceCurrentMoney, money)
	}
	borrowMoney, ok := sourceUser.CheckBorrowUser(dest)
	if borrowMoney < money {
		//沒有向dest借过这么多钱
		return CreateError("create borrow failed, borrow[%d=>%d %d] replyMoney[%d]", source, dest, borrowMoney, money)
	}

	loanMoney, ok := destUser.CheckLoanUser(source)
	if loanMoney < money {
		//没有给source借出过这么多钱
		return CreateError("create borrow failed, loan[%d=>%d %d] replyMoney[%d]", dest, source, loanMoney, money)
	}

	{
		insertBorrowSql := fmt.Sprintf("insert into trading_data (`source_id`, `dest_id`, `amount`, `oper_type`) values(%d, %d, %d, %d)", source, dest, money, ot_repay)
		updateSql1 := fmt.Sprintf("update trading_user set money=money-%d where id=%d", money, source)
		updateSql2 := fmt.Sprintf("update trading_user set money=money+%d where id=%d", money, dest)
		connect := GetDbMgr().GetDbConnect()
		Tx, err := connect.Begin()
		if err != nil {
			return err
		}

		_, err = Tx.Exec(insertBorrowSql)
		if err != nil {
			Tx.Rollback()
			return err
		}

		_, err = Tx.Exec(updateSql1)
		if err != nil {
			Tx.Rollback()
			return err
		}

		_, err = Tx.Exec(updateSql2)
		if err != nil {
			Tx.Rollback()
			return err
		}

		err = Tx.Commit()
		if err != nil {
			return err
		}
	}

	sourceUser.GiveMoneyToOther(dest, money)
	destUser.GaveMeMoney(source, money)

	return nil
}