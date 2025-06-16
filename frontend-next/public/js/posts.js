// posts.js
// const createPostButton = document.getElementById('createPostButton');
// const categoryButtons = document.querySelectorAll('#categoryOptions .button-side');
const profilePageButton = document.getElementById('profilePageButton');
const postsButton = document.getElementById('postsButton');


//load all the posts
function loadPosts() {
    fetch('http://localhost:8080/get-posts', {
        method: 'GET',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include'
    })
        .then(response => {
            if (response.status === 404) {
                console.log(error); // Handle posts not found
                return;
            }
            if (!response.ok) {
                throw new Error('Failed to fetch posts');
            }
            return response.json();
        })
        .then(posts => {
            const postContainer = document.querySelector('.container-post');
            postContainer.innerHTML = '';
            postContainer.innerHTML = '<h1>Posts</h1>';


            if (posts == null) {
                postContainer.innerHTML += "<p>No posts available.</p>";
                return;
            }

            posts.forEach(post => {
                const postElement = document.createElement('div');
                postElement.classList.add('post-post');
                postElement.innerHTML = `
            <div class="comment-post">
                <h2>${post.title}</h2>
                <p>${post.content}</p>
                <small>
                Posted by 
                <button onclick="window.location.href='/theirProfile?user=${encodeURIComponent(post.username)}'">${post.username}</button>
                on ${post.createdAt}
                </small>
                <br>
                <small>Category: ${post.categories.join(", ")}</small>
                <br><br>
                <button class="commentsButton button-main" data-post-id="${post.id}">See comments</button>
                <br><br>
                <span class="material-icons" onclick="likeDislikePost(${post.id}, true); "> thumb_up </span>
                <span class="material-icons" onclick="likeDislikePost(${post.id}, false); "> thumb_down </span>
            <small>
                <span id="likesCountPost${post.id}">Likes: 0</span>
                <span id="dislikesCountPost${post.id}">Dislikes: 0</span>
            </small>                    </div>
        `;
                postContainer.appendChild(postElement);
                //  Fetch updated likes/dislikes for this post
                getInteractions(post.id);
            });

            document.querySelectorAll('.commentsButton').forEach(button => {
                button.addEventListener('click', () => window.location.href = '/comments?post_id=' + button.dataset.postId);
            });

        })
        .catch(error => console.log(error));
}

//load only user posts
function loadMyPosts(myUsername) {

    const myProfileUsername = document.getElementById('myProfileUsername');
    myProfileUsername.innerHTML = myUsername;

    fetch('http://localhost:8080/get-myPosts', {
        method: 'GET',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include'
    })
        .then(response => {
            if (response.status === 404) {
                console.log(error); // Handle posts not found
                return;
            }
            if (!response.ok) {
                throw new Error('Failed to fetch posts');
            }
            return response.json();
        })
        .then(posts => {


            const postContainer = document.querySelector('.container-profilePost');
            postContainer.innerHTML = '';
            postContainer.innerHTML += '<h1>Posts</h1>';


            if (posts == null) {
                postContainer.innerHTML += "<p>No posts available.</p>";
                return;
            }

            posts.forEach(post => {
                localStorage.setItem('username', post.username); // Store username in localStorage
                const postElement = document.createElement('div');
                postElement.classList.add('post-post');
                postElement.innerHTML = `
                    <div class="comment-post">
                        <h2>${post.title}</h2>
                        <p>${post.content}</p>
                        <small>
                        Posted by 
                        <button onclick="window.location.href='/theirProfile?user=${encodeURIComponent(post.username)}'">${post.username}</button>
                        on ${post.createdAt}
                        </small>
                        <br>
                        <small>Category: ${post.categories.join(", ")}</small>
                        <br><br>
                        <button class="commentsButton button-main" data-post-id="${post.id}">See comments</button>
                        <br><br>
                        <span class="material-icons" onclick="likeDislikePost(${post.id}, true);"> thumb_up </span>
                        <span class="material-icons" onclick="likeDislikePost(${post.id}, false);"> thumb_down </span>
                        <small>
                            <span id="likesCountPost${post.id}">Likes: 0</span>
                            <span id="dislikesCountPost${post.id}">Dislikes: 0</span>
                        </small>                    
                    </div>
                `;
                postContainer.appendChild(postElement);
                //  Fetch updated likes/dislikes for this post
                getInteractions(post.id);
            });

            document.querySelectorAll('.commentsButton').forEach(button => {
                button.addEventListener('click', () => window.location.href = '/comments?post_id=' + button.dataset.postId);
            });

        })
        .catch(error => console.log(error));

    
    fetch("http://localhost:8080/editGet/" + myUsername, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json'
        },
        credentials: 'include'
    })
        .then(response => {
            if (!response.ok) {
                throw new Error('Failed to fetch profile data');
            }
            return response.json();
        })
        .then(profile => { 
            document.getElementById('profileBio').innerHTML = profile.bio || 'No bio available.';
            console.log("Profile Bio: ", profile.bio);
        })
        .catch(error => {
            console.error('Error fetching user bio:', error);
        });
    
}

