const returnToPost = document.getElementById("return-to-post");
if (returnToPost) returnToPost.addEventListener('click', () => window.location.href = '/posts');
// Create form submission
const createPostForm = document.getElementById('createPostForm');

//Create form
if (createPostForm) {
    createPostForm.addEventListener('submit', function (event) {

        event.preventDefault(); // Prevent default form submission

        // Get selected categories (checkboxes)
        const selectedCategories = [];
        document.querySelectorAll('input[name="category"]:checked').forEach((checkbox) => {
            selectedCategories.push(checkbox.value);
        });

        const postData = {
            title: document.getElementById('title').value,
            content: document.getElementById('content').value,
            categories: selectedCategories // Send array of selected categories
        };

        fetch('http://localhost:8080/create-post', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(postData),
            credentials: 'include' // Ensures session cookies are sent
        })
            .then(response => response.json())
            .then(data => {
                console.log(data.success ? "Post created successfully!" : "Error: " + data.message);
                if (data.success) createPostForm.reset();
                const socket = window.getSocket?.();
                if (socket && socket.readyState === WebSocket.OPEN) {
                    socket.send(JSON.stringify({ type: "new_post" }));
                    // socket.send(JSON.stringify({ from: userID, type: "notif" }));
                }
                window.location.href = '/posts';
            })
            .catch(error => console.log(error));
    });
}
