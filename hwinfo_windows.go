//go:build windows

package main

import (
	"fmt"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

type hwinfoWriter struct {
	basePath string
	debug    bool
}

func newHWiNFOFanWriter(groupName string, debug bool) *hwinfoWriter {
	groupName = sanitizeHWiNFOGroupName(groupName)
	return &hwinfoWriter{
		basePath: filepath.Join(`Software\HWiNFO64\Sensors\Custom`, groupName),
		debug:    debug,
	}
}

func (w *hwinfoWriter) WriteFans(sample fanSample) error {
	values := buildHWiNFOFanValues(sample)
	for _, v := range values {
		keyPath := filepath.Join(w.basePath, v.subKey)
		key, _, err := registry.CreateKey(registry.CURRENT_USER, keyPath, registry.SET_VALUE)
		if err != nil {
			return fmt.Errorf("create/open registry key HKCU\\%s failed: %w", keyPath, err)
		}

		if err := key.SetStringValue("Name", v.name); err != nil {
			key.Close()
			return fmt.Errorf("set Name in HKCU\\%s failed: %w", keyPath, err)
		}
		if err := key.SetDWordValue("Value", v.value); err != nil {
			key.Close()
			return fmt.Errorf("set Value in HKCU\\%s failed: %w", keyPath, err)
		}
		if err := key.Close(); err != nil {
			return fmt.Errorf("close registry key HKCU\\%s failed: %w", keyPath, err)
		}
		if w.debug {
			logf("[DEBUG] hwinfo HKCU\\%s Name=%q Value=%d", keyPath, v.name, v.value)
		}
	}
	return nil
}
