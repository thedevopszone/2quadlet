package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

	if *help || *podmanCmd == "" {
		printUsage()
		return
	}

	config, err := parsePodmanCommand(*podmanCmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fehler beim Parsen des Kommandos: %v\n", err)
		os.Exit(1)
	}

	quadletContent := generateQuadletFile(config)
	
	filename := config.ContainerName + ".container"
	if config.ContainerName == "" {
		filename = "container.container"
	}
	
	outputPath := filepath.Join(*outputDir, filename)
	
	err = os.WriteFile(outputPath, []byte(quadletContent), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fehler beim Schreiben der Datei: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Quadlet-Datei erstellt: %s\n", outputPath)
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

func parsePodmanCommand(cmd string) (*QuadletConfig, error) {
	config := &QuadletConfig{}
	
	// Entferne "podman run" vom Anfang
	cmd = strings.TrimSpace(cmd)
	if strings.HasPrefix(cmd, "podman run") {
		cmd = strings.TrimSpace(cmd[10:])
	}
	
	// Einfacher Parser für die wichtigsten Optionen
	args := splitCommand(cmd)
	
	i := 0
	for i < len(args) {
		arg := args[i]
		
		switch {
		case arg == "-d" || arg == "--detach":
			// Wird automatisch von systemd gehandhabt
			i++
		case arg == "--name":
			if i+1 < len(args) {
				config.ContainerName = args[i+1]
				i += 2
			} else {
				i++
			}
		case strings.HasPrefix(arg, "--name="):
			config.ContainerName = arg[7:]
			i++
		case arg == "-p" || arg == "--publish":
			if i+1 < len(args) {
				config.Ports = append(config.Ports, args[i+1])
				i += 2
			} else {
				i++
			}
		case strings.HasPrefix(arg, "-p=") || strings.HasPrefix(arg, "--publish="):
			port := strings.SplitN(arg, "=", 2)[1]
			config.Ports = append(config.Ports, port)
			i++
		case arg == "-v" || arg == "--volume":
			if i+1 < len(args) {
				config.Volumes = append(config.Volumes, args[i+1])
				i += 2
			} else {
				i++
			}
		case strings.HasPrefix(arg, "-v=") || strings.HasPrefix(arg, "--volume="):
			volume := strings.SplitN(arg, "=", 2)[1]
			config.Volumes = append(config.Volumes, volume)
			i++
		case arg == "-e" || arg == "--env":
			if i+1 < len(args) {
				config.Environment = append(config.Environment, args[i+1])
				i += 2
			} else {
				i++
			}
		case strings.HasPrefix(arg, "-e=") || strings.HasPrefix(arg, "--env="):
			env := strings.SplitN(arg, "=", 2)[1]
			config.Environment = append(config.Environment, env)
			i++
		case arg == "--network":
			if i+1 < len(args) {
				config.Network = args[i+1]
				i += 2
			} else {
				i++
			}
		case strings.HasPrefix(arg, "--network="):
			config.Network = arg[10:]
			i++
		case arg == "-u" || arg == "--user":
			if i+1 < len(args) {
				config.User = args[i+1]
				i += 2
			} else {
				i++
			}
		case strings.HasPrefix(arg, "-u=") || strings.HasPrefix(arg, "--user="):
			user := strings.SplitN(arg, "=", 2)[1]
			config.User = user
			i++
		case arg == "-w" || arg == "--workdir":
			if i+1 < len(args) {
				config.WorkingDir = args[i+1]
				i += 2
			} else {
				i++
			}
		case strings.HasPrefix(arg, "-w=") || strings.HasPrefix(arg, "--workdir="):
			workdir := strings.SplitN(arg, "=", 2)[1]
			config.WorkingDir = workdir
			i++
		case strings.HasPrefix(arg, "-"):
			// Unbekannte Option überspringen
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				i += 2 // Option mit Wert
			} else {
				i++ // Flag ohne Wert
			}
		default:
			// Das sollte das Image sein
			if config.Image == "" {
				config.Image = arg
				i++
				// Alles danach sind Kommando-Argumente
				for i < len(args) {
					config.Command = append(config.Command, args[i])
					i++
				}
			} else {
				i++
			}
		}
	}
	
	if config.Image == "" {
		return nil, fmt.Errorf("kein Image im Kommando gefunden")
	}
	
	return config, nil
}

func splitCommand(cmd string) []string {
	// Einfacher Kommando-Splitter der Anführungszeichen berücksichtigt
	var args []string
	var current strings.Builder
	inQuotes := false
	quoteChar := byte(0)
	
	for i := 0; i < len(cmd); i++ {
		char := cmd[i]
		
		if !inQuotes && (char == '"' || char == '\'') {
			inQuotes = true
			quoteChar = char
		} else if inQuotes && char == quoteChar {
			inQuotes = false
			quoteChar = 0
		} else if !inQuotes && char == ' ' {
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		} else {
			current.WriteByte(char)
		}
	}
	
	if current.Len() > 0 {
		args = append(args, current.String())
	}
	
	return args
}

func generateQuadletFile(config *QuadletConfig) string {
	var builder strings.Builder
	
	builder.WriteString("[Unit]\n")
	builder.WriteString("Description=Container " + config.ContainerName + "\n")
	builder.WriteString("Wants=network-online.target\n")
	builder.WriteString("After=network-online.target\n")
	builder.WriteString("RequiresMountsFor=%t/containers\n")
	builder.WriteString("\n")
	
	builder.WriteString("[Container]\n")
	builder.WriteString("Image=" + config.Image + "\n")
	
	if config.ContainerName != "" {
		builder.WriteString("ContainerName=" + config.ContainerName + "\n")
	}
	
	for _, port := range config.Ports {
		builder.WriteString("PublishPort=" + port + "\n")
	}
	
	for _, volume := range config.Volumes {
		builder.WriteString("Volume=" + volume + "\n")
	}
	
	for _, env := range config.Environment {
		builder.WriteString("Environment=" + env + "\n")
	}
	
	if len(config.Command) > 0 {
		builder.WriteString("Exec=" + strings.Join(config.Command, " ") + "\n")
	}
	
	if config.Network != "" && config.Network != "default" {
		builder.WriteString("Network=" + config.Network + "\n")
	}
	
	if config.User != "" {
		builder.WriteString("User=" + config.User + "\n")
	}
	
	if config.WorkingDir != "" {
		builder.WriteString("WorkingDir=" + config.WorkingDir + "\n")
	}
	
	// Standard Quadlet Einstellungen
	builder.WriteString("AutoUpdate=registry\n")
	builder.WriteString("Pull=newer\n")
	builder.WriteString("\n")
	
	builder.WriteString("[Service]\n")
	builder.WriteString("Restart=always\n")
	builder.WriteString("TimeoutStartSec=900\n")
	builder.WriteString("\n")
	
	builder.WriteString("[Install]\n")
	builder.WriteString("WantedBy=multi-user.target default.target\n")
	
	return builder.String()
}