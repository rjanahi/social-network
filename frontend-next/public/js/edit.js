const editProfileForm = document.getElementById('editProfileForm');
const editProfileMessage = document.getElementById('editProfileMessage');
// const username = localStorage.getItem('username') || ''; // Replace with actual username retrieval logic

function submitProfile(username) {
    const formData = new FormData(editProfileForm);
    const data = {
    username: formData.get('username'),
    bio: formData.get('bio'),
    fname: formData.get('fname'),
    lname: formData.get('lname'),
    email: formData.get('email'),
    age: parseInt(formData.get('age')) || 0,
    gender: formData.get('gender'),
};

    fetch("http://localhost:8080/editPost/" + username, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        credentials: 'include',
        body: JSON.stringify(data)
    })
        .then(response => response.json())
        .then(res => {
            if (!res.success) {
                throw new Error('Failed to update profile');
            }
            editProfileMessage.textContent = 'Profile updated successfully!';
            editProfileMessage.className = 'message success';
            window.location.href = '/myProfile';
        })
        .catch(error => {
            editProfileMessage.textContent = 'Error updating profile: ' + error.message;
            editProfileMessage.className = 'message error';
        });
}


function fetchUserProfile(username) {
    fetch("http://localhost:8080/editGet/" + username, {
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
            document.getElementById('username').value = profile.username || '';
            document.getElementById('bio').value = profile.bio || '';
            document.getElementById('fname').value = profile.fname || '';
            document.getElementById('lname').value = profile.lname || ''; 
            document.getElementById('email').value = profile.email || '';
            document.getElementById('gender').value = profile.gender || '';
            document.getElementById('email').value = profile.email || '';
            document.getElementById('age').value = profile.age || '';
            console.log("Profile data fetched successfully:", profile);
        })
        .catch(error => {
            editProfileMessage.textContent = 'Error fetching profile data: ' + error.message;
            editProfileMessage.className = 'message error';
        });
}


document.addEventListener('DOMContentLoaded', () => {
    checkSession().then(() => {
        const username = mycurrentUsername();
        console.log("Current username after session check:", username);
        fetchUserProfile(username);

        editProfileForm.addEventListener('submit', (event) => {
            event.preventDefault();
            submitProfile(username);
        });
    });
});
