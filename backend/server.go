package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"socialnetwork/pkg/db/sqlite"
	web "socialnetwork/web"
)

func main() {

	dbPath := "data/socialnetwork.db"

	if err := os.MkdirAll(filepath.Dir(dbPath), os.ModePerm); err != nil {
		log.Fatal("failed to create db folder:", err)
	}

	db, err := sqlite.ConnectAndMigrate(dbPath, "pkg/db/migrations/sqlite")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fmt.Println("✅ Database tables checked and initialized.")

	// Determine frontend directory robustly so backend can be started from repo root or backend/
	cwd, _ := os.Getwd()
	candidates := []string{
		filepath.Join(cwd, "frontend-next"),       // run from repo root
		filepath.Join(cwd, "..", "frontend-next"), // run from backend/
		filepath.Join(cwd, "backend", "frontend-next"),
	}
	var frontendPath string
	for _, c := range candidates {
		if fi, err := os.Stat(c); err == nil && fi.IsDir() {
			frontendPath = c
			break
		}
	}
	if frontendPath == "" {
		// fallback to ../frontend-next (old behavior) — may still fail if path missing
		frontendPath = "../frontend-next"
	}

	// Export STATIC_DIR for handlers that need to save files into the frontend public folder
	absFrontend, _ := filepath.Abs(frontendPath)
	_ = os.Setenv("STATIC_DIR", absFrontend)

	// Ensure public upload directories exist so file writes don't fail on fresh clones
	_ = os.MkdirAll(filepath.Join(absFrontend, "public", "img", "posts"), 0o755)
	_ = os.MkdirAll(filepath.Join(absFrontend, "public", "img", "avatars"), 0o755)

	fs := http.FileServer(http.Dir(frontendPath))
	http.Handle("/", fs)
	fmt.Printf("✅ Serving static files from %s\n", frontendPath)

	web.ConnectWeb(db)

	fmt.Println("🚀 Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
