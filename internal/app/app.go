package app

import (
	"bufio"
	"crypto/tls"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/tp_security/internal/config"
	"github.com/tp_security/internal/controller"
	"github.com/tp_security/internal/handler"
	"github.com/tp_security/internal/middleware"
	"github.com/tp_security/internal/repository"

	_ "github.com/jackc/pgx"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func Run(cfg *config.Config) error {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.User,
		cfg.DB.Pass,
		cfg.DB.Name,
		cfg.DB.SSLMode,
	)
	db, err := startPostgres(connStr)
	if err != nil {
		return err
	}

	repo := repository.New(db)

	handleFunc, err := handler.New(cfg, repo)
	if err != nil {
		return err
	}

	server := http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      middleware.AccessLog(handleFunc),
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	log.Printf("Starting HTTP/HTTPS proxy server on :%s", cfg.Port)
	if err := server.ListenAndServe(); err != nil {
		return err
	}

	return nil
}

func RunApi(cfg *config.Config) error {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.User,
		cfg.DB.Pass,
		cfg.DB.Name,
		cfg.DB.SSLMode,
	)
	db, err := startPostgres(connStr)
	if err != nil {
		return err
	}

	file, err := os.Open("dicc.txt")
	if err != nil {
		return err
	}
	defer file.Close()
	var paths []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		paths = append(paths, scanner.Text())
	}

	repo := repository.New(db)
	contr := controller.New(repo, paths)

	handleFunc := handler.NewApi(cfg, contr)

	server := http.Server{
		Addr:         ":" + cfg.PortAPI,
		Handler:      middleware.AccessLog(handleFunc),
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	log.Printf("Starting api server on :%s", cfg.PortAPI)
	if err := server.ListenAndServe(); err != nil {
		return err
	}

	return nil
}

func startPostgres(connStr string) (*sql.DB, error) {
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, fmt.Errorf("postgres connect: %w", err)
	}
	db.SetMaxOpenConns(10)

	retrying := 10
	i := 1
	log.Printf("try ping postgresql:%v", i)
	for err = db.Ping(); err != nil; err = db.Ping() {
		if i >= retrying {
			return nil, fmt.Errorf("postgres connect: %w", err)
		}
		i++
		time.Sleep(1 * time.Second)
		log.Printf("try ping postgresql: %v", i)
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS data (id SERIAL PRIMARY KEY, request JSON, response JSON)")
	if err != nil {
		log.Printf(err.Error())
		return nil, fmt.Errorf("create table: %w", err)
	}

	return db, nil
}
