package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"card-game-client/client"
	"card-game-client/handler"
)

func main() {
	addr := "127.0.0.1:9527"
	//username := "test_user"
	//password := "test_pass"

	cli := client.NewTCPClient(addr)
	if err := cli.Connect(); err != nil {
		fmt.Printf("Failed to connect to server %s: %v\n", addr, err)
		os.Exit(1)
	}
	fmt.Printf("Connected to server %s successfully\n", addr)

	//loginHandler := handler.NewLoginHandler(cli)
	//if err := loginHandler.Login(username, password); err != nil {
	//	fmt.Printf("Login failed: %v\n", err)
	//	cli.Close()
	//	os.Exit(1)
	//}
	//fmt.Printf("Login success, playerId=%s\n", loginHandler.GetPlayerID())

	hbManager := handler.NewHeartbeatManager(cli, 5*time.Second)
	hbManager.Start()
	fmt.Println("Heartbeat manager started (interval: 5s)")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigCh
	fmt.Printf("Received signal %v, shutting down...\n", sig)

	hbManager.Stop()
	cli.Close()
	fmt.Println("Client shutdown complete")
}
