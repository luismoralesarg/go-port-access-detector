package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type ConnectionAttempt struct {
	SourceIP string `json:"source_ip"`
	Port     string `json:"port"`
	Message  string `json:"message"`
}

var attempts []ConnectionAttempt

// Función para enviar notificaciones a NTFY
func sendToNTFY(topic, message string) error {
	url := fmt.Sprintf("https://ntfy.sh/%s", topic)
	req, err := http.NewRequest("POST", url, strings.NewReader(message))
	if err != nil {
		return err
	}
	req.Header.Set("Title", "Intento de conexión detectado")
	req.Header.Set("Priority", "high")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		return fmt.Errorf("error enviando notificación: %s", buf.String())
	}
	return nil
}

// Función para seguir los logs del sistema
func followLogs(filePath, topic string) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Error abriendo el archivo: %v", err)
	}
	defer file.Close()

	re := regexp.MustCompile(`SRC=(\d+\.\d+\.\d+\.\d+).*DPT=(\d+)`)
	scanner := bufio.NewScanner(file)

	// Ir al final del archivo
	file.Seek(0, os.SEEK_END)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "INTENTO:") {
			matches := re.FindStringSubmatch(line)
			if len(matches) == 3 {
				attempt := ConnectionAttempt{
					SourceIP: matches[1],
					Port:     matches[2],
					Message:  line,
				}
				attempts = append(attempts, attempt)

				// Enviar notificación a NTFY
				message := fmt.Sprintf("IP: %s intentó acceder al puerto %s", attempt.SourceIP, attempt.Port)
				if err := sendToNTFY(topic, message); err != nil {
					log.Printf("Error enviando notificación: %v", err)
				}
			}
		}
	}
}

// Endpoint de la API para consultar intentos
func getAttempts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(attempts)
}

func main() {
	// Configuración del tópico de NTFY
	ntfyTopic := "mi-topico-ntfy"
	logFilePath := "/var/log/syslog" // Cambiar si es diferente en tu sistema

	// Lanzar la captura de logs en una goroutine
	go followLogs(logFilePath, ntfyTopic)

	// Configurar los endpoints de la API
	http.HandleFunc("/connections", getAttempts)

	fmt.Println("Servidor corriendo en http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
