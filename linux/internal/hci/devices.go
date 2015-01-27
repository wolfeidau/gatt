package hci

import (
	"fmt"
	"syscall"
	"unsafe"

	"github.com/paypal/gatt/linux/internal/socket"
	"github.com/wolfeidau/gioctl"
)

const (
	IOCTLSize            = uintptr(4)
	HCIGetDeviceListCode = 72
	HCIMaxDevices        = 16
)

var (
	HCIGetDeviceList = gioctl.IoR(HCIGetDeviceListCode, 210, IOCTLSize) // HCIGETDEVLIST
	HCIGetDeviceInfo = gioctl.IoR(HCIGetDeviceListCode, 211, IOCTLSize) // HCIGETDEVINFO
)

type HCIDeviceRequest struct {
	DevId  uint16
	DevOpt uint32
}

type HCIDeviceListRequest struct {
	DevNum     uint16
	DevRequest [HCIMaxDevices]HCIDeviceRequest
}

type HCIDeviceInfo struct {
	DevId uint16
	name  [8]byte

	btAddr [6]byte

	Flags   uint32
	DevType uint8

	Features [8]uint8

	PktType    uint32
	LinkPolicy uint32
	LinkMode   uint32

	AclMtu  uint16
	AclPkts uint16
	ScoMtu  uint16
	ScoPkts uint16

	Stats HCIDeviceStats
}

func (hdi *HCIDeviceInfo) Name() string {
	return string(hdi.name[:])
}

func (hdi *HCIDeviceInfo) Addr() string {
	return fmt.Sprintf("%.2x:%.2x:%.2x:%.2x:%.2x:%.2x",
		hdi.btAddr[5], hdi.btAddr[4], hdi.btAddr[3], hdi.btAddr[2], hdi.btAddr[1], hdi.btAddr[0]) // yeah backwards, who knew right!?
}

type HCIDeviceStats struct {
	ErrRx  uint32
	ErrTx  uint32
	CmdTx  uint32
	EvtRx  uint32
	AclTx  uint32
	AclRx  uint32
	ScoTx  uint32
	ScoRx  uint32
	ByteRx uint32
	ByteTx uint32
}

type hciRequest struct {
	ogf    uint16
	ocf    uint16
	event  int
	cparam uintptr
	clen   int
	rparam uintptr
	rlen   int
}

type leSetAdvertiseEnable struct {
	enable uint8
}

func GetDeviceList() ([]*HCIDeviceInfo, error) {

	var (
		devList    []*HCIDeviceInfo
		devListReq HCIDeviceListRequest
	)

	fd, err := syscall.Socket(socket.AF_BLUETOOTH, syscall.SOCK_RAW, socket.BTPROTO_HCI)

	if err != nil {
		return devList, err
	}

	devListReq.DevNum = HCIMaxDevices

	err = gioctl.Ioctl(uintptr(fd), HCIGetDeviceList, uintptr(unsafe.Pointer(&devListReq)))

	if err != nil {
		return devList, err
	}

	for i := 0; i < int(devListReq.DevNum); i++ {
		// TODO check status of device
		var deviceInfo HCIDeviceInfo

		deviceInfo.DevId = uint16(i)

		err = gioctl.Ioctl(uintptr(fd), HCIGetDeviceInfo, uintptr(unsafe.Pointer(&deviceInfo)))

		if err != nil {
			return devList, err
		}

		devList = append(devList, &deviceInfo)
	}

	return devList, nil
}
