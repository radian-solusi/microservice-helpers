package helpers

import (
	"log"
	"os"
	"path/filepath"
	"time"
)

func isProduction() bool { return os.Getenv("GIN_MODE") == "release" }

func (h *Helpers) SetupLogging() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	if !h.IsProduction() {
		log.SetOutput(os.Stdout)
		return
	}
	if err := os.MkdirAll("logs", 0o755); err != nil {
		log.SetOutput(os.Stdout)
		log.Printf("create logs directory: %v", err)
		return
	}
	name := filepath.Join("logs", "app-"+time.Now().Format("2006-01-02")+".log")
	f, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	if err != nil {
		log.SetOutput(os.Stdout)
		log.Printf("open log file: %v", err)
		return
	}
	log.SetOutput(f)
}
