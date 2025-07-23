package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// ComposeFile repräsentiert eine Docker Compose-Datei
type ComposeFile struct {
	Version  string                 `yaml:"version,omitempty"`
	Services map[string]Service     `yaml:"services"`
	Networks map[string]Network     `yaml:"networks,omitempty"`
	Volumes  map[string]Volume      `yaml:"volumes,omitempty"`
}

// Service repräsentiert einen Service in Docker Compose
type Service struct {
	Image         string            `yaml:"image,omitempty"`
	ContainerName string            `yaml:"container_name,omitempty"`
	Ports         []string          `yaml:"ports,omitempty"`
	Volumes       []string          `yaml:"volumes,omitempty"`
	Environment   []string          `yaml:"environment,omitempty"`
	EnvFile       []string          `yaml:"env_file,omitempty"`
	DependsOn     []string          `yaml:"depends_on,omitempty"`
	Networks      []string          `yaml:"networks,omitempty"`
	Restart       string            `yaml:"restart,omitempty"`
	Command       interface{}       `yaml:"command,omitempty"`
	WorkingDir    string            `yaml:"working_dir,omitempty"`
	User          string            `yaml:"user,omitempty"`
	Labels        map[string]string `yaml:"labels,omitempty"`
}

// Network repräsentiert ein Netzwerk in Docker Compose
type Network struct {
	Driver string `yaml:"driver,omitempty"`
}

