package main

import "strings"

type hwinfoFanValue struct {
	subKey string
	name   string
	value  uint32
}

func buildHWiNFOFanValues(sample fanSample) []hwinfoFanValue {
	return []hwinfoFanValue{
		{
			subKey: "Fan0",
			name:   "CPU Fan",
			value:  uint32(sample.cpuRPM),
		},
		{
			subKey: "Fan1",
			name:   "GPU Fan",
			value:  uint32(sample.gpuRPM),
		},
	}
}

func sanitizeHWiNFOGroupName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "ECWatch"
	}

	replacer := strings.NewReplacer(`\`, "_", `/`, "_", ":", "_", "*", "_", "?", "_", `"`, "_", "<", "_", ">", "_", "|", "_")
	name = replacer.Replace(name)
	name = strings.TrimSpace(name)
	if name == "" {
		return "ECWatch"
	}
	return name
}
