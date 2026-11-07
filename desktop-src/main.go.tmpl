package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"

	appconfig "github.com/SoufianoDev/OhMyFreeAI/internal/application/config"
	"github.com/SoufianoDev/OhMyFreeAI/internal/infrastructure/sqlite"
	"github.com/SoufianoDev/OhMyFreeAI/migrations"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	configDirectory, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("resolve application config directory: %w", err)
	}
	databasePath := filepath.Join(configDirectory, "OhMyFreeAI", "state.db")

	startupContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	store, err := sqlite.Open(startupContext, databasePath)
	if err != nil {
		return fmt.Errorf("open application database: %w", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			log.Printf("close application database: %v", err)
		}
	}()

	if err := store.Migrate(startupContext, migrations.Files, "."); err != nil {
		return fmt.Errorf("migrate application database: %w", err)
	}

	configuration := appconfig.Service{Repository: store}
	service := NewAppService(configuration)

	app := application.New(application.Options{
		Name:        "OhMyFreeAI",
		Description: "Free local-first AI workspace",
		Services: []application.Service{
			application.NewService(service),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
	})

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "OhMyFreeAI",
		Width:            1280,
		Height:           820,
		BackgroundColour: application.NewRGB(247, 248, 250),
		URL:              "/",
	})

	if err := app.Run(); err != nil {
		return fmt.Errorf("run desktop application: %w", err)
	}
	return nil
}
