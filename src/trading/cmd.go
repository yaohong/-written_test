package main





//主动命令

//给别人借钱
type Cmd_AddLoan struct {
	targetId int64
	money int32
	replyCh chan int32
}

func NewCmdAddLoan(targetId int64, money int32) *Cmd_AddLoan {
	return &Cmd_AddLoan {
		targetId: targetId,
		money: money,
		replyCh: make(chan int32),
	}
}

//给别人还钱
type Cmd_GiveMoneyToOther struct {
	targetId int64
	money int32
	replyCh chan int32
}


func NewCmdGiveMoneyToOther(targetId int64, money int32) *Cmd_GiveMoneyToOther{
	return &Cmd_GiveMoneyToOther {
		targetId: targetId,
		money: money,
		replyCh: make(chan int32),
	}
}




//被动命令
//被人给我借钱
type Cmd_AddBorrow struct {
	serialNumber int64
	sourceId int64
	money int32
	replyCh chan int32
}

func NewCmdAddBorrow(serialNumber int64, sourceId int64, money int32) * Cmd_AddBorrow {
	return &Cmd_AddBorrow{
		serialNumber: serialNumber,
		sourceId: sourceId,
		money: money,
		replyCh: make(chan int32),
	}
}

//别人给我还钱
type Cmd_GaveMeMoney struct {
	serialNumber int64
	sourceId int64
	money int32
	replyCh chan int32
}

func NewCmdGaveMeMoney(serialNumber int64, sourceId int64, money int32) * Cmd_GaveMeMoney{
	return &Cmd_GaveMeMoney{
		serialNumber: serialNumber,
		sourceId: sourceId,
		money: money,
		replyCh: make(chan int32),
	}
}