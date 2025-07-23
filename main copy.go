package main

import (
	"flag"
	"fmt"
)


type QuadletConfig struct {
	Image       string
	ContainerName string
	Ports       []string
	Volumes     []string
	Environment []string
	Command     []string
	Network     string
	User        string
	WorkingDir  string
	AutoUpdate  string
	Pull        string
}


func main() {
	var (
		podmanCmd = flag.String("cmd", "", "Podman run Kommando")
		outputDir = flag.String("output", ".", "Ausgabeverzeichnis für Quadlet-Dateien")
		help      = flag.Bool("help", false, "Hilfe anzeigen")
	)
	flag.Parse()

	fmt.Println(*podmanCmd)
	fmt.Println(*outputDir)
	fmt.Println(*help)

	if *help || *podmanCmd == "" {
		printUsage()
		return
	}
}

func printUsage() {
	fmt.Println("Podman zu Quadlet Konverter")
	fmt.Println()
	fmt.Println("Verwendung:")
	fmt.Println("  podman-to-quadlet -cmd 'podman run [optionen] image [kommando]'")
	fmt.Println()
	fmt.Println("Optionen:")
	fmt.Println("  -cmd string     Podman run Kommando (erforderlich)")
	fmt.Println("  -output string  Ausgabeverzeichnis (Standard: aktuelles Verzeichnis)")
	fmt.Println("  -help          Diese Hilfe anzeigen")
	fmt.Println()
	fmt.Println("Beispiel:")
	fmt.Println("  podman-to-quadlet -cmd 'podman run -d --name webapp -p 8080:80 -v /data:/app/data nginx'")
}
