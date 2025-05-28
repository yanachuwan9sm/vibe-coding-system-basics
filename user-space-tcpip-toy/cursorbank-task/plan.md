# 簡易TCP/IPプロトコルスタック実装計画

## 概要
Go言語を使用して、物理層からTCP層までの簡易的なネットワークプロトコルスタックを実装します。
TUN/TAPデバイスを使用してパケットの送受信を行い、IPv4とTCPプロトコルを処理します。

## プロジェクト構造

```
tcp-ip-go/
├── cmd/
│   └── main.go                 # メインエントリーポイント
├── pkg/
│   ├── physical/               # 物理層
│   │   ├── tuntap.go          # TUN/TAPデバイス管理
│   │   └── packet.go          # パケット送受信
│   ├── ip/                    # IP層
│   │   ├── ipv4.go            # IPv4ヘッダ処理
│   │   ├── checksum.go        # チェックサム計算
│   │   └── parser.go          # パケット解析
│   ├── tcp/                   # TCP層
│   │   ├── tcp.go             # TCPヘッダ処理
│   │   ├── state.go           # TCP状態管理
│   │   ├── handshake.go       # 3-way handshake
│   │   └── checksum.go        # TCPチェックサム
│   └── utils/                 # 共通ユーティリティ
│       ├── debug.go           # デバッグ機能
│       └── pause.go           # 一時停止機能
├── test/                      # テストファイル
│   ├── physical_test.go
│   ├── ip_test.go
│   └── tcp_test.go
├── go.mod
├── go.sum
└── README.md
```

## 実装フェーズ

### フェーズ1: 物理層実装

#### 1.1 TUN/TAPデバイス管理 (`pkg/physical/tuntap.go`)

**目的**: TUN/TAPデバイスの作成、設定、管理を行う

**実装内容**:
- TUN/TAPデバイスの作成と設定
- デバイスファイルディスクリプタの管理
- ネットワークインターフェースの設定（IPアドレス、ルーティング）

**主要構造体**:
```go
type TunTapDevice struct {
    Name       string
    File       *os.File
    IPAddress  net.IP
    Netmask    net.IPMask
}
```

**主要メソッド**:
- `NewTunDevice(name string) (*TunTapDevice, error)`: TUNデバイス作成
- `SetIPAddress(ip net.IP, mask net.IPMask) error`: IPアドレス設定
- `Close() error`: デバイスクローズ

#### 1.2 パケット送受信 (`pkg/physical/packet.go`)

**目的**: 生のパケットデータの送受信を行う

**実装内容**:
- パケットの読み取り（受信）
- パケットの書き込み（送信）
- エラーハンドリング

**主要メソッド**:
- `ReadPacket() ([]byte, error)`: パケット受信
- `WritePacket(data []byte) error`: パケット送信

#### 1.3 物理層テスト (`test/physical_test.go`)

**テスト内容**:
- TUN/TAPデバイスの作成・設定テスト
- パケット送受信の基本動作テスト
- 実際のネットワークトラフィックの受信確認

### フェーズ2: IP層実装

#### 2.1 IPv4ヘッダ処理 (`pkg/ip/ipv4.go`)

**目的**: IPv4ヘッダの解析と生成を行う

**実装内容**:
- IPv4ヘッダ構造体の定義
- ヘッダのパース機能
- ヘッダの生成機能（`buildIPv4Header`）

**主要構造体**:
```go
type IPv4Header struct {
    Version        uint8
    IHL            uint8
    TOS            uint8
    TotalLength    uint16
    Identification uint16
    Flags          uint8
    FragmentOffset uint16
    TTL            uint8
    Protocol       uint8
    HeaderChecksum uint16
    SrcIP          net.IP
    DstIP          net.IP
    Options        []byte
}
```

**主要メソッド**:
- `ParseIPv4Header(data []byte) (*IPv4Header, error)`: ヘッダ解析
- `BuildIPv4Header(src, dst net.IP, protocol uint8, payloadLen int) *IPv4Header`: ヘッダ生成
- `ToBytes() []byte`: バイト配列への変換

#### 2.2 チェックサム計算 (`pkg/ip/checksum.go`)

**目的**: IPv4ヘッダチェックサムの計算と検証

**実装内容**:
- インターネットチェックサムアルゴリズム
- ヘッダチェックサムの計算
- チェックサムの検証

**主要メソッド**:
- `CalculateChecksum(data []byte) uint16`: チェックサム計算
- `VerifyChecksum(header *IPv4Header) bool`: チェックサム検証

#### 2.3 パケット解析 (`pkg/ip/parser.go`)

**目的**: 受信パケットの解析とプロトコル分岐

**実装内容**:
- IPv4パケットの解析
- プロトコル番号による分岐（TCP/UDP/ICMP）
- 一時停止機能の統合（`pauseIfNeeded("ip")`）

**主要メソッド**:
- `ProcessPacket(data []byte) error`: パケット処理
- `RouteToProtocol(header *IPv4Header, payload []byte) error`: プロトコル分岐

