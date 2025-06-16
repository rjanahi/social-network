const loginForm = document.getElementById('loginForm');
const loginSignUpButton = document.getElementById('signUpButtonLogin');


if (loginSignUpButton) loginSignUpButton.addEventListener('click', () => window.location.href='/signUp');

 // Login form 
    if (loginForm) {
        loginForm.addEventListener('submit', function (event) {

            event.preventDefault(); // Prevent default form submission
    
            const loginData = {
                userOremail: document.getElementById('userOremail').value,
                password: document.getElementById('pass').value
            };
    
            fetch('http://localhost:8080/login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(loginData),
                credentials: 'include' // Allows cookies to be sent with the request
            })
            .then(response => response.json())
            .then(data => {
                console.log(data.message);
                if (data.message === "Login successful.") {
                    Chatusername = data.username;
                    // localStorage.setItem('username', data.username); // Store username in localStorage
                    checkSession(); // Refresh session check
                    loginForm.reset();
                    window.location.href='/posts'; // Navigate to posts section
                    // loadPosts(); // Load posts after navigating to the posts section
                } else {
                    const error =  document.getElementById("logerror")
                    error.innerHTML = data.message;
                    checkSession(); // Update UI based on session status
                    loginForm.reset();
                }
            })
            .catch(error => console.log(error));
        });
    }