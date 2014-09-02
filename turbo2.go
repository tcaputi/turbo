package main

import(
	"github.com/logmein3546/websocketserver"
	"fmt"
	"net/http"
	//"encoding/json"
)

func onReceive(connid string, data string) string{
	fmt.Println(connid, data)
	return data
}

func homeHandler(res http.ResponseWriter, req *http.Request) {
	http.ServeFile(res, req, "C:\\Development\\Projects\\go\\src\\github.com\\logmein3546\\turbo2\\turbotest.html")
}

func main(){
	fmt.Println("starting...")
	http.HandleFunc("/", homeHandler)
	websocketserver.RegisterReceiveHandler(onReceive)
	websocketserver.Initialize("/ws")
	if err := http.ListenAndServe(":4000", nil); err != nil {
		fmt.Println("ListenAndServe:", err)
    }
}
