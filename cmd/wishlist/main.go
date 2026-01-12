package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"wishlist/internal/wish"

	//docker
	_ "github.com/lib/pq"
)

func main() {
	db, err := sql.Open("postgres",
		"host="+os.Getenv("DB_HOST")+
			" port="+os.Getenv("DB_PORT")+
			" user="+os.Getenv("DB_USER")+
			" password="+os.Getenv("DB_PASSWORD")+
			" dbname="+os.Getenv("DB_NAME")+
			" sslmode=disable",
	)
	if err != nil {
		log.Fatal("failed to open DB: ", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := db.PingContext(ctx); err != nil {
		log.Fatal("failed to connect to DB: ", err)
	}

	repo := wish.NewPostgresRepo(db)
	svc := wish.NewService(repo)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /wishes", wish.CreateWishHandler(svc))
	mux.HandleFunc("GET /wishes/{id}", wish.GetWishHandler(svc))
	mux.HandleFunc("GET /wishes", wish.ListWishHandler(svc))
	mux.HandleFunc("PATCH /wishes/{id}", wish.UpdateWishHandler(svc))
	mux.HandleFunc("DELETE /wishes/{id}", wish.DeleteWishHandler(svc))
	mux.HandleFunc("PATCH /wishes/{id}/buy", wish.BuyWishHandler(svc))
	mux.HandleFunc("GET /wishes/stats", wish.StatsHandler(svc))

	port := os.Getenv("PORT")

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("Server starting on port %s\n", port)
		if err := server.ListenAndServe(); err != nil {
			log.Fatal("server failed: ", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}
	log.Println("Server exited gracefully")
}
