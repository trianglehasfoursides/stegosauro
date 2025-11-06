package main

import (
	"context"
	"errors"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/keygen"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	godotenv.Load() // error gak papa

	dsn := os.Getenv("postgres")

	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt: false,
	})

	if err != nil {
		log.Error(err.Error())
		return
	}

	db.AutoMigrate(&post{})

	kp, err := keygen.New(
		"awesome",
		keygen.WithPassphrase(os.Getenv("secret")),
		keygen.WithKeyType(keygen.Ed25519),
	)

	if err != nil {
		log.Error(err.Error())
		return
	}

	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPEM(kp.RawPrivateKey()),
		wish.WithMiddleware(
			bubbletea.Middleware(handler),
			activeterm.Middleware(),
			logging.Middleware(),
		),
	)

	if err != nil {
		log.Error("Could not start server", "error", err)
		return
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Info("Starting SSH server", "host", host, "port", port)
	go func() {
		if err = s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Error("Could not start server", "error", err)
			done <- nil
		}
	}()

	<-done
	log.Info("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()
	if err := s.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Error("Could not stop server", "error", err)
		return
	}
}
