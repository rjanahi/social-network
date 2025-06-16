// script.js
// Get references to buttons
const signUpButton = document.getElementById('signUpButton');
const logInButton = document.getElementById('logInButton');
const postsButton = document.getElementById('postsButton');
const aboutUsButton = document.getElementById('aboutUsButton');
const logoutPostButton = document.getElementById('logoutPostButton');



let userID = 0;
let Chatusername;


document.addEventListener('DOMContentLoaded', () => {
    checkSession();
    // Event listeners for navigation buttons
    if (signUpButton) signUpButton.addEventListener('click', () => window.location.href='/signup.html');
    if (logInButton) logInButton.addEventListener('click', () => window.location.href='/login.html');
    if (postsButton) {
        postsButton.addEventListener('click', () => window.location.href='/posts.html');
    }
    if (aboutUsButton) aboutUsButton.addEventListener('click', () => window.location.href='/aboutUs.html');

    // Logout functionality
    if (logoutButton) {
        logoutButton.addEventListener('click', function () {

            fetch('http://localhost:8080/logout', {
                method: 'POST',
                credentials: 'include'
            })
                .then(response => response.json())
                .then(data => {
                    console.log(data.message);
                    checkSession(); // Refresh session check to update UI
                    window.location.href='/mainPage.html'; // Redirect to main page after logout
                })
                .catch(error => console.log(error));
        });
    }

});


