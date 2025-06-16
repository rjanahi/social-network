// chat.js
let loggedInUserId = null;
let selectedUserId = null;
let typingTimeout = null;
let currentOffset = 0;
let throttle = false;

let Myusername;
let Theirname;

const createGroupButton = document.getElementById("createGroupButton");
const groupContainer = document.getElementById("group-container");
const chatMain = document.getElementById("chat-main");
const chatWindow = document.getElementById("chatWindow");
const openChatButton = document.getElementById('openChatButton');

let currentHeight = 150; // Starting height
const maxHeight = 200; // Maximum height

const socket = window.getSocket?.();  

//todo: create a group with selected users
if (createGroupButton) {
  createGroupButton.addEventListener('click', () => {
    const groupContainer = document.getElementById("group-container");
    const groupMembers = document.getElementById("groupMembers");
    const createGroupForm = document.getElementById("createGroupForm");

    chatMain.style.display = "none";
    groupContainer.removeAttribute("hidden");

    fetchUserListGroup(); // Populate user checkboxes
    groupMembers.innerHTML = ""; // Clear previous members

    // Avoid attaching multiple submit handlers
    createGroupForm.onsubmit = (e) => {
      e.preventDefault();

      const groupNameInput = document.getElementById("groupName");
      const groupName = groupNameInput.value.trim();

      if (!groupName) return;

      const selectedUsers = [];
      document.querySelectorAll('input[name="user"]:checked').forEach((checkbox) => {
        selectedUsers.push(checkbox.value);
      });

      // TODO: Ask private users first
      // TODO: Create group with selected users
      // TODO: Send notifications

      console.log("Creating group:", groupName, "with users:", selectedUsers);
      // createGroup(groupName, selectedUsers); // uncomment when implemented

      // Reset UI
      groupNameInput.value = "";
      createGroupForm.reset();
      groupContainer.setAttribute("hidden", "true");
      chatMain.style.display = "block";
    };
  });
}


function fetchUserList() {

  fetch("http://localhost:8080/get-users", {
    method: "GET",
    credentials: "include",
  })
    .then((res) => {
      if (res.status === 401) {
        console.error("Unauthorized access. Please log in.");
        return;
      }
      if (res.status === 404) {
        console.log(error);
        return;
      }
      return res.json();
    })
    .then((users) => {
      const userList = document.getElementById("userList");
      userList.innerHTML = "";

      if (users == null) {

        return;
      } else {
        users.forEach((user) => {
          if (user.id !== loggedInUserId) {
            const li = document.createElement("li");
            li.dataset.userId = user.id;
            li.classList.add("user-item");

            // Create username span
            const usernameSpan = document.createElement("span");
            usernameSpan.textContent = user.username;

            // Create status dot
            const statusDot = document.createElement("span");
            statusDot.classList.add("status-dot");
            statusDot.classList.add(user.online ? "online" : "offline");

            // Append elements
            li.appendChild(usernameSpan);
            li.appendChild(statusDot);
            li.onclick = () => openChatWith(user.id, user.username);

            userList.appendChild(li);
          } else {
            Myusername = user.username;
          }
        });
      }
    })
    .catch((err) => console.log(err));
}

function fetchUserListGroup() {


  fetch("http://localhost:8080/get-users", {
    method: "GET",
    credentials: "include",
  })
    .then((res) => {
      if (res.status === 401) {
        console.error("Unauthorized access. Please log in.");
        return;
      }
      if (res.status === 404) {
        console.log(error);
        return;
      }
      return res.json();
    })
    .then((users) => {
      const userList = document.getElementById("groupMembers");
      userList.innerHTML = "";

      if (!users) return;

      users.forEach((user) => {
        //only show public users or/and private users that follow back
        if (user.id !== loggedInUserId && user.isPrivate !== false||user.isPrivate === true && user.isFollowing) {
          const label = document.createElement("label");
          label.style.display = "block";

          const checkbox = document.createElement("input");
          checkbox.type = "checkbox";
          checkbox.name = "user";
          checkbox.value = user.id;

          const textNode = document.createTextNode(` ${user.username}`);

          label.appendChild(checkbox);
          label.appendChild(textNode);

          userList.appendChild(label);
        } else {
          Myusername = user.username;
        }
      });
    })
    .catch((err) => console.log(err));
}


function setupChatForm() {
  const chatForm = document.getElementById("chatForm");
  const chatInput = document.getElementById("chatInput");

  chatForm.addEventListener("submit", (e) => {
    e.preventDefault();
    const content = chatInput.value.trim();
    if (content && selectedUserId) {
      sendMessage(selectedUserId, content);
      chatInput.value = ""; // Clear input after sending
    }
  });

  chatInput.addEventListener("input", () => {
    sendTypingSignal();
  });
}

function loadMessages(withId, offset = 0) {

  fetch(`http://localhost:8080/messages?with=${withId}&offset=${offset}`, { credentials: "include" })
    .then((res) => res.json())
    .then((messages) => {
      if (!messages || !Array.isArray(messages)) {
        console.warn("No messages received or invalid format.");
        return;
      }

      messages.forEach((msg) => prependMessageToChat(msg));
      throttle = false;
    })
    .catch((err) => console.log(err));
}

