// DBGJ macOS Agent · 桌面 MVP（自研协议客户端，不整包嵌入第三方远控）
// 对齐 plmnod：register / report / ws/agent · screen_* 
package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	server := flag.String("server", "https://plmnod.com", "monitor base URL")
	token := flag.String("token", "", "agent token（空则自动 register）")
	signKey := flag.String("signKey", "", "HMAC sign key（register 可回填）")
	flag.Parse()

	a := NewAgent(*server, *token, *signKey)
	if err := a.Start(); err != nil {
		log.Fatal(err)
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	a.Stop()
}
