package main
import (
	"net/http"
	"log"
	"io/ioutil"
	"simplejson"
)


func errorHandle(re interface{}, w http.ResponseWriter) {
	newJs := simplejson.New()
	newJs.Set("state", 0)
	if err, ok := re.(error);  ok {
		newJs.Set("error", err.Error())
	} else {
		log.Printf("create user error, %v", re)
		newJs.Set("error", "system error")
	}
	byte, _ := newJs.Encode()
	w.Write(byte)
}


func http_CreateUser(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if re := recover(); re != nil {
			errorHandle(re, w)
		}
	}()

	if r.Method != "POST" {
		log.Printf("error Method %s\n", r.Method)
		return
	}

	body, _ := ioutil.ReadAll(r.Body)

	js, err := simplejson.NewJson(body)
	if err != nil {
		panic(CreateError("simplejson.NewJson faied, error = %s", err.Error()))
	}

	initMoneyJs := js.Get("init_money")
	initMoney, err := initMoneyJs.Int()
	if err != nil {
		panic(CreateError("initMoneyJs.Int faied, error = %s", err.Error()))
	}
	log.Printf("create_user init_money=%d\n", initMoney)
	userId, err := GetTradingMgr().CreateUser(int32(initMoney))
	if err != nil {
		panic(CreateError("CreateUser failed, error = %s", err.Error()))
	}
	log.Printf("create_user success user_id=%d\n", userId)
	successJson := simplejson.New()
	successJson.Set("state", 0)

	dataJson := simplejson.New()
	dataJson.Set("user_id", userId)

	successJson.Set("data", dataJson)

	successByte , _ := successJson.Encode()

	w.Write(successByte)
}


func http_CreateBorrow(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if re := recover(); re != nil {
			errorHandle(re, w)
		}
	}()

	if r.Method != "POST" {
		log.Printf("error Method %s\n", r.Method)
		return
	}

	body, _ := ioutil.ReadAll(r.Body)

	js, err := simplejson.NewJson(body)
	if err != nil {
		panic(CreateError("simplejson.NewJson faied, error = %s", err.Error()))
	}

	sourceIdJs := js.Get("source_id")
	intSourceId, err := sourceIdJs.Int()
	if err != nil {
		panic(CreateError("sourceIdJs.Int faied, error = %s", err.Error()))
	}

	destIdJs := js.Get("dest_id")
	intDestId, err := destIdJs.Int()
	if err != nil {
		panic(CreateError("destIdJs.Int faied, error = %s", err.Error()))
	}

	moneyJs := js.Get("money")
	intMoney, err := moneyJs.Int()
	if err != nil {
		panic(CreateError("moneyJs.Int faied, error = %s", err.Error()))
	}



	err = GetTradingMgr().CreateBorrow(int64(intSourceId), int64(intDestId), int32(intMoney))
	if err != nil {
		panic(CreateError("CreateBorrow faied, error = %s", err.Error()))
	}

	successJson := simplejson.New()
	successJson.Set("state", 0)

	successByte , _ := successJson.Encode()

	w.Write(successByte)
}

func http_CreateRepay(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if re := recover(); re != nil {
			errorHandle(re, w)
		}
	}()

	if r.Method != "POST" {
		log.Printf("error Method %s\n", r.Method)
		return
	}

	body, _ := ioutil.ReadAll(r.Body)

	js, err := simplejson.NewJson(body)
	if err != nil {
		panic(CreateError("simplejson.NewJson faied, error = %s", err.Error()))
	}

	sourceIdJs := js.Get("source_id")
	intSourceId, err := sourceIdJs.Int()
	if err != nil {
		panic(CreateError("sourceIdJs.Int faied, error = %s", err.Error()))
	}

	destIdJs := js.Get("dest_id")
	intDestId, err := destIdJs.Int()
	if err != nil {
		panic(CreateError("destIdJs.Int faied, error = %s", err.Error()))
	}

	moneyJs := js.Get("money")
	intMoney, err := moneyJs.Int()
	if err != nil {
		panic(CreateError("moneyJs.Int faied, error = %s", err.Error()))
	}



	err = GetTradingMgr().CreateRepay(int64(intSourceId), int64(intDestId), int32(intMoney))
	if err != nil {
		panic(CreateError("CreateRepay faied, error = %s", err.Error()))
	}

	successJson := simplejson.New()
	successJson.Set("state", 0)

	successByte , _ := successJson.Encode()

	w.Write(successByte)
}

