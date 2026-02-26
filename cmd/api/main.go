package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cprt-lis/internal/config"
	"cprt-lis/internal/db"
	apphttp "cprt-lis/internal/http"
	"cprt-lis/internal/repository/postgres"
	"cprt-lis/internal/service"
)

func main() {
	cfg := config.Load()

	pool, err := db.NewPostgresPool(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db connection failed: %v", err)
	}
	defer pool.Close()

	userRepo := postgres.NewUserRepository(pool)
	patientRepo := postgres.NewPatientRepository(pool)
	billingRepo := postgres.NewBillingRepository(pool)
	orderRepo := postgres.NewOrderRepository(pool)
	labRepo := postgres.NewLabRepository(pool)

	authService := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTTTLMinutes)
	userService := service.NewUserService(userRepo)
	patientService := service.NewPatientService(patientRepo)
	billingService := service.NewBillingService(billingRepo)
	orderService := service.NewOrderService(orderRepo)
	labService := service.NewLabService(labRepo)

	handlers := apphttp.NewHandlers(authService, userService, patientService, billingService, orderService, labService, cfg)
	e := apphttp.NewRouter(handlers)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
		Handler:      e,
	}

	go func() {
		log.Printf("server listening on :%s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown error: %v", err)
	}
}
