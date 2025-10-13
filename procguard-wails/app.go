package main

import (
	"context"
	"database/sql"
	"os"
	"procguard-wails/internal/auth"
	"procguard-wails/internal/daemon"
	"procguard-wails/internal/data"
	"procguard-wails/internal/web"
)

// App struct
type App struct {
	ctx context.Context
	db  *sql.DB
	log data.Logger
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	db, err := data.InitDB()
	if err != nil {
		println("Error initializing database:", err.Error())
		return
	}
	a.db = db
	data.NewLogger(db)
	a.log = data.GetLogger()

	daemon.Start(a.log, a.db)

	exePath, err := os.Executable()
	if err != nil {
		a.log.Printf("Error getting executable path: %v", err)
	}
	if err := web.InstallNativeHost(exePath); err != nil {
		a.log.Printf("Failed to install native messaging host: %v\n", err)
	}
}

func (a *App) Search(query, since, until string) ([][]string, error) {
	return data.Search(a.db, query, since, until)
}

func (a *App) GetAppBlocklist() ([]string, error) {
	return data.LoadApp()
}

func (a *App) AddAppBlocklist(names []string) error {
	list, err := data.LoadApp()
	if err != nil {
		return err
	}
	for _, name := range names {
		if !contains(list, name) {
			list = append(list, name)
		}
	}
	return data.SaveApp(list)
}

func (a *App) RemoveAppBlocklist(names []string) error {
	list, err := data.LoadApp()
	if err != nil {
		return err
	}
	var newList []string
	for _, item := range list {
		if !contains(names, item) {
			newList = append(newList, item)
		}
	}
	return data.SaveApp(newList)
}

func (a *App) ClearAppBlocklist() error {
	return data.ClearApp()
}

func (a *App) GetWebBlocklist() ([]string, error) {
	return data.LoadWeb()
}

func (a *App) AddWebBlocklist(domain string) error {
	_, err := data.AddWeb(domain)
	return err
}

func (a *App) RemoveWebBlocklist(domain string) error {
	_, err := data.RemoveWeb(domain)
	return err
}

func (a *App) ClearWebBlocklist() error {
	return data.ClearWeb()
}

func (a *App) GetWebLogs(since, until string) ([][]string, error) {
	return data.GetWebLogs(a.db, since, until)
}

func (a *App) GetAutostartStatus() (bool, error) {
	cfg, err := data.LoadConfig()
	if err != nil {
		return false, err
	}
	return cfg.AutostartEnabled, nil
}

func (a *App) EnableAutostart() error {
	_, err := daemon.EnsureAutostart()
	return err
}

func (a *App) DisableAutostart() error {
	return daemon.RemoveAutostart()
}

func (a *App) HasPassword() (bool, error) {
	cfg, err := data.LoadConfig()
	if err != nil {
		return false, err
	}
	return cfg.PasswordHash != "", nil
}

func (a *App) Login(password string) (bool, error) {
	cfg, err := data.LoadConfig()
	if err != nil {
		return false, err
	}
	return auth.CheckPasswordHash(password, cfg.PasswordHash), nil
}

func (a *App) SetPassword(password string) error {
	cfg, err := data.LoadConfig()
	if err != nil {
		return err
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return err
	}
	cfg.PasswordHash = hash
	return cfg.Save()
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
