package web

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"log"
	"net/http"
	cor "socialnetwork/pkg/apis"
	"socialnetwork/pkg/apis/chat"
	e "socialnetwork/pkg/apis/error"
	"socialnetwork/pkg/apis/like"
	likerepo "socialnetwork/pkg/apis/like/repo"
	p "socialnetwork/pkg/apis/post"
	u "socialnetwork/pkg/apis/user"
	database "socialnetwork/pkg/db"
	"strconv"
	"text/template"
	"time"
)

type Page struct {
	Title string
}

func isAuthenticated(db *sql.DB, r *http.Request) bool {
	userID, loggedIn := u.ValidateSession(db, r)
	return loggedIn && userID > 0
}

func ConnectWeb(db *sql.DB) {
	// // Optionally clear all tables if needed
	// if err := clearAllTables(db); err != nil {
	// 	fmt.Println("Error clearing tables:", err)
	// 	return
	// }

	// Define a handler for all paths (main page and dynamic routes)
	http.HandleFunc("/", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		mainPageHandler(w, r, db)
	}))

	http.HandleFunc("/signup", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			mainPageHandler(w, r, db)
		}

		if r.Method == http.MethodPost {
			u.Register(db, w, r) // Call the user registration function
			return
		}

	}))

	http.HandleFunc("/login", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			mainPageHandler(w, r, db)
		}
		u.Login(db, w, r) // Call the Login function from Register.go
	}))

	http.HandleFunc("/get-posts", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			mainPageHandler(w, r, db)
		}

		p.GetPosts(db, w, r)
	}))

	http.HandleFunc("/get-myPosts", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			mainPageHandler(w, r, db)
		}
		userID, loggedIn := u.ValidateSession(db, r)
		if !loggedIn {
			http.Error(w, "Unauthorized. Please log in.", http.StatusUnauthorized)
			return
		}

		posts, err := database.GetPostsByUserID(db, userID)
		if err != nil {
			e.ErrorHandler(w, r, 500)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(posts)
	}))

	http.HandleFunc("/get-otherPosts/", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			mainPageHandler(w, r, db)
		}
		_, loggedIn := u.ValidateSession(db, r)
		if !loggedIn {
			http.Error(w, "Unauthorized. Please log in.", http.StatusUnauthorized)
			return
		}

		username := strings.TrimPrefix(r.URL.Path, "/get-otherPosts/")
		fmt.Println(username)

		uid, err := database.GetUserID(db, username)
		if err != nil || uid == 0 {
			e.ErrorHandler(w, r, 404)
			return
		}
		fmt.Println("User ID:", uid)
		posts, err := database.GetPostsByUserID(db, uid)
		if err != nil {
			e.ErrorHandler(w, r, 500)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(posts)
	}))

	http.HandleFunc("/create-post", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		if !isAuthenticated(db, r) {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		p.CreatePost(db, w, r) //  This is the API to save posts
	}))

	// likes
	likesRepo := likerepo.NewLikesRepository(db)
	likesService := like.NewLikesService(likesRepo)
	likesController := like.NewLikesController(*likesService)

	http.HandleFunc("/likeDislikePost", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		likesController.LikeDislikePost(w, r, db)
	}))

	http.HandleFunc("/likeDislikeComment", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		likesController.InteractWithComment(w, r, db)
	}))

	http.HandleFunc("/getInteractions", cor.WithCORS(likesController.GetInteractions))

	http.HandleFunc("/comments", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		postIDStr := r.URL.Query().Get("post_id")
		postID, err := strconv.Atoi(postIDStr)
		if err != nil || postID <= 0 {
			e.ErrorHandler(w, r, 400)
			return
		}

		// Fetch post details
		post, err := database.GetPostByPostID(db, postID)
		if err != nil || len(post) == 0 {
			e.ErrorHandler(w, r, 404)
			return
		}

		// Fetch comments
		comments, err := p.GetCommentsByPostID(db, postID)
		if err != nil {
			e.ErrorHandler(w, r, 500)
			return
		}

		// Combine post and comments into a single response
		response := map[string]interface{}{
			"post":     post[0], // Assuming GetPostByPostID returns a slice
			"comments": comments,
		}

		// Return combined response as JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))

	http.HandleFunc("/create-comment", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		p.CreateComment(db, w, r) // Ensure this handles comment creation
	}))

	http.HandleFunc("/category/", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		category := strings.TrimPrefix(r.URL.Path, "/category/")
		fmt.Println(category)
		if category == "Liked" {
			p.GetPostbyIsLiked(db, w, r)
			return
		}
		p.GetPostbyCategory(db, w, r, category)
	}))

	http.HandleFunc("/editGet/", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		username := strings.TrimPrefix(r.URL.Path, "/editGet/")
		fmt.Println(username)

		u.GetProfileHandler(db, w, r, username)

	}))

	http.HandleFunc("/editPost/", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		username := strings.TrimPrefix(r.URL.Path, "/editPost/")
		fmt.Println(username)

		u.UpdateProfileHandler(db, w, r, username)

	}))

	http.HandleFunc("/check-session", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		userID, loggedIn := u.ValidateSession(db, r)
		username, _ := database.GetUsernameUsingID(db, userID)
		response := map[string]interface{}{
			"loggedIn": loggedIn,
			"userID":   userID,
			"username": username,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))

	http.HandleFunc("/logout", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		// Clear the session token cookie
		cookie := &http.Cookie{
			Name:     "session_token",
			Value:    "",
			Expires:  time.Now().Add(-1 * time.Hour), // Expire immediately
			HttpOnly: true,
			Path:     "/",
		}
		http.SetCookie(w, cookie)

		// Invalidate session in the database (optional but recommended)
		cookie, err := r.Cookie("session_token")
		cookieINT, _ := strconv.Atoi(cookie.Value)
		if err == nil {
			// Call a function to delete the session from the database
			err := database.DeleteSession(db, cookieINT)
			if err != nil {
				fmt.Println(" Error deleting session:", err)
				e.ErrorHandler(w, r, 500)
			}
		}

		response := map[string]string{"message": "Logged out successfully"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))

	chatHub := chat.NewHub(db)
	go chatHub.Run()

	http.HandleFunc("/ws", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		chat.ServeWs(chatHub, w, r)
	}))

	http.HandleFunc("/messages", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		userID, loggedIn := u.ValidateSession(db, r)
		if !loggedIn {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		withIDStr := r.URL.Query().Get("with")
		offsetStr := r.URL.Query().Get("offset")
		withID, err := strconv.Atoi(withIDStr)
		if err != nil || withID <= 0 {
			e.ErrorHandler(w, r, 404)
			return
		}
		offset, _ := strconv.Atoi(offsetStr)

		query := `SELECT sender_id, receiver_id, content, created_at FROM messages
	          WHERE (sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)
	          ORDER BY created_at DESC LIMIT 10 OFFSET ?`
		rows, err := db.Query(query, userID, withID, withID, userID, offset)
		if err != nil {
			e.ErrorHandler(w, r, 500)
			return
		}
		defer rows.Close()

		var messages []chat.Frontend
		for rows.Next() {
			var m chat.Frontend
			if err := rows.Scan(&m.From, &m.To, &m.Content, &m.Timestamp); err == nil {
				messages = append(messages, m)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if messages == nil {
			messages = []chat.Frontend{}
		}
		json.NewEncoder(w).Encode(messages)

	}))

	http.HandleFunc("/get-users", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		_, loggedIn := u.ValidateSession(db, r)
		if !loggedIn {
			var empty []map[string]interface{}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(empty)
			return
		}
		if r.Method == http.MethodGet {
			onlineIDs := chatHub.GetOnlineUserIDs() // ✅ Get list of online users
			onlineSet := make(map[int]bool)
			for _, id := range onlineIDs {
				onlineSet[id] = true
			}

			rows, err := db.Query(`SELECT id, username FROM users`)
			if err != nil {
				http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
				e.ErrorHandler(w, r, 500)
				return
			}
			defer rows.Close()

			var users []map[string]interface{}
			for rows.Next() {
				var id int
				var username string
				if err := rows.Scan(&id, &username); err == nil {
					users = append(users, map[string]interface{}{
						"id":       id,
						"username": username,
						"online":   onlineSet[id], // ✅ Add online status
					})
				}
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(users)
		}

	}))

	http.HandleFunc("/error/", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		num, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/error/"))
		if r.Method == http.MethodGet {
			e.ErrorHandler(w, r, num)
		}
		if err != nil {
			e.ErrorHandler(w, r, 500) // If there's an error converting, return a 500 error
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(500)
			return
		}
		e.ErrorHandler(w, r, num)

	}))

	fmt.Println("Listening on: http://localhost:8080/")
	if err := http.ListenAndServe("0.0.0.0:8080", nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func mainPageHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Serve the main page for all paths
	tmpl, err := template.ParseFiles("templates/mainPage")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Pass data to the template if needed
	data := Page{
		Title: "Hello",
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func clearAllTables(db *sql.DB) error {
	tables, err := getTableNames(db)
	if err != nil {
		return err
	}

	for _, table := range tables {
		query := `DELETE FROM ` + table
		_, err := db.Exec(query)
		if err != nil {
			return err
		}
	}

	return nil
}

func getTableNames(db *sql.DB) ([]string, error) {
	query := `SELECT name FROM sqlite_master WHERE type='table'`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}
		tables = append(tables, tableName)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	fmt.Println(tables)

	return tables, nil
}