#### 2.4 IP層テスト (`test/ip_test.go`)

**テスト内容**:
- IPv4ヘッダの解析・生成テスト
- チェックサム計算の正確性テスト
- プロトコル分岐の動作テスト

### フェーズ3: TCP層実装

#### 3.1 TCPヘッダ処理 (`pkg/tcp/tcp.go`)

**目的**: TCPヘッダの解析と生成を行う

**実装内容**:
- TCPヘッダ構造体の定義
- ヘッダのパース機能
- ヘッダの生成機能（`buildTCPHeader`）

**主要構造体**:
```go
type TCPHeader struct {
    SrcPort    uint16
    DstPort    uint16
    SeqNum     uint32
    AckNum     uint32
    DataOffset uint8
    Flags      TCPFlags
    Window     uint16
    Checksum   uint16
    UrgentPtr  uint16
    Options    []byte
}

type TCPFlags struct {
    FIN bool
    SYN bool
    RST bool
    PSH bool
    ACK bool
    URG bool
    ECE bool
    CWR bool
}
```

#### 3.2 TCP状態管理 (`pkg/tcp/state.go`)

**目的**: TCP接続の状態管理

**実装内容**:
- TCP状態の定義（CLOSED, LISTEN, SYN_SENT, SYN_RECEIVED, ESTABLISHED, FIN_WAIT_1, etc.）
- 状態遷移の管理
- 接続テーブルの管理

**主要構造体**:
```go
type TCPState int

const (
    CLOSED TCPState = iota
    LISTEN
    SYN_SENT
    SYN_RECEIVED
    ESTABLISHED
    FIN_WAIT_1
    FIN_WAIT_2
    CLOSE_WAIT
    CLOSING
    LAST_ACK
    TIME_WAIT
)

type TCPConnection struct {
    LocalAddr  net.TCPAddr
    RemoteAddr net.TCPAddr
    State      TCPState
    SeqNum     uint32
    AckNum     uint32
    Window     uint16
}
```

#### 3.3 3-way handshake (`pkg/tcp/handshake.go`)

**目的**: TCP接続確立の3-way handshakeを実装

**実装内容**:
- SYNパケットの処理
- SYN-ACKパケットの生成・処理
- ACKパケットの処理
- handshake完了時の一時停止（`pauseIfNeeded("tcp")`）

**主要メソッド**:
- `HandleSYN(conn *TCPConnection, header *TCPHeader) error`: SYN処理
- `SendSYNACK(conn *TCPConnection) error`: SYN-ACK送信
- `HandleACK(conn *TCPConnection, header *TCPHeader) error`: ACK処理

#### 3.4 TCPチェックサム (`pkg/tcp/checksum.go`)

**目的**: TCPチェックサムの計算（疑似ヘッダを含む）

**実装内容**:
- 疑似ヘッダの生成
- TCPチェックサムの計算
- チェックサムの検証

### フェーズ4: 共通ユーティリティ

#### 4.1 デバッグ機能 (`pkg/utils/debug.go`)

**目的**: デバッグ情報の出力とログ管理

**実装内容**:
- パケットダンプ機能
- ヘッダ情報の表示
- ログレベル管理

#### 4.2 一時停止機能 (`pkg/utils/pause.go`)

**目的**: デバッグ用の一時停止機能

**実装内容**:
- `pauseIfNeeded(layer string)` 関数の実装
- 環境変数による制御
- ユーザー入力待ち

### フェーズ5: メインアプリケーション

#### 5.1 メインエントリーポイント (`cmd/main.go`)

**目的**: 全体の統合とメインループ

**実装内容**:
- TUN/TAPデバイスの初期化
- パケット受信ループ
- 各層の統合
- エラーハンドリング

## 実装順序

1. **物理層の基盤実装**
   - TUN/TAPデバイス管理
   - 基本的なパケット送受信

2. **物理層テスト**
   - 実際のパケット受信確認
   - デバイス設定の検証

3. **IP層実装**
   - IPv4ヘッダ処理
   - チェックサム計算
   - プロトコル分岐

4. **TCP層実装**
   - TCPヘッダ処理
   - 状態管理
   - 3-way handshake

5. **統合テスト**
   - 全体の動作確認
   - 実際のTCP通信テスト

## 技術的考慮事項

### セキュリティ
- root権限が必要なTUN/TAPデバイス操作
- 適切な権限管理

### パフォーマンス
- 効率的なパケット処理
- メモリ使用量の最適化

### エラーハンドリング
- ネットワークエラーの適切な処理
- リソースリークの防止

### テスト戦略
- 単体テスト
- 統合テスト
- 実際のネットワーク環境でのテスト

## 必要な外部依存関係

```go
// go.mod
module tcp-ip-go

go 1.21

require (
    golang.org/x/sys v0.15.0  // システムコール用
)
```

この実装計画に従って、段階的に簡易TCP/IPプロトコルスタックを構築していきます。 