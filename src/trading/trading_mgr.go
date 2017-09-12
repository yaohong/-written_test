package main
import (
	"sync"
	"fmt"
)

const (
	ot_borrow int32 = 0
	ot_repay int32 = 1
)

var gTradingMgr *TradingMgr = nil


type TradingMgr struct {
	mutex sync.Mutex
	allTradingUser map[int64] *TradingUser

}

//在main函数里调用
func InitGlobalTradingMgr() {
	if gTradingMgr  == nil {
		gTradingMgr = &TradingMgr {
			allTradingUser: make(map[int64] *TradingUser),
		}
	}
}


func GetTradingMgr() *TradingMgr {
	return gTradingMgr
}



func (self *TradingMgr)GetUser(userId int64) (*TradingUser, bool) {
	return self.allTradingUser[userId]
}



func (self *TradingMgr)CreateUser(initMoney int32) (int64, error) {
	if initMoney < 0 {
		return 0, CreateError("initMoney[%d] exeption", initMoney)
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

	self.mutex.Lock()
	self.allTradingUser[newId] = newUser
	self.mutex.Unlock()

	return newId, nil
}


func (self *TradingMgr)CreateBorrow(source int64, dest int64, money int32) error {
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
		insertBorrowSql := fmt.Sprintf("insert into borrow_data (`source_id`, `dest_id`, `amount`, `oper_type`) values(%d, %d, %d, %d)", source, dest, money, ot_borrow)
		updateSql1 := fmt.Sprintf("update trading_user set money=money+%d where id=%d", money, source)
		updateSql2 := fmt.Sprintf("update trading_user set money=money-%d where id=%d", money, dest)
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
		insertBorrowSql := fmt.Sprintf("insert into borrow_data (`source_id`, `dest_id`, `amount`, `oper_type`) values(%d, %d, %d, %d)", source, dest, money, ot_repay)
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