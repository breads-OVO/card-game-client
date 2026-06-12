package main

import (
	"bufio"
	"card-game-client/client"
	"card-game-client/handler"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	addr := "127.0.0.1:9527"

	cli := client.NewTCPClient(addr)
	if err := cli.Connect(); err != nil {
		fmt.Printf("Failed to connect to server %s: %v\n", addr, err)
		os.Exit(1)
	}
	fmt.Printf("Connected to server %s successfully\n", addr)

	hbManager := handler.NewHeartbeatManager(cli, 5*time.Second)
	hbManager.Start()
	fmt.Println("Heartbeat manager started (interval: 5s)")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	AuthorManager := handler.NewAuthorManager(cli)
	// 启动一个 goroutine 监听系统信号 (Ctrl+C)
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigCh
		fmt.Printf("\nReceived signal %v, shutting down...\n", sig)
		cancel() // 通知主流程退出
	}()

	// 启动一个 goroutine 监听用户输入
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for {
			fmt.Print("> ") // 命令提示符
			if !scanner.Scan() {
				break // 输入结束 (如 Ctrl+D)
			}
			input := strings.TrimSpace(scanner.Text())
			if input == "" {
				continue
			}

			// 解析命令和参数 (支持空格分隔)
			parts := strings.Fields(input)
			cmd := parts[0]
			// 分发命令
			handleCommand(cmd, AuthorManager, cancel)
		}
	}()
	printHelp()
	<-ctx.Done()
	hbManager.Stop()
	cli.Close()
	fmt.Println("Client shutdown complete")
}

// handleCommand 处理用户输入的命令
func handleCommand(cmd string, authorManager *handler.AuthorManager, cancel context.CancelFunc) {

	switch strings.ToLower(cmd) {
	case "exit", "quit":
		fmt.Println("Exiting...")
		cancel() // 触发退出
	case "register":
		err := authorManager.Register()
		if err != nil {
			fmt.Println("Register failed:", err)
			return
		}
	case "login":
		err := authorManager.Login()
		if err != nil {
			fmt.Println("Login failed:", err)
			return
		}
	case "logout":
		err := authorManager.Logout()
		if err != nil {
			fmt.Println("Logout failed:", err)
			return
		}

	default:
		fmt.Printf("Unknown command: %s. Type 'help' for usage.\n", cmd)
	}
}

// printHelp 打印帮助信息
func printHelp() {
	fmt.Println("\n=== Available Commands ===")
	fmt.Println("  help        - Show this help message")
	fmt.Println("  send <msg>  - Send a custom message")
	fmt.Println("  exit/quit   - Shutdown the client")
	fmt.Println("==========================")
}