//load posts with specific category
function loadCategoryPosts(category) {
    fetch('http://localhost:8080/category/' + category, {
        method: 'GET',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include'
    })
        .then(response => {
            if (response.status === 404) {
                console.log(error); // Handle category not found
                return;
            }
            if (!response.ok) {
                throw new Error('Network response was not ok');
            }
            return response.json();
        })
        .then(posts => {
            const postContainer = document.querySelector('.container-post');

            if (!postContainer) {
                console.error("Post container not found!");
                return;
            }

            postContainer.innerHTML = ''; // Clear previous posts
            postContainer.innerHTML += '<h1>Posts</h1>';

            if (!Array.isArray(posts)) {
                console.error("Expected posts to be an array, but got:", posts);
                postContainer.innerHTML += "<p>Error loading posts.</p>";
                return;
            }

            if (posts.length === 0) {
                postContainer.innerHTML += "<p>No posts available.</p>";
                return;
            }

            posts.forEach(post => {
                const postElement = document.createElement('div');
                postElement.classList.add('post-post');
                postElement.innerHTML = `
            <div class="comment-post">
                <h2>${post.title}</h2>
                <p>${post.content}</p>
                <small>
                Posted by 
                <button onclick="window.location.href='/theirProfile?user=${encodeURIComponent(post.username)}'">${post.username}</button>
                on ${post.createdAt}
                </small>
                <br>
                <small>Category: ${post.categories.join(", ")}</small>
                <br><br>
                <button class="commentsButton button-main" data-post-id="${post.id}">See comments</button>
                <br><br>
                <span class="material-icons" onclick="likeDislikePost(${post.id}, true);"> thumb_up </span>
                <span class="material-icons" onclick="likeDislikePost(${post.id}, false);"> thumb_down </span><small>
                <span id="likesCountPost${post.id}">Likes: 0</span>
                <span id="dislikesCountPost${post.id}">Dislikes: 0</span></small>
                </div>
                `;
                postContainer.appendChild(postElement);
                //  Fetch updated likes/dislikes for this post
                getInteractions(post.id);
            });

            document.querySelectorAll('.commentsButton').forEach(button => {
                button.addEventListener('click', () => window.location.href = '/comments?post_id=' + button.dataset.postId);
            });
        })
        .catch(error => console.log(error));
}

//load another user's posts
function loadTheirProfile(username,myUsername) {
    console.log("Loading profile for username:", username);
    console.log("Current username:", myUsername);
    // Check if the username matches the current user's username
    if (username == myUsername) {
        console.log("Redirecting to myProfile for current user: ", myUsername);
        window.location.href = '/myProfile';
        return;
    }
    const profileUsername = document.getElementById('profileUsername');
    profileUsername.innerHTML = username;
    fetch('http://localhost:8080/get-otherPosts/' + username, {
        method: 'GET',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include'
    })
        .then(response => {
            if (response.status === 404) {
                console.log(error); // Handle posts not found
                return;
            }
            if (!response.ok) {
                throw new Error('Failed to fetch posts');
            }
            return response.json();
        })
        .then(posts => {
            const postContainer = document.querySelector('.container-theirProfilePost');
            postContainer.innerHTML = '';
            postContainer.innerHTML += '<h1>Posts</h1>';


            if (posts == null) {
                postContainer.innerHTML += "<p>No posts available.</p>";
                return;
            }

            posts.forEach(post => {
                const postElement = document.createElement('div');
                postElement.classList.add('post-post');
                postElement.innerHTML = `
                    <div class="comment-post">
                        <h2>${post.title}</h2>
                        <p>${post.content}</p>
                        <small>Posted by <button onclick="window.location.href='/theirProfile?user=' + encodeURIComponent('${post.username}')">${post.username}</button> 
                        on ${post.createdAt}</small><br>
                        <small>Category: ${post.categories.join(", ")}</small>
                        <br><br>
                        <button class="commentsButton button-main" data-post-id="${post.id}">See comments</button>
                        <br><br>
                        <span class="material-icons" onclick="likeDislikePost(${post.id}, true);"> thumb_up </span>
                        <span class="material-icons" onclick="likeDislikePost(${post.id}, false);"> thumb_down </span>
                        <small>
                            <span id="likesCountPost${post.id}">Likes: 0</span>
                            <span id="dislikesCountPost${post.id}">Dislikes: 0</span>
                        </small>                    
                    </div>
                `;
                postContainer.appendChild(postElement);
                //  Fetch updated likes/dislikes for this post
                getInteractions(post.id);
            });

            document.querySelectorAll('.commentsButton').forEach(button => {
                button.addEventListener('click', () => window.location.href = '/comments?post_id=' + button.dataset.postId);
            });

        })
        .catch(error => console.log(error));
}

// function toggleDropdown(id) {
//     var dropdown = document.getElementById(id);
//     dropdown.style.display = dropdown.style.display === 'block' ? 'none' : 'block'; // Toggle visibility
// }

document.addEventListener('DOMContentLoaded', () => {
    checkSession().then(() => {
        const myUsername = mycurrentUsername();
        console.log("Current username:", myUsername);
        const path = window.location.pathname;

        if (path === '/myProfile') {
            loadMyPosts(myUsername);
            return;
        }

        if (path === '/posts') {
            loadPosts();
            return;
        }

        if (path === '/theirProfile') {
            const params = new URLSearchParams(window.location.search);
            const theirUsername = params.get('user');
            console.log("Username from URL:", theirUsername);

            if (!theirUsername) {
                console.error("No username provided in URL.");
                return;
            }

            loadTheirProfile(theirUsername,myUsername);
            return;
        }

        // Optional: handle unknown routes
        console.warn("No matching handler for:", path);
    }).catch((err) => {
        console.error("Session check failed:", err);
        window.location.href = '/login';
    });
});


// categoryButtons.forEach(button => {
//     button.addEventListener('click', () => {
//         const category = button.value; // Get the category from the button's value
//         history.pushState(null, '', `?${encodeURIComponent(category)}`);
//         loadCategoryPosts(category);
//     });
// });

// if (createPostButton) createPostButton.addEventListener('click', () => window.location.href = "/createPost");

if (profilePageButton) profilePageButton.addEventListener('click', () => {
    window.location.href = '/myProfile';
});

if (postsButton) postsButton.addEventListener('click', () => {
    loadPosts();
});