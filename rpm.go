package main

const (
	addrCPURPMByte1 = 0x0464
	addrCPURPMByte2 = 0x0465
	addrGPURPMByte1 = 0x046C
	addrGPURPMByte2 = 0x046B
	addrCPUFanDuty  = 0x075B
	addrGPUFanDuty  = 0x075C
	addrFanAlert    = 0x0741
)

func rpmFromBytes(byte1, byte2 byte) int {
	return int(uint16(byte1)<<8 | uint16(byte2))
}
