package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
)

type ProcessInfo struct {
	TotalRAM  int       `json:"Totalram"`
	FreeRAM   int       `json:"Freeram"`
	Procs     int       `json:"Procs"`
	Processes []Process `json:"Processes"`
}

type Process struct {
	PID         int     `json:"PID"`
	Name        string  `json:"Name"`
	Cmdline     string  `json:"Cmdline"`
	Vsz         int     `json:"vsz"`
	Rss         int     `json:"rss"`
	MemoryUsage float64 `json:"Memory_Usage"`
	CPUUsage    int     `json:"CPU_Usage"`
}

func monitorProcesses() error {
	// Leer archivo /proc
	data, err := ioutil.ReadFile("/proc/sysinfo_so1_201902125")
	if err != nil {
		return fmt.Errorf("error leyendo /proc: %v", err)
	}

	// Parsear JSON
	var procInfo ProcessInfo
	err = json.Unmarshal(data, &procInfo)
	if err != nil {
		return fmt.Errorf("error parseando JSON: %v", err)
	}

	// Mostrar información
	fmt.Printf("RAM Total: %d KB, Libre: %d KB, Procesos: %d\n",
		procInfo.TotalRAM, procInfo.FreeRAM, len(procInfo.Processes))

	// Guardar en BD
	err = saveMetrics(procInfo.TotalRAM, procInfo.FreeRAM, len(procInfo.Processes))
	if err != nil {
		log.Printf("Error guardando los datos: %v", err)
	}

	analyzeContainers(procInfo.Processes)

	return nil
}

func analyzeContainers(processes []Process) {
	containerCount := 0
	for _, proc := range processes {
		if isContainerProcess(proc) {
			containerCount++
			fmt.Printf("Contenedor detectado: PID %d, %s\n", proc.PID, proc.Name)
		}
	}
	fmt.Printf("Total contenedores activos: %d\n", containerCount)
}

func isContainerProcess(proc Process) bool {
	return contains(proc.Name, "containerd") ||
		contains(proc.Cmdline, "docker") ||
		contains(proc.Cmdline, "container")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}
