package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"tcp-ip-go/pkg/physical"
)

// SimplePacketHandler シンプルなパケットハンドラー
type SimplePacketHandler struct {
	PacketCount int
}

func (h *SimplePacketHandler) HandlePacket(data []byte) error {
	h.PacketCount++

	// パケット情報を表示
	info := physical.GetPacketInfo(data)
	fmt.Printf("[%d] %s\n", h.PacketCount, info)

	// 詳細なパケットダンプ（最初の数パケットのみ）
	if h.PacketCount <= 3 {
		physical.DumpPacket(data, fmt.Sprintf("パケット#%d", h.PacketCount))
	}

	return nil
}

func main() {
	fmt.Println("=== 簡易TCP/IPプロトコルスタック - 物理層テスト ===")

	// 既存のTUNデバイスに接続
	fmt.Println("TUNデバイス 'tun0' に接続中...")
	device, err := physical.OpenExistingTunDevice("tun0")
	if err != nil {
		fmt.Printf("エラー: TUNデバイスの接続に失敗しました: %v\n", err)
		fmt.Println("以下のコマンドでTUNデバイスを作成してください:")
		fmt.Println("ip tuntap add mode tun dev tun0 && ip link set tun0 up && ip addr add 10.0.0.1/24 dev tun0")
		os.Exit(1)
	}
	defer device.Close()

	// IPアドレスを設定
	ip := net.ParseIP("10.0.0.1")
	mask := net.CIDRMask(24, 32)
	err = device.SetIPAddress(ip, mask)
	if err != nil {
		fmt.Printf("エラー: IPアドレス設定に失敗しました: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("デバイス情報: %s\n", device.String())
	fmt.Println()

	// パケットハンドラーを作成
	handler := &SimplePacketHandler{}

	// シグナルハンドリング
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 停止チャネル
	stopChan := make(chan struct{})

	// シグナル受信時の処理
	go func() {
		sig := <-sigChan
		fmt.Printf("\n受信したシグナル: %v\n", sig)
		fmt.Println("プログラムを終了します...")
		close(stopChan)
	}()

	fmt.Println("パケット受信を開始します...")
	fmt.Println("テスト用コマンド:")
	fmt.Println("  ping -c 3 10.0.0.1")
	fmt.Println("  curl -m 5 http://10.0.0.1")
	fmt.Println("終了するには Ctrl+C を押してください")
	fmt.Println()

	// パケット受信ループを開始
	err = device.StartPacketLoop(handler, stopChan)
	if err != nil {
		fmt.Printf("エラー: パケット受信ループで問題が発生しました: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n総受信パケット数: %d\n", handler.PacketCount)
	fmt.Println("プログラムが正常に終了しました")
}
