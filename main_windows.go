//go:build windows

package main

import (
	"context"
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	ioctlGpdAcpiECRead = 0x9C40A488
	ecBufferSize       = 0x400000
)

type ecReader struct {
	handle syscall.Handle
	buffer []byte
	debug  bool
}

func newECReader(devicePath string, debug bool) (*ecReader, error) {
	pathPtr, err := syscall.UTF16PtrFromString(devicePath)
	if err != nil {
		return nil, fmt.Errorf("invalid device path %q: %w", devicePath, err)
	}

	handle, err := syscall.CreateFile(
		pathPtr,
		syscall.GENERIC_READ|syscall.GENERIC_WRITE,
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE,
		nil,
		syscall.OPEN_EXISTING,
		syscall.FILE_ATTRIBUTE_NORMAL,
		0,
	)
	if err != nil {
		return nil, fmt.Errorf("open %s failed: %w", devicePath, err)
	}
	if handle == syscall.InvalidHandle {
		return nil, fmt.Errorf("open %s failed: invalid handle", devicePath)
	}

	return &ecReader{
		handle: handle,
		buffer: make([]byte, ecBufferSize),
		debug:  debug,
	}, nil
}

func (r *ecReader) Close() error {
	if r.handle == 0 || r.handle == syscall.InvalidHandle {
		return nil
	}
	err := syscall.CloseHandle(r.handle)
	r.handle = 0
	return err
}

func (r *ecReader) ReadByte(addr uint32) (byte, error) {
	binary.LittleEndian.PutUint32(r.buffer[:4], addr)

	var bytesReturned uint32
	err := syscall.DeviceIoControl(
		r.handle,
		ioctlGpdAcpiECRead,
		&r.buffer[0],
		uint32(len(r.buffer)),
		&r.buffer[0],
		uint32(len(r.buffer)),
		&bytesReturned,
		nil,
	)
	if err != nil {
		return 0, fmt.Errorf("DeviceIoControl(ECREAD addr=0x%04X) failed: %w", addr, err)
	}

	value := r.buffer[0]
	if r.debug {
		logf("[DEBUG] ec[0x%04X]=0x%02X bytesReturned=%d", addr, value, bytesReturned)
	}
	return value, nil
}

func readFanSample(r *ecReader) (fanSample, error) {
	cpuHi, err := r.ReadByte(addrCPURPMByte1)
	if err != nil {
		return fanSample{}, fmt.Errorf("read CPU RPM byte1(0x%04X): %w", addrCPURPMByte1, err)
	}
	cpuLo, err := r.ReadByte(addrCPURPMByte2)
	if err != nil {
		return fanSample{}, fmt.Errorf("read CPU RPM byte2(0x%04X): %w", addrCPURPMByte2, err)
	}
	gpuHi, err := r.ReadByte(addrGPURPMByte1)
	if err != nil {
		return fanSample{}, fmt.Errorf("read GPU RPM byte1(0x%04X): %w", addrGPURPMByte1, err)
	}
	gpuLo, err := r.ReadByte(addrGPURPMByte2)
	if err != nil {
		return fanSample{}, fmt.Errorf("read GPU RPM byte2(0x%04X): %w", addrGPURPMByte2, err)
	}
	cpuDuty, err := r.ReadByte(addrCPUFanDuty)
	if err != nil {
		return fanSample{}, fmt.Errorf("read CPU duty(0x%04X): %w", addrCPUFanDuty, err)
	}
	gpuDuty, err := r.ReadByte(addrGPUFanDuty)
	if err != nil {
		return fanSample{}, fmt.Errorf("read GPU duty(0x%04X): %w", addrGPUFanDuty, err)
	}
	alert, err := r.ReadByte(addrFanAlert)
	if err != nil {
		return fanSample{}, fmt.Errorf("read fan alert(0x%04X): %w", addrFanAlert, err)
	}

	return fanSample{
		cpuHi:   cpuHi,
		cpuLo:   cpuLo,
		gpuHi:   gpuHi,
		gpuLo:   gpuLo,
		cpuRPM:  rpmFromBytes(cpuHi, cpuLo),
		gpuRPM:  rpmFromBytes(gpuHi, gpuLo),
		cpuDuty: cpuDuty,
		gpuDuty: gpuDuty,
		alert:   alert,
	}, nil
}

func main() {
	devicePath := flag.String("device", `\\.\ACPIDriver`, `ACPI device path`)
	interval := flag.Duration("interval", time.Second, "poll interval")
	once := flag.Bool("once", false, "read once and exit")
	debug := flag.Bool("debug", false, "print per-register debug details")
	hwinfo := flag.Bool("hwinfo", false, "write CPU/GPU fan RPM to HWiNFO custom sensors in HKCU")
	hwinfoGroup := flag.String("hwinfo-group", "ECWatch", "HWiNFO custom sensor group name")
	flag.Parse()

	logf("starting ecwatch device=%s interval=%s", *devicePath, interval.String())

	reader, err := newECReader(*devicePath, *debug)
	if err != nil {
		fatalf("%v", err)
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			logf("close handle failed: %v", closeErr)
		}
	}()

	logf("device open successful")
	var hwWriter *hwinfoWriter
	if *hwinfo {
		hwWriter = newHWiNFOFanWriter(*hwinfoGroup, *debug)
		logf("hwinfo output enabled: HKCU\\Software\\HWiNFO64\\Sensors\\Custom\\%s\\Fan0|Fan1", sanitizeHWiNFOGroupName(*hwinfoGroup))
	}

	printOnce := func() {
		s, readErr := readFanSample(reader)
		if readErr != nil {
			logf("read error: %v", readErr)
			return
		}
		if hwWriter != nil {
			if err := hwWriter.WriteFans(s); err != nil {
				logf("hwinfo write error: %v", err)
			}
		}

		logf(
			"CPU=%dRPM GPU=%dRPM CPU_DUTY=%d%% GPU_DUTY=%d%% ALERT=0x%02X raw(cpu:%02X %02X gpu:%02X %02X)",
			s.cpuRPM,
			s.gpuRPM,
			s.cpuDuty,
			s.gpuDuty,
			s.alert,
			s.cpuHi,
			s.cpuLo,
			s.gpuHi,
			s.gpuLo,
		)
	}

	printOnce()
	if *once {
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	ticker := time.NewTicker(*interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logf("stopping")
			return
		case <-ticker.C:
			printOnce()
		}
	}
}

func logf(format string, args ...any) {
	fmt.Printf("[%s] %s\n", time.Now().Format("15:04:05"), fmt.Sprintf(format, args...))
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "[%s] %s\n", time.Now().Format("15:04:05"), fmt.Sprintf(format, args...))
	os.Exit(1)
}
