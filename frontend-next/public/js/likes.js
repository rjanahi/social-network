function likeDislikeComment(commentId, isLike) {

    console.log(` Sending Like/Dislike request for Comment ID: ${commentId}, Is Like: ${isLike}`);

    let likesElement = document.getElementById(`likesCountComment${commentId}`);
    let dislikesElement = document.getElementById(`dislikesCountComment${commentId}`);

    if (!likesElement || !dislikesElement) {
        console.error(` Elements for comment ${commentId} not found.`);
        return;
    }


    fetch('http://localhost:8080/likeDislikeComment', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ comment_id: commentId, is_like: isLike }),
        credentials: 'include'
    })
        .then(response => response.json())
        .then(data => {
            console.log(" Like/Dislike Comment Response:", data);

            if (data.message === 'Interaction updated successfully') {
                likesElement.innerText = `Likes: ${data.likes}`;
                dislikesElement.innerText = `Dislikes: ${data.dislikes}`;
            } else {
                console.log(data.error || "Something went wrong.");
            }
            const socket = window.getSocket?.();
            if (socket && socket.readyState === WebSocket.OPEN) {
                socket.send(JSON.stringify({
                    type: "new_commentLike",
                    comment_id: parseInt(commentId),
                    is_like: isLike
                }));
            } else {
                console.warn("WebSocket not open — skipping broadcast.");
            }

        })
        .catch(error => console.log(error));
}

function likeDislikePost(postId, isLike) {


    let likesElement = document.getElementById(`likesCountPost${postId}`);
    let dislikesElement = document.getElementById(`dislikesCountPost${postId}`);

    if (!likesElement || !dislikesElement) {
        console.error(` Elements for post ${postId} not found.`);
        return;
    }

    fetch('http://localhost:8080/likeDislikePost', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ post_id: postId, is_like: isLike }),
        credentials: 'include'
    })
        .then(response => response.json())
        .then(data => {
            console.log(" Like/Dislike Response:", data);

            if (data.message === 'Interaction updated successfully') {

                //  Update UI only after getting the correct values from backend
                likesElement.innerText = `Likes: ${data.likes}`;
                dislikesElement.innerText = `Dislikes: ${data.dislikes}`;
            } else {
                console.log(data.error || "Something went wrong.");
            }
            const socket = window.getSocket?.();
            if (socket && socket.readyState === WebSocket.OPEN) {
                socket.send(JSON.stringify({
                    type: "new_postLike",
                    comment_id: parseInt(postId),
                    is_like: isLike
                }));
            } else {
                console.warn("WebSocket not open — skipping broadcast.");
            }

        })
        .catch(error => {
            console.error(' Error:', error);
            console.log(error)
        });
}

function getInteractions(postId, commentId = null) {


    let requestBody = commentId ? { comment_id: commentId } : { post_id: postId };

    fetch('http://localhost:8080/getInteractions', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(requestBody),
        credentials: 'include'
    })
        .then(response => response.json())
        .then(data => {
            console.log(" Updated Interaction Data:", data);

            if (!data || typeof data.likes === "undefined" || typeof data.dislikes === "undefined") {
                console.error(" Invalid data received:", data);
                return;
            }

            if (commentId) {
                //  Update comment likes/dislikes
                let likesElement = document.getElementById(`likesCountComment${commentId}`);
                let dislikesElement = document.getElementById(`dislikesCountComment${commentId}`);

                if (likesElement) likesElement.innerText = `Likes: ${data.likes}`;
                if (dislikesElement) dislikesElement.innerText = `Dislikes: ${data.dislikes}`;
            } else {
                //  Update post likes/dislikes
                let likesElement = document.getElementById(`likesCountPost${postId}`);
                let dislikesElement = document.getElementById(`dislikesCountPost${postId}`);

                if (likesElement) likesElement.innerText = `Likes: ${data.likes}`;
                if (dislikesElement) dislikesElement.innerText = `Dislikes: ${data.dislikes}`;
            }
        })
        .catch(error => console.log(error));
}

window.getInteractions = getInteractions;
window.likeDislikeComment = likeDislikeComment;
window.likeDislikePost = likeDislikePost;