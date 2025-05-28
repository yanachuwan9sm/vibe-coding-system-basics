package physical

import (
	"fmt"
	"io"
	"time"
)

// PacketHandler パケット処理のインターフェース
type PacketHandler interface {
	HandlePacket(data []byte) error
}

// ReadPacket TUNデバイスからパケットを読み取る
func (d *TunTapDevice) ReadPacket() ([]byte, error) {
	if !d.IsOpen() {
		return nil, fmt.Errorf("デバイスが開いていません")
	}

	// パケット用のバッファを準備（MTU 1500 + ヘッダ余裕）
	buffer := make([]byte, 2048)

	n, err := d.File.Read(buffer)
	if err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("TUNデバイスが閉じられました")
		}
		return nil, fmt.Errorf("パケット読み取りエラー: %v", err)
	}

	if n == 0 {
		return nil, fmt.Errorf("空のパケットを受信しました")
	}

	// 実際に読み取ったサイズのスライスを返す
	packet := make([]byte, n)
	copy(packet, buffer[:n])

	fmt.Printf("パケット受信: %d バイト\n", n)
	return packet, nil
}

// WritePacket TUNデバイスにパケットを書き込む
func (d *TunTapDevice) WritePacket(data []byte) error {
	if !d.IsOpen() {
		return fmt.Errorf("デバイスが開いていません")
	}

	if len(data) == 0 {
		return fmt.Errorf("空のパケットは送信できません")
	}

	n, err := d.File.Write(data)
	if err != nil {
		return fmt.Errorf("パケット送信エラー: %v", err)
	}

	if n != len(data) {
		return fmt.Errorf("パケット送信が不完全です: %d/%d バイト", n, len(data))
	}

	fmt.Printf("パケット送信: %d バイト\n", n)
	return nil
}

// ReadPacketWithTimeout タイムアウト付きでパケットを読み取る
func (d *TunTapDevice) ReadPacketWithTimeout(timeout time.Duration) ([]byte, error) {
	if !d.IsOpen() {
		return nil, fmt.Errorf("デバイスが開いていません")
	}

	// チャネルを使用してタイムアウト処理
	resultChan := make(chan struct {
		packet []byte
		err    error
	}, 1)

	go func() {
		packet, err := d.ReadPacket()
		resultChan <- struct {
			packet []byte
			err    error
		}{packet, err}
	}()

	select {
	case result := <-resultChan:
		return result.packet, result.err
	case <-time.After(timeout):
		return nil, fmt.Errorf("パケット読み取りタイムアウト")
	}
}

// StartPacketLoop パケット受信ループを開始する
func (d *TunTapDevice) StartPacketLoop(handler PacketHandler, stopChan <-chan struct{}) error {
	if !d.IsOpen() {
		return fmt.Errorf("デバイスが開いていません")
	}

	fmt.Printf("TUNデバイス '%s' でパケット受信ループを開始します\n", d.Name)

	for {
		select {
		case <-stopChan:
			fmt.Printf("パケット受信ループを停止します\n")
			return nil
		default:
			// パケットを読み取り（短いタイムアウト）
			packet, err := d.ReadPacketWithTimeout(100 * time.Millisecond)
			if err != nil {
				// タイムアウトの場合は継続
				if err.Error() == "パケット読み取りタイムアウト" {
					continue
				}
				return fmt.Errorf("パケット受信エラー: %v", err)
			}

			// パケットハンドラーで処理
			if handler != nil {
				err = handler.HandlePacket(packet)
				if err != nil {
					fmt.Printf("パケット処理エラー: %v\n", err)
					// エラーが発生しても継続
				}
			}
		}
	}
}

// DumpPacket パケットの内容を16進数でダンプする
func DumpPacket(data []byte, prefix string) {
	fmt.Printf("%s パケットダンプ (%d バイト):\n", prefix, len(data))

	for i := 0; i < len(data); i += 16 {
		// オフセットを表示
		fmt.Printf("%04x: ", i)

		// 16進数表示
		for j := 0; j < 16; j++ {
			if i+j < len(data) {
				fmt.Printf("%02x ", data[i+j])
			} else {
				fmt.Printf("   ")
			}

			// 8バイトごとにスペースを追加
			if j == 7 {
				fmt.Printf(" ")
			}
		}

		// ASCII表示
		fmt.Printf(" |")
		for j := 0; j < 16 && i+j < len(data); j++ {
			b := data[i+j]
			if b >= 32 && b <= 126 {
				fmt.Printf("%c", b)
			} else {
				fmt.Printf(".")
			}
		}
		fmt.Printf("|\n")
	}
	fmt.Println()
}

// GetPacketInfo パケットの基本情報を取得する（IPv4の場合）
func GetPacketInfo(data []byte) string {
	if len(data) < 20 {
		return fmt.Sprintf("パケットが短すぎます (%d バイト)", len(data))
	}

	// IPv4ヘッダの基本情報を取得
	version := (data[0] >> 4) & 0x0F
	if version != 4 {
		return fmt.Sprintf("IPv4以外のパケット (version: %d)", version)
	}

	protocol := data[9]
	srcIP := fmt.Sprintf("%d.%d.%d.%d", data[12], data[13], data[14], data[15])
	dstIP := fmt.Sprintf("%d.%d.%d.%d", data[16], data[17], data[18], data[19])

	protocolName := "Unknown"
	switch protocol {
	case 1:
		protocolName = "ICMP"
	case 6:
		protocolName = "TCP"
	case 17:
		protocolName = "UDP"
	}

	return fmt.Sprintf("IPv4 %s: %s -> %s (%d バイト)",
		protocolName, srcIP, dstIP, len(data))
}
