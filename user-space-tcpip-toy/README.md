# 簡易TCP/IPプロトコルスタック

Go言語で実装する簡易的なTCP/IPプロトコルスタックです。TUN/TAPデバイスを使用してパケットの送受信を行い、IPv4とTCPプロトコルを処理します。

## 現在の実装状況

### ✅ フェーズ1: 物理層実装 (完了)
- TUN/TAPデバイス管理
- パケット送受信機能
- 基本テスト

### 🚧 フェーズ2: IP層実装 (予定)
- IPv4ヘッダ処理
- チェックサム計算
- プロトコル分岐

### 🚧 フェーズ3: TCP層実装 (予定)
- TCPヘッダ処理
- 状態管理
- 3-way handshake

## 必要な環境

- Go 1.21以上
- Linux環境（TUN/TAPデバイス対応）
- root権限（TUN/TAPデバイス操作のため）

## セットアップ

### 1. TUNデバイスの作成

```bash
# TUNデバイスを作成し、IPアドレスを設定
sudo ip tuntap add mode tun dev tun0
sudo ip link set tun0 up
sudo ip addr add 10.0.0.1/24 dev tun0
```

### 2. 依存関係のインストール

```bash
go mod tidy
```

## 使用方法

### メインプログラムの実行

```bash
# root権限で実行
sudo go run cmd/main.go
```

### テストの実行

```bash
# 基本テスト
sudo go test ./test -v

# 特定のテスト
sudo go test ./test -v -run TestTunDeviceCreation

# ベンチマーク
sudo go test ./test -v -bench=.
```

### パケット受信テスト

プログラム実行後、別のターミナルで以下のコマンドを実行してパケットを送信できます：

```bash
# ICMPパケット（ping）
ping -c 3 10.0.0.1

# TCPパケット（curl）
curl -m 5 http://10.0.0.1

# UDPパケット（nc）
echo "test" | nc -u 10.0.0.1 8080
```

## プロジェクト構造

```
tcp-ip-go/
├── cmd/
│   └── main.go                 # メインエントリーポイント
├── pkg/
│   ├── physical/               # 物理層
│   │   ├── tuntap.go          # TUN/TAPデバイス管理
│   │   └── packet.go          # パケット送受信
│   ├── ip/                    # IP層（予定）
│   ├── tcp/                   # TCP層（予定）
│   └── utils/                 # 共通ユーティリティ（予定）
├── test/                      # テストファイル
│   └── physical_test.go       # 物理層テスト
├── go.mod
├── go.sum
└── README.md
```

## 実装済み機能

### 物理層 (`pkg/physical/`)

#### TUN/TAPデバイス管理 (`tuntap.go`)
- `NewTunDevice()`: 新しいTUNデバイスの作成
- `OpenExistingTunDevice()`: 既存のTUNデバイスへの接続
- `SetIPAddress()`: IPアドレスとネットマスクの設定
- `SetNonBlocking()`: ノンブロッキングモードの設定
- `Close()`: デバイスのクローズ

#### パケット送受信 (`packet.go`)
- `ReadPacket()`: パケットの読み取り
- `WritePacket()`: パケットの書き込み
- `ReadPacketWithTimeout()`: タイムアウト付きパケット読み取り
- `StartPacketLoop()`: パケット受信ループ
- `DumpPacket()`: パケットの16進数ダンプ
- `GetPacketInfo()`: パケット基本情報の取得

## 使用例

```go
package main

import (
    "net"
    "tcp-ip-go/pkg/physical"
)

func main() {
    // TUNデバイスに接続
    device, err := physical.OpenExistingTunDevice("tun0")
    if err != nil {
        panic(err)
    }
    defer device.Close()

    // IPアドレスを設定
    ip := net.ParseIP("10.0.0.1")
    mask := net.CIDRMask(24, 32)
    device.SetIPAddress(ip, mask)

    // パケットを受信
    packet, err := device.ReadPacket()
    if err != nil {
        panic(err)
    }

    // パケット情報を表示
    info := physical.GetPacketInfo(packet)
    println(info)
}
```

## トラブルシューティング

### 権限エラー
```
permission denied
```
→ root権限で実行してください：`sudo go run cmd/main.go`

### TUNデバイスが見つからない
```
no such device
```
→ TUNデバイスを作成してください：
```bash
sudo ip tuntap add mode tun dev tun0
sudo ip link set tun0 up
sudo ip addr add 10.0.0.1/24 dev tun0
```

### パケットが受信されない
- TUNデバイスが正しく設定されているか確認
- ファイアウォールの設定を確認
- 別のターミナルからテストパケットを送信

## 今後の実装予定

1. **IP層の実装**
   - IPv4ヘッダの解析と生成
   - チェックサム計算
   - プロトコル分岐（TCP/UDP/ICMP）

2. **TCP層の実装**
   - TCPヘッダの解析と生成
   - 3-way handshake
   - 状態管理

3. **統合テスト**
   - 実際のTCP通信テスト
   - パフォーマンステスト

## ライセンス

MIT License 