function prependMessageToChat(msg) {
  const container = document.getElementById("chatWindow");
  const node = document.createElement("div");
  node.classList.add("chat-message");


  if (msg.from === loggedInUserId) {
    node.classList.add("my-message");
    node.innerHTML = `<strong>${Myusername}</strong> <strong>${new Date(
      msg.timestamp
    ).toLocaleString()}:</strong><br> ${msg.content}`;
  } else {
    node.classList.add("received-message");
    node.innerHTML = `<strong>${Theirname}</strong> <strong>${new Date(
      msg.timestamp
    ).toLocaleString()}:</strong><br> ${msg.content}`;
  }

  // Insert the new message at the top
  container.insertBefore(node, container.firstChild);
}

function updateUserStatus(username, status) {
  const userList = document
    .getElementById("userList")
    .getElementsByTagName("li");
  for (let li of userList) {
    const userItem = li.querySelector(".username");
    if (userItem && userItem.textContent === username) {
      const statusDot = li.querySelector(".status-dot");
      if (status === "online") {
        statusDot.classList.add("online");
        statusDot.classList.remove("offline");
      } else {
        statusDot.classList.add("offline");
        statusDot.classList.remove("online");
      }
      break;
    }
  }
}

function openChatWith(userId, username) {
  groupContainer.setAttribute("hidden", "true");
  chatMain.style.display = "block";
  Theirname = username;
  selectedUserId = userId;
  currentOffset = 0;
  chatWindow.style.display = "flex";
  document.getElementById("chatWindow").innerHTML = "";
  document.getElementById(
    "chatWithLabel"
  ).textContent = `Chat with ${username}`;
  window.location.href="/chat.html";

  const chatForm = document.getElementById("chatForm");
  chatForm.style.display = "flex";

  loadMessages(userId);
  setupScroll(userId);
}

function sendMessage(toId, content) {
  if (!socket || socket.readyState !== WebSocket.OPEN) return;

  const message = {
    type: "message",
    from: loggedInUserId,
    to: toId,
    content: content,
    timestamp: new Date().toISOString(),
  };

  socket.send(JSON.stringify(message));
  appendMessageToChat(message);
}

function sendTypingSignal() {
  if (!socket || !selectedUserId) return;

  const signal = {
    type: "typing",
    from: loggedInUserId,
    username: Myusername,
    to: selectedUserId,
  };

  socket.send(JSON.stringify(signal));
}

function showTypingIndicator(username) {
  const container = document.getElementById("chatWindow");

  // Check if the typing indicator already exists
  const existingTypingMsg = container.querySelector(".typing-message");
  if (!existingTypingMsg) {
    const typingNode = document.createElement("div");
    typingNode.classList.add("typing-message");
    typingNode.textContent = `${username} is typing...`;
    container.appendChild(typingNode);
    container.scrollTop = container.scrollHeight;
  }

  // Reset the timeout for hiding the indicator
  if (typingTimeout) clearTimeout(typingTimeout);
  typingTimeout = setTimeout(() => {
    if (existingTypingMsg) {
      existingTypingMsg.remove();
    }
  }, 1000);
}

function appendMessageToChat(msg) {
  const messagesContainer = document.getElementById("chatWindow");
  const newMessage = document.createElement("div");
  newMessage.classList.add("chat-message");

  if (msg.from === loggedInUserId) {
    newMessage.classList.add("my-message");
    newMessage.innerHTML = `<strong>${Myusername}</strong> <strong>${new Date(
      msg.timestamp
    ).toLocaleString()}:</strong><br> ${msg.content}`;
  } else {
    newMessage.classList.add("received-message");
    newMessage.innerHTML = `<strong>${Theirname}</strong> <strong>${new Date(
      msg.timestamp
    ).toLocaleString()}:</strong><br> ${msg.content}`;
  }

  // Append the new message instead of prepending
  messagesContainer.append(newMessage);
  messagesContainer.scrollTop = messagesContainer.scrollHeight;
}

function loadAndInitChat(userId) {

  loggedInUserId = userId;
  connectWebSocket(userId);
  fetchUserList();
  setupChatForm();
  updateUserListPeriodically();
  chatWindow.scrollTop = chatWindow.scrollHeight;
}

function setupScroll(chatUserId) {
  const container = document.getElementsByClassName("chat-window")[0];
  container.addEventListener("scroll", () => {
    if (container.scrollTop === 0 && !throttle) {
      throttle = true;
      if (currentHeight < maxHeight) {
        currentHeight += 10;
        container.style.height = `${currentHeight}px`;
      }
      currentOffset += 10;
      loadMessages(chatUserId, currentOffset);
      setTimeout(() => {
        throttle = false;
      }, 1500);
    }
  });
}

function hideAllSections() {
  document.getElementById("mainPage").hidden = true;
  document.getElementById("signUpSection").hidden = true;
  document.getElementById("logInSection").hidden = true;
  document.getElementById("postPageSection").hidden = true;
  document.getElementById("createPostSection").hidden = true;
  document.getElementById("aboutUsSection").hidden = true;
  document.getElementById("commentsSection").hidden = true;
  document.getElementById("chatSection").hidden = true;
}

function returnToPosts() {
  hideAllSections();
  document.getElementById("postPageSection").hidden = false;
  history.pushState(null, "", "/posts");
}

function updateUserListPeriodically() {


  setInterval(() => {
    fetchUserList();
  }, 300);
}


    if (openChatButton) {
        openChatButton.addEventListener('click', () => {
        loadAndInitChat(userID);
            window.location.href='/chat.html';
            loadAndInitChat(userID);
        });
    }
// Expose functions globally
window.returnToPosts = returnToPosts;
window.loadAndInitChat = loadAndInitChat;

