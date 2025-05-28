package test

import (
	"net"
	"tcp-ip-go/pkg/physical"
	"testing"
	"time"
)

// TestPacketHandler テスト用のパケットハンドラー
type TestPacketHandler struct {
	ReceivedPackets [][]byte
	PacketCount     int
}

func (h *TestPacketHandler) HandlePacket(data []byte) error {
	h.PacketCount++
	// パケットのコピーを保存
	packet := make([]byte, len(data))
	copy(packet, data)
	h.ReceivedPackets = append(h.ReceivedPackets, packet)

	// パケット情報を表示
	info := physical.GetPacketInfo(data)
	println("受信パケット:", info)

	return nil
}

// TestTunDeviceCreation TUNデバイスの作成テスト
func TestTunDeviceCreation(t *testing.T) {
	// 既存のTUNデバイスに接続
	device, err := physical.OpenExistingTunDevice("tun0")
	if err != nil {
		t.Fatalf("TUNデバイスの接続に失敗: %v", err)
	}
	defer device.Close()

	// デバイス情報の確認
	if device.GetName() != "tun0" {
		t.Errorf("デバイス名が正しくありません: expected 'tun0', got '%s'", device.GetName())
	}

	if !device.IsOpen() {
		t.Error("デバイスが開いていません")
	}

	t.Logf("TUNデバイス接続成功: %s", device.String())
}

// TestIPAddressSetting IPアドレス設定テスト
func TestIPAddressSetting(t *testing.T) {
	device, err := physical.OpenExistingTunDevice("tun0")
	if err != nil {
		t.Fatalf("TUNデバイスの接続に失敗: %v", err)
	}
	defer device.Close()

	// IPアドレスとネットマスクを設定
	ip := net.ParseIP("10.0.0.1")
	mask := net.CIDRMask(24, 32)

	err = device.SetIPAddress(ip, mask)
	if err != nil {
		t.Fatalf("IPアドレス設定に失敗: %v", err)
	}

	// 設定されたIPアドレスを確認
	gotIP, gotMask := device.GetIPAddress()
	if !gotIP.Equal(ip) {
		t.Errorf("IPアドレスが正しくありません: expected %s, got %s", ip.String(), gotIP.String())
	}

	if !net.IP(gotMask).Equal(net.IP(mask)) {
		t.Errorf("ネットマスクが正しくありません: expected %s, got %s",
			net.IP(mask).String(), net.IP(gotMask).String())
	}

	t.Logf("IPアドレス設定成功: %s", device.String())
}

// TestNonBlockingMode ノンブロッキングモードテスト
func TestNonBlockingMode(t *testing.T) {
	device, err := physical.OpenExistingTunDevice("tun0")
	if err != nil {
		t.Fatalf("TUNデバイスの接続に失敗: %v", err)
	}
	defer device.Close()

	// ノンブロッキングモードを有効にする
	err = device.SetNonBlocking(true)
	if err != nil {
		t.Fatalf("ノンブロッキングモード設定に失敗: %v", err)
	}

	// ノンブロッキングモードを無効にする
	err = device.SetNonBlocking(false)
	if err != nil {
		t.Fatalf("ブロッキングモード設定に失敗: %v", err)
	}

	t.Log("ノンブロッキングモード設定成功")
}

// TestPacketReception パケット受信テスト
func TestPacketReception(t *testing.T) {
	device, err := physical.OpenExistingTunDevice("tun0")
	if err != nil {
		t.Fatalf("TUNデバイスの接続に失敗: %v", err)
	}
	defer device.Close()

	// IPアドレスを設定
	ip := net.ParseIP("10.0.0.1")
	mask := net.CIDRMask(24, 32)
	device.SetIPAddress(ip, mask)

	t.Log("パケット受信テストを開始します...")
	t.Log("別のターミナルで以下のコマンドを実行してください:")
	t.Log("ping -c 3 10.0.0.1")

	// テスト用ハンドラーを作成
	handler := &TestPacketHandler{
		ReceivedPackets: make([][]byte, 0),
	}

	// 停止チャネルを作成
	stopChan := make(chan struct{})

	// 5秒後に停止
	go func() {
		time.Sleep(5 * time.Second)
		close(stopChan)
	}()

	// パケット受信ループを開始
	err = device.StartPacketLoop(handler, stopChan)
	if err != nil {
		t.Fatalf("パケット受信ループエラー: %v", err)
	}

	t.Logf("受信したパケット数: %d", handler.PacketCount)

	// 受信したパケットの詳細を表示
	for i, packet := range handler.ReceivedPackets {
		t.Logf("パケット %d: %s", i+1, physical.GetPacketInfo(packet))
		if len(packet) > 0 {
			physical.DumpPacket(packet, "受信")
		}
	}
}

// TestPacketTimeout タイムアウト付きパケット受信テスト
func TestPacketTimeout(t *testing.T) {
	device, err := physical.OpenExistingTunDevice("tun0")
	if err != nil {
		t.Fatalf("TUNデバイスの接続に失敗: %v", err)
	}
	defer device.Close()

	// 短いタイムアウトでパケット受信を試行
	packet, err := device.ReadPacketWithTimeout(100 * time.Millisecond)
	if err != nil && err.Error() == "パケット読み取りタイムアウト" {
		t.Log("タイムアウトが正常に動作しています")
	} else if err != nil {
		t.Logf("予期しないエラー: %v", err)
	} else {
		t.Logf("パケットを受信しました: %s", physical.GetPacketInfo(packet))
	}
}

// BenchmarkPacketReception パケット受信のベンチマーク
func BenchmarkPacketReception(b *testing.B) {
	device, err := physical.OpenExistingTunDevice("tun0")
	if err != nil {
		b.Fatalf("TUNデバイスの接続に失敗: %v", err)
	}
	defer device.Close()

	// ベンチマーク用の簡単なハンドラー
	handler := &TestPacketHandler{}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// タイムアウト付きでパケット受信を試行
		packet, err := device.ReadPacketWithTimeout(1 * time.Millisecond)
		if err == nil {
			handler.HandlePacket(packet)
		}
	}
}
