function connectWebSocket(userId) {
  if (userId == null || userId == "undefined" || userId == 0) {
    console.log("WebSocket Not connected");
    return;
  }

  socket = new WebSocket(`ws://localhost:8080/ws?user_id=${userId}`);
  window.socket = socket; 

  socket.onerror = () => {
    console.log("WebSocket error occurred");
  };

  socket.onopen = () => {
    console.log("WebSocket connected");
  };

  socket.onmessage = (event) => {
    const msg = JSON.parse(event.data);
    console.log(msg.type);
    if (msg.type === "typing") {
      if (msg.from === selectedUserId) {
        showTypingIndicator(Theirname);
      }
      return;
    }

    if (msg.type === "message") {
      appendMessageToChat(msg);
      return;
    }

    if (msg.type === "status_update") {
      updateUserStatus(msg.username, msg.status);
    }

    if (msg.type === "new_user") {
      fetchUserList();
      return;
    }
    if (msg.type === "new_post") {
      if (window.location.pathname === "/posts") {
        loadPosts();
      }
      return;
    }
    if (msg.type === "notif") {
      alert("New Notification");
      return;
    }

    if (msg.type === "new_comment") {
      console.log("Post ID:", msg.post_id);
      if ( window.location.search.includes(`${msg.post_id}`)) {
        loadCommentsForPost(msg.post_id);
        return;
      }
    };

    if (msg.type === "new_postLike") {

      getInteractions(msg.post_id);
      if ( window.location.search.includes(`Liked`)) {
        loadCategoryPosts(`Liked`);
        return;
      }

      return;
    }

    if (msg.type === "new_commentLike") {

      getInteractions(null, msg.comment_id);

      return;
    }

    socket.onclose = () => {
      console.log("WebSocket disconnected");
    };
  }
}

function disconnectWeb() {
  socket.close();
  console.log("Socket closed.");
}

function getSocket() {
  return socket;
}

window.getSocket = getSocket;
window.connectWebSocket = connectWebSocket;
window.disconnectWeb = disconnectWeb;
