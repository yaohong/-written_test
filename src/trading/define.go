package main






const (
	ecode_success int32 = 0
	ecode_timeout int32 = 1001							//超时
	ecode_lock int32 = 1002								//被锁定了
	ecode_generate_serial_number_failed int32 = 1003	//生成流水号失败
	ecode_db_error int32 = 1004							//数据库错误

	ecode_money_not_enough int32 = 1005		//钱不够
	ecode_20001 = 20001						//没有给对方借过这么多钱
	ecode_20002 = 20002						//对方没有给我借过钱
)

const (
	normal_timeout = 3 * 1000
)

const (
	ot_borrow int32 = 0
	ot_repay int32 = 1
)

const (
	md_active int32 = 0				//主动
	md_passive int32 = 1			//被动
)