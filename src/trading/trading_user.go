package main
import (
	"log"
)


const (

)


type TradingUser struct {
	id int64
	currentMoney int32				//当前拥有的钱
	allBorrow map[int64] int32		//记录向那些ID借过钱
	allLoan map[int64] int32        //记录向那些ID借出过钱
}


func NewTradingUser(id int64, initMoney int32) *TradingUser {
	return &TradingUser{
		id: id,
		currentMoney: initMoney,
		allBorrow: make(map[int64] int32),
		allLoan: make(map[int64] int32),
	}
}


func (self *TradingUser)TotalBorrow() int32 {
	var totalValue int32
	for _, v:= range self.allBorrow {
		totalValue += v
	}

	return totalValue
}

func (self *TradingUser)TotalLoan() int32 {
	var totalValue int32
	for _, v:= range self.allLoan {
		totalValue += v
	}

	return totalValue
}

//添家一笔借入从DB
func (self *TradingUser)AddBorrowFromDb(destId int64, money int32) {
	old, ok := self.allBorrow[destId]
	if ok {
		self.allBorrow[destId] = old + money
	} else {
		self.allBorrow[destId] = money
	}
}

//添加一笔借出从DB
func (self *TradingUser)AddLoanFromDb (destId int64, money int32)  {

	old, ok := self.allLoan[destId]
	if ok {
		self.allLoan[destId] = old + money
	} else {
		self.allLoan[destId] = money
	}
}

//别人给自己还钱从DB
func (self *TradingUser)AddGaveMeMoneyFromDb(destId int64, money int32) {
	loanMoney, ok := self.allLoan[destId]
	if !ok {
		self.allLoan[destId] = 0 - money
	} else {
		self.allLoan[destId] = loanMoney - money
	}

}

//给别人还钱
func (self *TradingUser)AddGiveMoneyToOtherFromDb(destIn int64, money int32) {
	borrowMoney, ok := self.allBorrow[destIn]
	if !ok {
		self.allBorrow[destIn] = 0 - money
	} else {
		self.allBorrow[destIn] = borrowMoney - money
	}

}

////////////////////////////////////////////////////////////////////////////////////////////

//添加一笔借入
func (self *TradingUser)AddBorrow(destId int64, money int32) {
	old, ok := self.allBorrow[destId]
	if ok {
		self.allBorrow[destId] = old + money
	} else {
		self.allBorrow[destId] = money
	}
	self.currentMoney += money
}

//添加一笔借出
func (self *TradingUser)AddLoan (destId int64, money int32)  {
	if money > self.currentMoney {
		//钱不够了不能借了,外部检测过了
		log.Fatalln("amount[%d] > self.currentMoney[%d]", money, self.currentMoney)
		return
	}

	old, ok := self.allLoan[destId]
	if ok {
		self.allLoan[destId] = old + money
	} else {
		self.allLoan[destId] = money
	}
	self.currentMoney -= money

}

//别人给自己还钱
func (self *TradingUser)GaveMeMoney(destId int64, money int32) {
	loanMoney, ok := self.allLoan[destId]
	if !ok {
		log.Fatalln("id[%d] find loan[%d] failed", self.id, destId)
		return
	}

	if loanMoney < money {
		//还的钱大于我借出的钱
		log.Fatalln("loanMoney[%d] < money[%d]", loanMoney, money)
		return
	}

	self.allLoan[destId] = loanMoney - money
	self.currentMoney += money

}

//给别人还钱
func (self *TradingUser)GiveMoneyToOther(destIn int64, money int32) {
	borrowMOney, ok := self.allBorrow[destIn]
	if !ok {
		log.Fatalln("id[%d] find loan[%d] failed", self.id, destIn)
		return
	}

	if money > self.currentMoney {
		//钱不够了不能借了,外部检测过了
		log.Fatalln("amount[%d] > self.currentMoney[%d]", money, self.currentMoney)
		return
	}

	if borrowMOney < money {
		//还的钱大于我借出的钱
		log.Fatalln("borrowMOney[%d] < money[%d]", borrowMOney, money)
		return
	}



	self.allBorrow[destIn] = borrowMOney - money
	self.currentMoney -= money

}


func (self *TradingUser)GetCurrentMoney() int32 {
	return self.currentMoney
}


//检测给自己借过钱的人
func (self *TradingUser)CheckBorrowUser(id int64) (int32, bool) {
	v, ok := self.allBorrow[id]
	return v, ok
}

//检测向我借过钱的人
func (self *TradingUser)CheckLoanUser(id int64) (int32, bool) {
	v, ok := self.allLoan[id]
	return v, ok
}