func http_GetUserInfo(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if re := recover(); re != nil {
			errorHandle(re, w)
		}
	}()

	if r.Method != "POST" {
		log.Printf("error Method %s\n", r.Method)
		return
	}

	body, _ := ioutil.ReadAll(r.Body)

	js, err := simplejson.NewJson(body)
	if err != nil {
		panic(CreateError("simplejson.NewJson faied, error = %s", err.Error()))
	}

	sourceIdJs := js.Get("user_id")
	intUserId, err := sourceIdJs.Int()
	if err != nil {
		panic(CreateError("sourceIdJs.Int %v faied, error = %s", sourceIdJs, err.Error()))
	}


	user, ok := GetTradingMgr().GetUser(int64(intUserId))
	if !ok {
		panic(CreateError("user_id[%d] not exist", intUserId))
	}

	currentMoney := user.GetCurrentMoney()			//当前剩下的千
	totalBorrow := user.TotalBorrow()				//一共借的钱
	totalLoan := user.TotalLoan()					//一共借出去的钱

	successJson := simplejson.New()
	successJson.Set("state", 0)

	dataJson := simplejson.New()
	dataJson.Set("current_money", currentMoney)
	dataJson.Set("total_borrow", totalBorrow)
	dataJson.Set("total_loan", totalLoan)

	successJson.Set("data", dataJson)

	successByte , _ := successJson.Encode()

	w.Write(successByte)
}

func http_ViewUserRelation(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if re := recover(); re != nil {
			errorHandle(re, w)
		}
	}()

	if r.Method != "POST" {
		log.Printf("error Method %s\n", r.Method)
		return
	}

	body, _ := ioutil.ReadAll(r.Body)

	js, err := simplejson.NewJson(body)
	if err != nil {
		panic(CreateError("simplejson.NewJson faied, error = %s", err.Error()))
	}

	sourceIdJs := js.Get("source_id")
	intSourceId, err := sourceIdJs.Int()
	if err != nil {
		panic(CreateError("sourceIdJs.Int faied, error = %s", err.Error()))
	}

	destIdJs := js.Get("dest_id")
	intDestId, err := destIdJs.Int()
	if err != nil {
		panic(CreateError("destIdJs.Int faied, error = %s", err.Error()))
	}


	//查看source给dest借了多少钱，dest给source借了多少钱
	sourceUser, ok := GetTradingMgr().GetUser(int64(intSourceId))
	if !ok {
		panic(CreateError("source user_id[%d] not exist", intSourceId))
	}

	_, ok = GetTradingMgr().GetUser(int64(intDestId))
	if !ok {
		panic(CreateError("dest_id[%d] not exist", intDestId))
	}

	//看dest是否给自己借过钱
	borrowValue, ok := sourceUser.CheckBorrowUser(int64(intDestId))
	if !ok {
		//没有给自己借过钱
		borrowValue = 0
	}

	//检测我是否给dest借过钱
	loanValue, ok := sourceUser.CheckLoanUser(int64(intDestId))
	if !ok {
		loanValue = 0
	}

	successJson := simplejson.New()
	successJson.Set("state", 0)

	dataJson := simplejson.New()
	dataJson.Set("borrow_value", borrowValue)
	dataJson.Set("loan_value", loanValue)

	successJson.Set("data", dataJson)

	successByte , _ := successJson.Encode()

	w.Write(successByte)
}

