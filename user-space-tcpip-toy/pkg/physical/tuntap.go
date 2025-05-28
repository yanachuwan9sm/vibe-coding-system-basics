package physical

import (
	"fmt"
	"net"
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

// TUN/TAPデバイスの設定用定数
const (
	TUNSETIFF    = 0x400454ca
	IFF_TUN      = 0x0001
	IFF_TAP      = 0x0002
	IFF_NO_PI    = 0x1000
	CLONE_DEVICE = "/dev/net/tun"
)

// ifreq構造体 - ioctl用
type ifreq struct {
	Name  [16]byte
	Flags uint16
	pad   [22]byte
}

// TunTapDevice TUN/TAPデバイスを管理する構造体
type TunTapDevice struct {
	Name      string
	File      *os.File
	IPAddress net.IP
	Netmask   net.IPMask
	fd        int
}

// NewTunDevice 新しいTUNデバイスを作成する
func NewTunDevice(name string) (*TunTapDevice, error) {
	// /dev/net/tunを開く
	file, err := os.OpenFile(CLONE_DEVICE, os.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("TUNデバイスファイルを開けません: %v", err)
	}

	// ifreq構造体を準備
	var req ifreq
	copy(req.Name[:], name)
	req.Flags = IFF_TUN | IFF_NO_PI

	// TUNデバイスを作成
	fd := int(file.Fd())
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), TUNSETIFF, uintptr(unsafe.Pointer(&req)))
	if errno != 0 {
		file.Close()
		return nil, fmt.Errorf("TUNデバイスの作成に失敗しました: %v", errno)
	}

	device := &TunTapDevice{
		Name: name,
		File: file,
		fd:   fd,
	}

	fmt.Printf("TUNデバイス '%s' を作成しました\n", name)
	return device, nil
}

// OpenExistingTunDevice 既存のTUNデバイスに接続する
func OpenExistingTunDevice(name string) (*TunTapDevice, error) {
	// /dev/net/tunを開く
	file, err := os.OpenFile(CLONE_DEVICE, os.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("TUNデバイスファイルを開けません: %v", err)
	}

	// ifreq構造体を準備
	var req ifreq
	copy(req.Name[:], name)
	req.Flags = IFF_TUN | IFF_NO_PI

	// 既存のTUNデバイスに接続
	fd := int(file.Fd())
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), TUNSETIFF, uintptr(unsafe.Pointer(&req)))
	if errno != 0 {
		file.Close()
		return nil, fmt.Errorf("既存のTUNデバイスへの接続に失敗しました: %v", errno)
	}

	device := &TunTapDevice{
		Name: name,
		File: file,
		fd:   fd,
	}

	fmt.Printf("既存のTUNデバイス '%s' に接続しました\n", name)
	return device, nil
}

// SetIPAddress IPアドレスとネットマスクを設定する
func (d *TunTapDevice) SetIPAddress(ip net.IP, mask net.IPMask) error {
	d.IPAddress = ip
	d.Netmask = mask

	fmt.Printf("TUNデバイス '%s' にIPアドレス %s/%s を設定しました\n",
		d.Name, ip.String(), net.IP(mask).String())
	return nil
}

// GetIPAddress 設定されているIPアドレスを取得する
func (d *TunTapDevice) GetIPAddress() (net.IP, net.IPMask) {
	return d.IPAddress, d.Netmask
}

// GetName デバイス名を取得する
func (d *TunTapDevice) GetName() string {
	return d.Name
}

// GetFD ファイルディスクリプタを取得する
func (d *TunTapDevice) GetFD() int {
	return d.fd
}

// IsOpen デバイスが開いているかチェックする
func (d *TunTapDevice) IsOpen() bool {
	return d.File != nil
}

// SetNonBlocking ノンブロッキングモードを設定する
func (d *TunTapDevice) SetNonBlocking(nonblocking bool) error {
	if !d.IsOpen() {
		return fmt.Errorf("デバイスが開いていません")
	}

	flags, err := unix.FcntlInt(uintptr(d.fd), unix.F_GETFL, 0)
	if err != nil {
		return fmt.Errorf("フラグの取得に失敗しました: %v", err)
	}

	if nonblocking {
		flags |= unix.O_NONBLOCK
	} else {
		flags &^= unix.O_NONBLOCK
	}

	_, err = unix.FcntlInt(uintptr(d.fd), unix.F_SETFL, flags)
	if err != nil {
		return fmt.Errorf("ノンブロッキングモードの設定に失敗しました: %v", err)
	}

	fmt.Printf("TUNデバイス '%s' のノンブロッキングモード: %t\n", d.Name, nonblocking)
	return nil
}

// Close TUNデバイスを閉じる
func (d *TunTapDevice) Close() error {
	if d.File != nil {
		err := d.File.Close()
		d.File = nil
		d.fd = -1
		fmt.Printf("TUNデバイス '%s' を閉じました\n", d.Name)
		return err
	}
	return nil
}

// String デバイス情報を文字列で返す
func (d *TunTapDevice) String() string {
	if d.IPAddress != nil && d.Netmask != nil {
		return fmt.Sprintf("TUNDevice{Name: %s, IP: %s/%s, Open: %t}",
			d.Name, d.IPAddress.String(), net.IP(d.Netmask).String(), d.IsOpen())
	}
	return fmt.Sprintf("TUNDevice{Name: %s, IP: not set, Open: %t}", d.Name, d.IsOpen())
}
