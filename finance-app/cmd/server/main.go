package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"finance-app/internal/handlers"
	"finance-app/internal/repository"
	"finance-app/internal/services"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	repo, err := repository.NewExcelRepository("data/finanzas.xlsx")
	if err != nil {
		return err
	}

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	accountService := services.NewAccountService(repo.Accounts)
	handlers.NewAccountHandler(accountService).RegisterRoutes(mux)
	categoryService := services.NewCategoryService(repo.Categories)
	handlers.NewCategoryHandler(categoryService).RegisterRoutes(mux)
	transactionService := services.NewTransactionService(repo.Transactions, repo.Accounts, repo.Categories)
	handlers.NewTransactionHandler(transactionService).RegisterRoutes(mux)
	transferService := services.NewTransferService(repo.Transfers, repo.Accounts)
	handlers.NewTransferHandler(transferService).RegisterRoutes(mux)
	dashboardService := services.NewDashboardService(repo.Accounts, repo.Categories, repo.Transactions, repo.Transfers)
	handlers.NewDashboardHandler(dashboardService).RegisterRoutes(mux)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte("{\"status\":\"ok\"}\n")); err != nil {
			log.Printf("Error al responder /health: %v", err)
		}
	})

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("Iniciando servidor HTTP en %s", server.Addr)
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	case <-ctx.Done():
		stop()
		log.Print("Cerrando servidor HTTP")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		if closeErr := server.Close(); closeErr != nil {
			log.Printf("Error al forzar el cierre del servidor: %v", closeErr)
		}
		return err
	}
	if err := <-serverErrors; !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	log.Print("Servidor HTTP detenido")
	return nil
}
