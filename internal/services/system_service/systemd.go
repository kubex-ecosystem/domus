// Package systemservice provides a systemd-compatible supervision service for the database container.
package systemservice

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"sync/atomic"
	"time"

	"github.com/kubex-ecosystem/domus/internal/interfaces"
	"github.com/kubex-ecosystem/domus/internal/module/kbx"
	"github.com/kubex-ecosystem/domus/internal/types"
	"github.com/kubex-ecosystem/logz"
	_ "github.com/lib/pq" // se for postgres, por exemplo
)

type Config struct {
	ContainerName        string
	DBAddr               string // "127.0.0.1:5432"
	DBDSN                string // opcional, pra SELECT 1
	CheckInterval        time.Duration
	MaxFailBeforeRestart int
	MaxRestartFail       int
}

var lastOK atomic.Value

func StartSystemService(args *kbx.InitArgs) error {
	if args == nil {
		args = &kbx.InitArgs{}
	}
	logz.SetDebugMode(args.Debug)
	sigChan := make(chan string, 1)
	signalManager := types.NewSignalManager(sigChan, logz.GetLoggerZ("ds"))

	go func(sm interfaces.ISignalManager[chan string]) {
		if err := sm.ListenForSignals(); err != nil {
			log.Printf("[ERROR] signal manager failed: %v", err)
		}
	}(signalManager)

	cfg := Config{
		ContainerName:        "gnyx-db",
		DBAddr:               "127.0.0.1:5432",
		DBDSN:                "", // se quiser SQL real
		CheckInterval:        3 * time.Second,
		MaxFailBeforeRestart: 3,
		MaxRestartFail:       2,
	}
	userID := os.Getuid()
	dockerSocketPath := "/run/user/" + logz.Sprintf("%d", userID) + "/docker.sock"
	if _, err := os.Stat(dockerSocketPath); err == nil {
		os.Setenv("DOCKER_HOST", "unix://"+dockerSocketPath)
	}
	if err := ensureDBUp(cfg); err != nil {
		return logz.Errorf("failed to bring DB up: %v", err)
	}

	lastOK.Store(time.Now())

	go superviseLoop(cfg)
	go startHealthServer()

	msg := <-sigChan
	log.Printf("[INFO] received signal: %s", msg)

	signalManager.StopListening()
	return nil
}

func ensureDBUp(cfg Config) error {
	if !isContainerRunning(cfg.ContainerName) {
		if err := dockerCmd("start", cfg.ContainerName); err != nil {
			return err
		}
	}

	// readiness: tcp (e opcionalmente SQL)
	deadline := time.Now().Add(30 * time.Second)
	for {
		if time.Now().After(deadline) {
			return errors.New("db did not become ready in time")
		}
		if checkDBTCP(cfg.DBAddr) == nil && checkDBSQL(cfg.DBDSN) == nil {
			return nil
		}
		time.Sleep(2 * time.Second)
	}
}

func superviseLoop(cfg Config) {
	failSeq := 0
	restartFailures := 0

	for {
		time.Sleep(cfg.CheckInterval)
		if err := checkDBTCP(cfg.DBAddr); err != nil || checkDBSQL(cfg.DBDSN) != nil {
			failSeq++
			log.Printf("[WARN] DB healthcheck failed (%d): %v", failSeq, err)
		} else {
			failSeq = 0
			lastOK.Store(time.Now())
			continue
		}

		if failSeq >= cfg.MaxFailBeforeRestart {
			log.Printf("[INFO] restarting DB container %s", cfg.ContainerName)
			if err := dockerCmd("restart", cfg.ContainerName); err != nil {
				log.Printf("[ERROR] docker restart failed: %v", err)
				restartFailures++
			} else {
				// re-wait readiness
				if err := ensureDBUp(cfg); err != nil {
					log.Printf("[ERROR] DB did not recover after restart: %v", err)
					restartFailures++
				} else {
					failSeq = 0
					restartFailures = 0
					continue
				}
			}
		}

		if restartFailures >= cfg.MaxRestartFail {
			log.Printf("[FATAL] DB unhealthy after restarts, exiting DS so systemd can restart everything")
			os.Exit(1)
		}
	}
}

func isContainerRunning(name string) bool {
	cmd := exec.Command("docker", "inspect", "-f", "{{.State.Running}}", name)
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return string(out)[:4] == "true"
}

func dockerCmd(args ...string) error {
	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func checkDBTCP(addr string) error {
	if addr == "" {
		return nil
	}
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return err
	}
	_ = conn.Close()
	return nil
}

func checkDBSQL(dsn string) error {
	if dsn == "" {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	return db.PingContext(ctx)
}

func startHealthServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		tAny := lastOK.Load()
		t, _ := tAny.(time.Time)
		if t.IsZero() || time.Since(t) > 10*time.Second {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"degraded"}`))
			return
		}
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	s := &http.Server{
		Addr:    "127.0.0.1:9123",
		Handler: mux,
	}
	if err := s.ListenAndServe(); err != nil {
		log.Printf("[WARN] health server stopped: %v", err)
	}
}
