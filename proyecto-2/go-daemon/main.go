package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

func main() {
	fmt.Println("Iniciando Daemon...")

	if os.Geteuid() != 0 {
		log.Fatal("Este daemon debe ejecutarse como root (sudo)")
	}

	fmt.Println("Cargando módulo de kernel...")
	err := loadKernelModule()
	if err != nil {
		log.Printf("No se pudo cargar módulo: %v", err)
	}

	fmt.Println("Iniciando base de datos...")
	err = initDatabase()
	if err != nil {
		log.Fatalf("Error iniciando la base de datos: %v", err)
	}

	fmt.Println("Configurando cronjob...")
	setupCronJob()

	fmt.Println("Iniciando loop de monitoreo (20 segundos)...")
	startMonitoringLoop()
}

func loadKernelModule() error {

	cmd := exec.Command("sh", "-c", "cd /home/christian/Documentos/Lab_sopes/201902125_LAB_SO1_2S2025/proyecto-2/modulo-kernel && sudo insmod sysinfo.ko")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Error cargando módulo: %v, output: %s", err, string(output))
	}
	fmt.Printf("Módulo de kernel cargado correctamente: %s\n", string(output))
	return nil
}

func setupCronJob() {

	fmt.Println("Cronjob configurado (simulado por ahora)")
}

func startMonitoringLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for t := range ticker.C {
		fmt.Printf("Ciclo de monitoreo: %s\n", t.Format("15:04:05"))

		// Leer y analizar información de /proc
		err := monitorProcesses()
		if err != nil {
			log.Printf("Error en monitoreo: %v", err)
		}
	}
}
