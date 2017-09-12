package main
import (
	"net/http"
	"log"
)




func main() {
	InitGlobalDbMgr()
	GetDbMgr().Init("root:123456@tcp(127.0.0.1:3306)/trading?charset=utf8")
	InitGlobalTradingMgr()


	http.HandleFunc("/createUser", http_CreateUser)
	http.HandleFunc("/createBorrow", http_CreateBorrow)
	http.HandleFunc("/createRepay", http_CreateRepay)
	http.HandleFunc("/getUserInfo", http_GetUserInfo)
	http.HandleFunc("/viewUserRelation", http_ViewUserRelation)

	server := &http.Server{Addr: ":8889", Handler: nil}
	log.Printf("start server success.\n")
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