// Volume repräsentiert ein Volume in Docker Compose
type Volume struct {
	Driver string `yaml:"driver,omitempty"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: compose-to-quadlet <docker-compose.yml> [output-directory]")
		fmt.Println("Beispiel: compose-to-quadlet docker-compose.yml ./quadlet")
		os.Exit(1)
	}

	inputFile := os.Args[1]
	outputDir := "./quadlet"
	if len(os.Args) >= 3 {
		outputDir = os.Args[2]
	}

	// Ausgabeverzeichnis erstellen
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Fehler beim Erstellen des Ausgabeverzeichnisses: %v", err)
	}

	// Docker Compose-Datei lesen und parsen
	compose, err := parseComposeFile(inputFile)
	if err != nil {
		log.Fatalf("Fehler beim Parsen der Compose-Datei: %v", err)
	}

	// Netzwerke konvertieren
	if err := convertNetworks(compose.Networks, outputDir); err != nil {
		log.Printf("Warnung bei Netzwerk-Konvertierung: %v", err)
	}

	// Volumes konvertieren
	if err := convertVolumes(compose.Volumes, outputDir); err != nil {
		log.Printf("Warnung bei Volume-Konvertierung: %v", err)
	}

	// Services konvertieren
	if err := convertServices(compose.Services, outputDir); err != nil {
		log.Fatalf("Fehler bei Service-Konvertierung: %v", err)
	}

	fmt.Printf("Konvertierung abgeschlossen! Quadlet-Dateien wurden in '%s' erstellt.\n", outputDir)
	fmt.Println("\nUm die Container zu starten:")
	fmt.Printf("1. Kopiere die .container, .network und .volume Dateien nach ~/.config/containers/systemd/\n")
	fmt.Printf("2. Führe aus: systemctl --user daemon-reload\n")
	fmt.Printf("3. Starte Services mit: systemctl --user start <service-name>\n")
}

func parseComposeFile(filename string) (*ComposeFile, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("Datei konnte nicht gelesen werden: %v", err)
	}

	var compose ComposeFile
	if err := yaml.Unmarshal(data, &compose); err != nil {
		return nil, fmt.Errorf("YAML konnte nicht geparst werden: %v", err)
	}

	return &compose, nil
}

func convertNetworks(networks map[string]Network, outputDir string) error {
	for name, network := range networks {
		filename := filepath.Join(outputDir, name+".network")
		content := generateNetworkFile(name, network)
		
		if err := ioutil.WriteFile(filename, []byte(content), 0644); err != nil {
			return fmt.Errorf("Netzwerk-Datei konnte nicht geschrieben werden: %v", err)
		}
		fmt.Printf("Erstellt: %s\n", filename)
	}
	return nil
}

func convertVolumes(volumes map[string]Volume, outputDir string) error {
	for name, volume := range volumes {
		filename := filepath.Join(outputDir, name+".volume")
		content := generateVolumeFile(name, volume)
		
		if err := ioutil.WriteFile(filename, []byte(content), 0644); err != nil {
			return fmt.Errorf("Volume-Datei konnte nicht geschrieben werden: %v", err)
		}
		fmt.Printf("Erstellt: %s\n", filename)
	}
	return nil
}

func convertServices(services map[string]Service, outputDir string) error {
	for name, service := range services {
		filename := filepath.Join(outputDir, name+".container")
		content := generateContainerFile(name, service)
		
		if err := ioutil.WriteFile(filename, []byte(content), 0644); err != nil {
			return fmt.Errorf("Container-Datei konnte nicht geschrieben werden: %v", err)
		}
		fmt.Printf("Erstellt: %s\n", filename)
	}
	return nil
}

func generateNetworkFile(name string, network Network) string {
	var content strings.Builder
	
	content.WriteString("[Network]\n")
	if network.Driver != "" {
		content.WriteString(fmt.Sprintf("Driver=%s\n", network.Driver))
	}
	content.WriteString(fmt.Sprintf("NetworkName=%s\n", name))
	
	return content.String()
}

func generateVolumeFile(name string, volume Volume) string {
	var content strings.Builder
	
	content.WriteString("[Volume]\n")
	if volume.Driver != "" {
		content.WriteString(fmt.Sprintf("Driver=%s\n", volume.Driver))
	}
	content.WriteString(fmt.Sprintf("VolumeName=%s\n", name))
	
	return content.String()
}

func generateContainerFile(name string, service Service) string {
	var content strings.Builder
	
	content.WriteString("[Unit]\n")
	content.WriteString(fmt.Sprintf("Description=%s container\n", name))
	
	// Abhängigkeiten
	if len(service.DependsOn) > 0 {
		var after []string
		for _, dep := range service.DependsOn {
			after = append(after, dep+".service")
		}
		content.WriteString(fmt.Sprintf("After=%s\n", strings.Join(after, " ")))
		content.WriteString(fmt.Sprintf("Requires=%s\n", strings.Join(after, " ")))
	}
	
	content.WriteString("\n[Container]\n")
	
	// Image
	if service.Image != "" {
		content.WriteString(fmt.Sprintf("Image=%s\n", service.Image))
	}
	
	// Container Name
	containerName := service.ContainerName
	if containerName == "" {
		containerName = name
	}
	content.WriteString(fmt.Sprintf("ContainerName=%s\n", containerName))
	
	// Ports
	for _, port := range service.Ports {
		content.WriteString(fmt.Sprintf("PublishPort=%s\n", port))
	}
	
	// Volumes
	for _, volume := range service.Volumes {
		content.WriteString(fmt.Sprintf("Volume=%s\n", volume))
	}
	
	// Environment Variables
	for _, env := range service.Environment {
		content.WriteString(fmt.Sprintf("Environment=%s\n", env))
	}
	
	// Environment Files
	for _, envFile := range service.EnvFile {
		content.WriteString(fmt.Sprintf("EnvironmentFile=%s\n", envFile))
	}
	
	// Networks
	for _, network := range service.Networks {
		content.WriteString(fmt.Sprintf("Network=%s.network\n", network))
	}
	
	// Working Directory
	if service.WorkingDir != "" {
		content.WriteString(fmt.Sprintf("WorkingDir=%s\n", service.WorkingDir))
	}
	
	// User
	if service.User != "" {
		content.WriteString(fmt.Sprintf("User=%s\n", service.User))
	}
	
	// Command
	if service.Command != nil {
		switch cmd := service.Command.(type) {
		case string:
			if cmd != "" {
				content.WriteString(fmt.Sprintf("Exec=%s\n", cmd))
			}
		case []interface{}:
			var cmdParts []string
			for _, part := range cmd {
				cmdParts = append(cmdParts, fmt.Sprintf("%v", part))
			}
			if len(cmdParts) > 0 {
				content.WriteString(fmt.Sprintf("Exec=%s\n", strings.Join(cmdParts, " ")))
			}
		}
	}
	
	// Labels
	if len(service.Labels) > 0 {
		// Labels sortieren für konsistente Ausgabe
		var labelKeys []string
		for key := range service.Labels {
			labelKeys = append(labelKeys, key)
		}
		sort.Strings(labelKeys)
		
		for _, key := range labelKeys {
			content.WriteString(fmt.Sprintf("Label=%s=%s\n", key, service.Labels[key]))
		}
	}
	
	content.WriteString("\n[Service]\n")
	
	// Restart-Policy
	restart := service.Restart
	if restart == "" {
		restart = "no"
	}
	
	switch restart {
	case "always":
		content.WriteString("Restart=always\n")
	case "unless-stopped":
		content.WriteString("Restart=always\n")
	case "on-failure":
		content.WriteString("Restart=on-failure\n")
	default:
		content.WriteString("Restart=no\n")
	}
	
	content.WriteString("\n[Install]\n")
	content.WriteString("WantedBy=multi-user.target default.target\n")
	
	return content.String()
}