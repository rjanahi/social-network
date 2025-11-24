package web

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	cor "socialnetwork/pkg/apis"
	"socialnetwork/pkg/apis/chat"
	e "socialnetwork/pkg/apis/error"
	g "socialnetwork/pkg/apis/group"
	"socialnetwork/pkg/apis/like"
	likerepo "socialnetwork/pkg/apis/like/repo"
	p "socialnetwork/pkg/apis/post"
	u "socialnetwork/pkg/apis/user"
	database "socialnetwork/pkg/db"
)

type Page struct {
	Title string
}

func isAuthenticated(db *sql.DB, r *http.Request) bool {
	userID, loggedIn := u.ValidateSession(db, r)
	return loggedIn && userID > 0
}

func ConnectWeb(db *sql.DB) {
	// Initialize WebSocket hub first so it can be used by other handlers
	chatHub := chat.NewHub(db)
	go chatHub.Run()

	// // Optionally clear all tables if needed
	// if err := clearAllTables(db); err != nil {
	// 	fmt.Println("Error clearing tables:", err)
	// 	return
	// }

	http.HandleFunc("/api/follow", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		// 1) Authentication
		userID, ok := u.ValidateSession(db, r)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// 2) Parse targetId from query
		q := r.URL.Query().Get("targetId")
		targetID, err := strconv.Atoi(q)
		if err != nil || targetID <= 0 {
			http.Error(w, "Invalid targetId", http.StatusBadRequest)
			return
		}

		// 3) Perform follow or unfollow
		switch r.Method {
		case http.MethodPost:
			_, err = db.Exec(
				`INSERT OR IGNORE INTO userFollow(follower_id, following_id) VALUES (?, ?)`,
				userID, targetID,
			)
		case http.MethodDelete:
			_, err = db.Exec(
				`DELETE FROM userFollow WHERE follower_id = ? AND following_id = ?`,
				userID, targetID,
			)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	}))

	if err := ensureEventPollTables(db, chatHub); err != nil {
		fmt.Println("Error creating event tables:", err)
	}

	http.HandleFunc("/api/follow/counts", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := u.ValidateSession(db, r)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var followers, following int

		// Count who follows you (distinct follower ids)
		err1 := db.QueryRow(`SELECT COUNT(DISTINCT follower_id) FROM userFollow WHERE following_id = ?`, userID).Scan(&followers)

		// Count who you're following (distinct following ids)
		err2 := db.QueryRow(`SELECT COUNT(DISTINCT following_id) FROM userFollow WHERE follower_id = ?`, userID).Scan(&following)

		if err1 != nil || err2 != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int{
			"followers": followers,
			"following": following,
		})
	}))

	http.HandleFunc("/api/users/", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		// Strip prefix to get "{suffix…}"
		path := strings.TrimPrefix(r.URL.Path, "/api/users/")

		// 1) GET /api/users/{id}/isFollowing
		if strings.HasSuffix(path, "/isFollowing") && r.Method == http.MethodGet {
			userID, ok := u.ValidateSession(db, r)
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			idStr := strings.TrimSuffix(path, "/isFollowing")
			targetID, err := strconv.Atoi(idStr)
			if err != nil {
				http.Error(w, "Invalid user ID", http.StatusBadRequest)
				return
			}
			var exists bool
			err = db.QueryRow(
				`SELECT COUNT(*)>0 FROM userFollow WHERE follower_id=? AND following_id=?`,
				userID, targetID,
			).Scan(&exists)
			if err != nil {
				http.Error(w, "Database error", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]bool{"isFollowing": exists})
			return
		}

		// 2) POST/DELETE /api/users/{id}/follow
		if strings.HasSuffix(path, "/follow") {
			userID, ok := u.ValidateSession(db, r)
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			idStr := strings.TrimSuffix(path, "/follow")
			targetID, err := strconv.Atoi(idStr)
			if err != nil {
				http.Error(w, "Invalid user ID", http.StatusBadRequest)
				return
			}

			switch r.Method {
			case http.MethodPost:
				_, err = db.Exec(
					`INSERT OR IGNORE INTO userFollow(follower_id, following_id) VALUES (?, ?)`,
					userID, targetID,
				)
			case http.MethodDelete:
				_, err = db.Exec(
					`DELETE FROM userFollow WHERE follower_id = ? AND following_id = ?`,
					userID, targetID,
				)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			if err != nil {
				http.Error(w, "Database error", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]bool{"success": true})
			return
		}

		// 3) Anything else under /api/users/ → 404
		http.NotFound(w, r)
	}))

	http.HandleFunc("/signup", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			u.Register(db, w, r) // Call the user registration function
			return
		}
	}))

	http.HandleFunc("/login", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		u.Login(db, w, r) // Call the Login function from Register.go
	}))

	http.HandleFunc("/get-posts", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		p.GetPosts(db, w, r)
	}))

	http.HandleFunc("/get-myPosts", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		userID, loggedIn := u.ValidateSession(db, r)
		if !loggedIn {
			http.Error(w, "Unauthorized. Please log in.", http.StatusUnauthorized)
			return
		}
		fmt.Println("User ID:", userID)
		fmt.Println("Logged In:", loggedIn)

		posts, err := database.GetPostsByUserID(db, userID)
		if err != nil {
			e.ErrorHandler(w, r, 500)
			fmt.Println("Error fetching posts for user ID:", userID, "Error:", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"posts":   posts,
		})
	}))

	http.HandleFunc("/get-otherPosts/", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		viewerID, loggedIn := u.ValidateSession(db, r)
		if !loggedIn {
			http.Error(w, "Unauthorized. Please log in.", http.StatusUnauthorized)
			return
		}

		username := strings.TrimPrefix(r.URL.Path, "/get-otherPosts/")
		fmt.Println("Username:", username)

		uid, err := database.GetUserID(db, username)
		if err != nil || uid == 0 {
			e.ErrorHandler(w, r, 404)
			fmt.Println("Error fetching user ID for username:", username)
			return
		}

		// Fetch profile privacy and basic profile fields first
		var existingBio sql.NullString
		var (
			id             int
			fname          string
			lname          string
			email          string
			age            int
			gender         string
			bio            string
			avatar         string
			isPrivate      bool
			followerCount  int
			followingCount int
		)
		err = db.QueryRow(`
				SELECT u.id, u.firstname, u.lastname, u.email, u.age, u.gender, u.bio, u.avatar_url, isPrivate,
					(SELECT COUNT(DISTINCT f.follower_id) FROM userFollow f JOIN users uf ON uf.id = f.follower_id WHERE f.following_id = u.id),
					(SELECT COUNT(DISTINCT f.following_id) FROM userFollow f JOIN users uf2 ON uf2.id = f.following_id WHERE f.follower_id = u.id)
				FROM users u
				WHERE u.username = ?
			`, username).Scan(&id, &fname, &lname, &email, &age, &gender, &existingBio, &avatar, &isPrivate, &followerCount, &followingCount)
		if err != nil {
			http.Error(w, "Profile lookup error", http.StatusInternalServerError)
			fmt.Println("Error fetching profile for username:", username, "Error:", err)
			return
		}
		if existingBio.Valid {
			bio = existingBio.String
		} else {
			bio = ""
		}

		// Determine whether the viewer can see the full profile
		canView := false
		if !isPrivate || viewerID == uid {
			canView = true
		} else {
			// Check if viewer follows the target user
			var cnt int
			err = db.QueryRow(`SELECT COUNT(*) FROM userFollow WHERE follower_id = ? AND following_id = ?`, viewerID, uid).Scan(&cnt)
			if err == nil && cnt > 0 {
				canView = true
			}
		}

		// Determine follow relationship and pending request status between viewer and target
		isFollowing := false
		_ = db.QueryRow(`SELECT COUNT(*) > 0 FROM userFollow WHERE follower_id = ? AND following_id = ?`, viewerID, uid).Scan(&isFollowing)

		var requestStatus string
		_ = db.QueryRow(`SELECT status FROM follow_requests WHERE requester_id = ? AND target_id = ?`, viewerID, uid).Scan(&requestStatus)

		var posts []map[string]interface{}
		if canView {
			rows, err := db.Query(`
				SELECT DISTINCT
					p.id, u.username, p.title, p.content,
					COALESCE(p.privacy_level, 0) AS privacy_level,
					p.imgOrgif,
					p.created_at,
					COALESCE(likes.count, 0)    AS likes_count,
					COALESCE(dislikes.count, 0) AS dislikes_count
				FROM posts p
				JOIN users u ON p.user_id = u.id
				LEFT JOIN (
				SELECT post_id, COUNT(*) AS count FROM likes
				WHERE is_like = 1 AND comment_id IS NULL GROUP BY post_id
				) likes    ON p.id = likes.post_id
				LEFT JOIN (
				SELECT post_id, COUNT(*) AS count FROM likes
				WHERE is_like = 0 AND comment_id IS NULL GROUP BY post_id
				) dislikes ON p.id = dislikes.post_id
				WHERE p.user_id = ?
				AND (
					COALESCE(p.privacy_level,0) = 0                         -- Public
					OR (? = p.user_id)                                         -- Viewer is owner
					OR (COALESCE(p.privacy_level,0) = 1 AND EXISTS (           -- Followers-only
						SELECT 1 FROM userFollow f
						WHERE f.follower_id = ? AND f.following_id = p.user_id
					))
					OR (COALESCE(p.privacy_level,0) = 2 AND EXISTS (           -- Selected followers
						SELECT 1 FROM post_permissions pp
						WHERE pp.post_id = p.id AND pp.user_id = ?
					))
				)
				ORDER BY p.created_at DESC
			`, uid, viewerID, viewerID, viewerID)
			if err != nil {
				e.ErrorHandler(w, r, 500)
				fmt.Println("Error fetching posts for username:", username, "Error:", err)
				return
			}
			defer rows.Close()

			for rows.Next() {
				var (
					postID, privacyLevel, likesCount, dislikesCount int
					postUsername, title, content, imgOrgif          string
					createdAt                                       time.Time
				)
				if err := rows.Scan(
					&postID, &postUsername, &title, &content,
					&privacyLevel, &imgOrgif, &createdAt, &likesCount, &dislikesCount,
				); err != nil {
					e.ErrorHandler(w, r, 500)
					return
				}

				categories, _ := database.GetCategoriesByPostID(db, postID)
				privacyText := map[int]string{0: "Public", 1: "Followers", 2: "Private"}[privacyLevel]

				posts = append(posts, map[string]interface{}{
					"id":             postID,
					"username":       postUsername,
					"title":          title,
					"content":        content,
					"privacy_level":  privacyLevel,
					"privacy_text":   privacyText,
					"imgOrgif":       imgOrgif,
					"categories":     categories,
					"likes_count":    likesCount,
					"dislikes_count": dislikesCount,
					"createdAt":      createdAt.Format("2006-01-02 15:04:05"),
				})
			}
		} else {
			// Viewer cannot view private profile details
			posts = []map[string]interface{}{}
		}
		// (posts already populated above when canView == true)

		// If viewer can see, include followers and following lists; otherwise keep them empty
		var followersData []map[string]interface{}
		var followingData []map[string]interface{}
		if canView {
			// Fetch raw lists from DB
			rawFollowers, _ := database.GetUserFollowers(db, uid)
			rawFollowing, _ := database.GetUserFollowing(db, uid)

			// Deduplicate by id to ensure counts match unique users
			followersMap := make(map[int]map[string]interface{})
			for _, f := range rawFollowers {
				if id, ok := f["id"].(int); ok {
					followersMap[id] = f
				}
			}
			for _, v := range followersMap {
				followersData = append(followersData, v)
			}

			followingMap := make(map[int]map[string]interface{})
			for _, f := range rawFollowing {
				if id, ok := f["id"].(int); ok {
					followingMap[id] = f
				}
			}
			for _, v := range followingMap {
				followingData = append(followingData, v)
			}
		} else {
			followersData = []map[string]interface{}{}
			followingData = []map[string]interface{}{}
		}

		profileObj := map[string]interface{}{
			"id":             id,
			"username":       username,
			"fname":          fname,
			"lname":          lname,
			"email":          map[bool]interface{}{true: email, false: ""}[canView],
			"age":            map[bool]interface{}{true: age, false: 0}[canView],
			"gender":         map[bool]interface{}{true: gender, false: ""}[canView],
			"bio":            map[bool]interface{}{true: bio, false: ""}[canView],
			"avatar":         avatar,
			"isPrivate":      isPrivate,
			"isFollowing":    isFollowing,
			"requestStatus":  requestStatus,
			"followerCount":  followerCount,
			"followingCount": followingCount,
			"followers":      followersData,
			"following":      followingData,
			"canView":        canView,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"profile": profileObj,
			"posts":   posts,
		})
	}))

	http.HandleFunc("/create-post", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		if !isAuthenticated(db, r) {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		p.CreatePost(db, chatHub, w, r) //  This is the API to save posts with WebSocket support
	}))

	// Add this after your /create-post handler
	http.HandleFunc("/get-followers", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Validate session and get user ID
		userID, loggedIn := u.ValidateSession(db, r)
		if !loggedIn {
			http.Error(w, "Unauthorized. Please log in.", http.StatusUnauthorized)
			return
		}

		// Get user's followers (people who follow you)
		followers, err := database.GetUserFollowers(db, userID)
		if err != nil {
			fmt.Println("Error getting followers:", err)
			http.Error(w, "Failed to get followers", http.StatusInternalServerError)
			return
		}

		// Get users you are following
		following, err := database.GetUserFollowing(db, userID)
		if err != nil {
			fmt.Println("Error getting following:", err)
			http.Error(w, "Failed to get following", http.StatusInternalServerError)
			return
		}

		// Combine both lists and remove duplicates (users who follow each other)
		chatUsers := make(map[int]map[string]interface{})

		// Add followers
		for _, follower := range followers {
			if id, ok := follower["id"].(int); ok {
				chatUsers[id] = follower
			}
		}

		// Add following
		for _, followingUser := range following {
			if id, ok := followingUser["id"].(int); ok {
				chatUsers[id] = followingUser
			}
		}

		// Convert map to slice
		var chatUsersList []map[string]interface{}
		for _, user := range chatUsers {
			chatUsersList = append(chatUsersList, user)
		}

		response := map[string]interface{}{
			"followers": followers,
			"following": chatUsersList, // Send combined list as "following" for backward compatibility
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))

	// Get chat history between two users
	http.HandleFunc("/get-chat-history", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Validate session
		userID, loggedIn := u.ValidateSession(db, r)
		if !loggedIn {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Get other user ID from query params
		otherUserIDStr := r.URL.Query().Get("user_id")
		if otherUserIDStr == "" {
			http.Error(w, "user_id parameter required", http.StatusBadRequest)
			return
		}

		otherUserID, err := strconv.Atoi(otherUserIDStr)
		if err != nil {
			http.Error(w, "Invalid user_id", http.StatusBadRequest)
			return
		}

		// Check if users can chat (follow relationship required)
		canChat, err := database.CheckFollowRelationship(db, userID, otherUserID)
		if err != nil {
			http.Error(w, "Error checking permissions", http.StatusInternalServerError)
			return
		}

		if !canChat {
			http.Error(w, "You can only chat with users you follow or who follow you", http.StatusForbidden)
			return
		}

		// Fetch chat history
		messages, err := database.GetChatHistory(db, userID, otherUserID)
		if err != nil {
			fmt.Println("Error fetching chat history:", err)
			http.Error(w, "Failed to fetch chat history", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"messages": messages,
		})
	}))

	// likes
	likesRepo := likerepo.NewLikesRepository(db)
	likesService := like.NewLikesService(likesRepo)
	likesController := like.NewLikesController(*likesService, chatHub)

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

		// Patch: Add 'image' alias for frontend compatibility (like group posts)
		postObj := post[0]
		if img, ok := postObj["imgOrgif"]; ok {
			postObj["image"] = img
		}
		response := map[string]interface{}{
			"post":     postObj,
			"comments": comments,
		}
		// Return combined response as JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))

	http.HandleFunc("/create-comment", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		p.CreateComment(db, chatHub, w, r) // Pass hub for real-time updates
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
			fmt.Println("Invalid method for /editPost/")
			return
		}
		username := strings.TrimPrefix(r.URL.Path, "/editPost/")
		fmt.Println("Update for: " + username)

		u.UpdateProfileHandler(db, w, r, username)
	}))

	// Group API endpoints
	http.HandleFunc("/create-group", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		g.CreateGroup(db, chatHub, w, r)
	}))

	http.HandleFunc("/get-groups", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		g.GetGroups(db, w, r)
	}))

	http.HandleFunc("/get-user-groups", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		g.GetUserGroups(db, w, r)
	}))

	http.HandleFunc("/group-details/", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		groupIDStr := strings.TrimPrefix(r.URL.Path, "/group-details/")
		g.GetGroupDetails(db, w, r, groupIDStr)
	}))

	http.HandleFunc("/invite-to-group", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		g.InviteUserToGroup(db, chatHub, w, r)
	}))

	http.HandleFunc("/respond-group-invitation", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		g.RespondToGroupInvitation(db, chatHub, w, r)
	}))

	http.HandleFunc("/request-join-group", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		g.RequestToJoinGroup(db, chatHub, w, r)
	}))

	http.HandleFunc("/leave-group", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		g.LeaveGroup(db, chatHub, w, r)
	}))

	http.HandleFunc("/get-group-invitations", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		g.GetGroupInvitations(db, w, r)
	}))

	http.HandleFunc("/get-invitable-users", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		g.GetInvitableUsers(db, w, r)
	}))

	http.HandleFunc("/send-bulk-invitations", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		g.SendBulkInvitations(db, chatHub, w, r)
	}))

	http.HandleFunc("/get-sent-invitations", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		g.GetSentInvitations(db, w, r)
	}))

	http.HandleFunc("/get-received-invitations", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		g.GetReceivedInvitations(db, w, r)
	}))

	http.HandleFunc("/cancel-invitation", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		g.CancelInvitation(db, w, r)
	}))

	// New endpoints for pending group join requests
	http.HandleFunc("/get-pending-join-requests", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		g.GetPendingJoinRequests(db, w, r)
	}))

	http.HandleFunc("/respond-join-request", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		g.RespondToJoinRequest(db, chatHub, w, r)
	}))

	http.HandleFunc("/groups/create-post", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		g.CreateGroupPost(db, chatHub, w, r)
	}))

	http.HandleFunc("/groups/comments", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		g.AddGroupPostComment(db, chatHub, w, r)
	}))

	http.HandleFunc("/groups/likes", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		g.ToggleGroupPostLike(db, chatHub, w, r)
	}))

	http.HandleFunc("/groups/dislikes", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		g.ToggleGroupPostDislike(db, chatHub, w, r)
	}))

	// Create an event with a built-in YES/NO poll
	http.HandleFunc("/create-event", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		userID, loggedIn := u.ValidateSession(db, r)
		if !loggedIn {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			GroupID     int    `json:"group_id"`
			Title       string `json:"title"`
			Description string `json:"description"`
			EventDate   string `json:"event_date"` // optional ISO datetime from client
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Bad JSON", http.StatusBadRequest)
			return
		}
		if req.GroupID <= 0 || strings.TrimSpace(req.Title) == "" || strings.TrimSpace(req.Description) == "" {
			http.Error(w, "Missing fields", http.StatusBadRequest)
			return
		}

		// Check requester is accepted member
		var cnt int
		if err := db.QueryRow(`
        SELECT COUNT(*) FROM group_members
        WHERE group_id = ? AND user_id = ? AND status = 'accepted'
    `, req.GroupID, userID).Scan(&cnt); err != nil || cnt == 0 {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		// Insert event (store optional event_date)
		var evDate interface{}
		if strings.TrimSpace(req.EventDate) == "" {
			evDate = nil
		} else {
			evDate = req.EventDate
		}

		res, err := db.Exec(`
		INSERT INTO events (group_id, creator_id, title, description, event_date, created_at)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, req.GroupID, userID, req.Title, req.Description, evDate)
		if err != nil {
			http.Error(w, "DB error", http.StatusInternalServerError)
			return
		}
		id, _ := res.LastInsertId()

		// Get username for WebSocket broadcast
		var username string
		db.QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&username)

		// Query event poll counts and user's vote
		var yesCount, noCount int
		_ = db.QueryRow(`SELECT COUNT(*) FROM event_votes WHERE event_id = ? AND choice='yes'`, id).Scan(&yesCount)
		_ = db.QueryRow(`SELECT COUNT(*) FROM event_votes WHERE event_id = ? AND choice='no'`, id).Scan(&noCount)
		var uv sql.NullString
		_ = db.QueryRow(`SELECT choice FROM event_votes WHERE event_id = ? AND user_id = ?`, id, userID).Scan(&uv)
		var userVote string
		if uv.Valid {
			userVote = uv.String
		}

		// Broadcast new event to all group members via WebSocket
		eventNotification := chat.Frontend{
			Type:      "new_groupEvent",
			From:      userID,
			Username:  username,
			GroupID:   req.GroupID,
			PostId:    int(id),
			Content:   req.Title + ": " + req.Description,
			EventDate: req.EventDate,
			Timestamp: time.Now(),
			YesCount:  yesCount,
			NoCount:   noCount,
			UserVote:  userVote,
		}

		// Get group member IDs
		rows, err := db.Query(`
			SELECT user_id 
			FROM group_members 
			WHERE group_id = ? AND status = 'accepted'
		`, req.GroupID)
		if err == nil {
			defer rows.Close()
			var memberIDs []int
			for rows.Next() {
				var memberID int
				if rows.Scan(&memberID) == nil {
					memberIDs = append(memberIDs, memberID)
				}
			}

			chatHub.Mutex.RLock()
			for _, memberID := range memberIDs {
				// Do not send the event notification back to the creator
				if memberID == userID {
					continue
				}
				if client, ok := chatHub.Clients[memberID]; ok {
					select {
					case client.Send <- eventNotification:
					default:
						// Skip if buffer full
					}
				}
			}
			chatHub.Mutex.RUnlock()
		}

		// Success
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  true,
			"event_id": id,
		})
	}))

	// List events for a group with poll counts and this user's vote
	http.HandleFunc("/groups/", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/groups/")
		parts := strings.Split(path, "/")

		// ---------- POSTS ----------
		// GET /groups/{gid}/posts
		if len(parts) == 2 && parts[1] == "posts" && r.Method == http.MethodGet {
			gid := parts[0] // keep as string if your g.* wrappers expect string
			g.GetGroupPosts(db, w, r, gid)
			return
		}

		// GET /groups/{gid}/posts/{pid}/comments
		if len(parts) == 4 && parts[1] == "posts" && parts[3] == "comments" && r.Method == http.MethodGet {
			gid := parts[0]
			pid := parts[2]
			g.HandleGetGroupPostComments(db, w, r, gid, pid)
			return
		}

		// DELETE /groups/{gid}/posts/{pid}
		if len(parts) == 3 && parts[1] == "posts" && r.Method == http.MethodDelete {
			gid := parts[0]
			pid := parts[2]
			g.DeleteGroupPost(db, w, r, gid, pid)
			return
		}

		// ---------- EVENTS ----------
		// GET /groups/{gid}/events
		if len(parts) == 2 && parts[1] == "events" && r.Method == http.MethodGet {
			userID, loggedIn := u.ValidateSession(db, r)
			if !loggedIn {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			groupID, err := strconv.Atoi(parts[0])
			if err != nil || groupID <= 0 {
				http.Error(w, "Invalid group id", http.StatusBadRequest)
				return
			}

			// Must be accepted member to view
			var cnt int
			if err := db.QueryRow(`
            SELECT COUNT(*) FROM group_members
            WHERE group_id = ? AND user_id = ? AND status = 'accepted'
        `, groupID, userID).Scan(&cnt); err != nil || cnt == 0 {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			rows, err := db.Query(`
			SELECT id, title, description, datetime(created_at), event_date
			FROM events
			WHERE group_id = ?
			ORDER BY created_at DESC
		`, groupID)
			if err != nil {
				http.Error(w, "DB error", http.StatusInternalServerError)
				return
			}
			defer rows.Close()

			type EventOut struct {
				ID          int64   `json:"id"`
				Title       string  `json:"title"`
				Description string  `json:"description"`
				CreatedAt   string  `json:"created_at"`
				EventDate   *string `json:"event_date"`
				YesCount    int     `json:"yes_count"`
				NoCount     int     `json:"no_count"`
				UserVote    *string `json:"user_vote"` // "yes"|"no"|nil
			}

			var out []EventOut
			for rows.Next() {
				var ev EventOut
				if err := rows.Scan(&ev.ID, &ev.Title, &ev.Description, &ev.CreatedAt, &ev.EventDate); err == nil {
					_ = db.QueryRow(`SELECT COUNT(*) FROM event_votes WHERE event_id = ? AND choice='yes'`, ev.ID).Scan(&ev.YesCount)
					_ = db.QueryRow(`SELECT COUNT(*) FROM event_votes WHERE event_id = ? AND choice='no'`, ev.ID).Scan(&ev.NoCount)
					var uv sql.NullString
					_ = db.QueryRow(`SELECT choice FROM event_votes WHERE event_id = ? AND user_id = ?`, ev.ID, userID).Scan(&uv)
					if uv.Valid {
						v := uv.String
						ev.UserVote = &v
					}
					out = append(out, ev)
				}
			}

			w.Header().Set("Content-Type", "application/json")
			if out == nil {
				out = []EventOut{}
			}
			json.NewEncoder(w).Encode(out)
			return
		}

		// POST /groups/events/vote?event_id=123&choice=yes|no
		if len(parts) == 2 && parts[0] == "events" && parts[1] == "vote" && r.Method == http.MethodPost {
			userID, loggedIn := u.ValidateSession(db, r)
			if !loggedIn {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			eventID, _ := strconv.Atoi(r.URL.Query().Get("event_id"))
			choice := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("choice")))
			if eventID <= 0 || (choice != "yes" && choice != "no") {
				http.Error(w, "Invalid params", http.StatusBadRequest)
				return
			}

			var groupID int
			if err := db.QueryRow(`SELECT group_id FROM events WHERE id=?`, eventID).Scan(&groupID); err != nil {
				http.Error(w, "Event not found", http.StatusNotFound)
				return
			}
			var cnt int
			if err := db.QueryRow(`
            SELECT COUNT(*) FROM group_members
            WHERE group_id = ? AND user_id = ? AND status = 'accepted'
        `, groupID, userID).Scan(&cnt); err != nil || cnt == 0 {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			_, err := db.Exec(`
            INSERT INTO event_votes (event_id, user_id, choice)
            VALUES (?, ?, ?)
            ON CONFLICT(event_id, user_id) DO UPDATE SET choice=excluded.choice
        `, eventID, userID, choice)
			if err != nil {
				http.Error(w, "DB error", http.StatusInternalServerError)
				return
			}

			// Get username for WebSocket broadcast
			var username string
			db.QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&username)

			// Query updated counts and user's vote
			var yesCount, noCount int
			_ = db.QueryRow(`SELECT COUNT(*) FROM event_votes WHERE event_id = ? AND choice='yes'`, eventID).Scan(&yesCount)
			_ = db.QueryRow(`SELECT COUNT(*) FROM event_votes WHERE event_id = ? AND choice='no'`, eventID).Scan(&noCount)

			voteNotification := chat.Frontend{
				Type:      "group_event_vote_update",
				From:      userID,
				Username:  username,
				GroupID:   groupID,
				PostId:    eventID,
				YesCount:  yesCount,
				NoCount:   noCount,
				UserVote:  choice,
				Timestamp: time.Now(),
			}

			// Get group member IDs
			memberRows, err := db.Query(`
				SELECT user_id 
				FROM group_members 
				WHERE group_id = ? AND status = 'accepted'
			`, groupID)
			if err == nil {
				defer memberRows.Close()
				var memberIDs []int
				for memberRows.Next() {
					var memberID int
					if memberRows.Scan(&memberID) == nil {
						memberIDs = append(memberIDs, memberID)
					}
				}

				chatHub.Mutex.RLock()
				for _, memberID := range memberIDs {
					if client, ok := chatHub.Clients[memberID]; ok {
						select {
						case client.Send <- voteNotification:
						default:
							// Skip if buffer full
						}
					}
				}
				chatHub.Mutex.RUnlock()
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"success": true})
			return
		}

		http.NotFound(w, r)
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
		cookie := &http.Cookie{
			Name:     "session_token",
			Value:    "",
			Expires:  time.Now().Add(-1 * time.Hour), // Expire immediately
			HttpOnly: true,
			Path:     "/",
			Domain:   "localhost",
		}
		http.SetCookie(w, cookie)

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

	// WebSocket endpoint (chatHub already initialized at the top of ConnectWeb)
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
		userID, loggedIn := u.ValidateSession(db, r)
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

			// Get current user's username for follow checking
			currentUsername, err := database.GetUsernameUsingID(db, userID)
			if err != nil {
				http.Error(w, "Failed to get current user", http.StatusInternalServerError)
				e.ErrorHandler(w, r, 500)
				return
			}

			rows, err := db.Query(`SELECT id, username, isPrivate FROM users`)
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
				var isPrivate bool
				if err := rows.Scan(&id, &username, &isPrivate); err == nil {
					// Check if current user is following this user
					isFollowing := u.IsFollowing(db, currentUsername, username)

					users = append(users, map[string]interface{}{
						"id":          id,
						"username":    username,
						"online":      onlineSet[id], // ✅ Add online status
						"isPrivate":   isPrivate,
						"isFollowing": isFollowing,
					})
				}
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(users)
		}
	}))

	http.HandleFunc("/check-following/", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		_, loggedIn := u.ValidateSession(db, r)
		if !loggedIn {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		path := strings.TrimPrefix(r.URL.Path, "/check-following/")
		parts := strings.Split(path, "/")
		if len(parts) != 2 {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}

		followerUsername := parts[0]
		followingUsername := parts[1]

		isFollowing := u.IsFollowing(db, followerUsername, followingUsername)

		response := map[string]interface{}{
			"isFollowing": isFollowing,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))

	http.HandleFunc("/group-messages", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		userID, loggedIn := u.ValidateSession(db, r)
		if !loggedIn {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		groupIDStr := r.URL.Query().Get("group_id")
		offsetStr := r.URL.Query().Get("offset")
		if groupIDStr == "" {
			http.Error(w, "Missing group_id", http.StatusBadRequest)
			return
		}
		groupID, err := strconv.Atoi(groupIDStr)
		if err != nil || groupID <= 0 {
			http.Error(w, "Invalid group_id", http.StatusBadRequest)
			return
		}
		offset, _ := strconv.Atoi(offsetStr)

		// Optional: ensure the requester is a member of the group
		var count int
		err = db.QueryRow(`
		SELECT COUNT(*) FROM group_members
		WHERE group_id = ? AND user_id = ? AND status = 'accepted'
	`, groupID, userID).Scan(&count)
		if err != nil || count == 0 {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		const q = `
		SELECT gm.sender_id, u.username, gm.content, gm.created_at
		FROM group_messages gm
		JOIN users u ON gm.sender_id = u.id
		WHERE gm.group_id = ?
		ORDER BY gm.created_at DESC
		LIMIT 30 OFFSET ?
	`
		rows, err := db.Query(q, groupID, offset)
		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var messages []chat.Frontend
		for rows.Next() {
			var m chat.Frontend
			m.Type = "group_message"
			m.GroupID = groupID
			if err := rows.Scan(&m.From, &m.Username, &m.Content, &m.Timestamp); err == nil {
				messages = append(messages, m)
			}
		}
		if messages == nil {
			messages = []chat.Frontend{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(messages)
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

func ensureEventPollTables(db *sql.DB, hub *chat.Hub) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			group_id   INTEGER NOT NULL,
			creator_id INTEGER NOT NULL,
			title       TEXT NOT NULL,
			description TEXT NOT NULL,
			event_date  DATETIME,
			created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS event_votes (
            event_id INTEGER NOT NULL,
            user_id  INTEGER NOT NULL,
            choice   TEXT NOT NULL CHECK (choice IN ('yes','no')),
            UNIQUE(event_id, user_id)
        );`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return err
		}
	}

	// Try to add event_date column if the table already existed without it. Ignore error if it already exists.
	_, _ = db.Exec(`ALTER TABLE events ADD COLUMN event_date DATETIME`)

	// Add comprehensive profile and follow endpoints
	http.HandleFunc("/profile/complete", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		u.GetCompleteProfileHandler(db, w, r)
	}))

	http.HandleFunc("/follow-user", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		u.FollowUser(db, hub, w, r)
	}))

	http.HandleFunc("/unfollow-user", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		u.UnfollowUser(db, hub, w, r)
	}))

	http.HandleFunc("/handle-follow-request", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		u.HandleFollowRequest(db, hub, w, r)
	}))

	http.HandleFunc("/all-users", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		u.GetAllUsers(db, w, r)
	}))

	http.HandleFunc("/follow-status", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		u.GetFollowStatus(db, w, r)
	}))

	http.HandleFunc("/toggle-privacy", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		u.TogglePrivacy(db, hub, w, r)
	}))

	http.HandleFunc("/update-profile", cor.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		// Get current user from session
		userID, loggedIn := u.ValidateSession(db, r)
		if !loggedIn {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Get username for the existing UpdateProfileHandler
		var username string
		err := db.QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&username)
		if err != nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		u.UpdateProfileHandler(db, w, r, username)
	}))

	// JSON version of update profile (simpler and more reliable)

	return nil
}
