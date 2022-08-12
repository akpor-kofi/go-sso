const signupForm = document.querySelector('form')
const buttonElement = signupForm.querySelector('.signup-btn');

signupForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    buttonElement.innerText = "Loading..."

    // check password here

    const formData = new FormData()
    formData.append('name', document.getElementById('name').value);
    formData.append('email', document.getElementById('email').value);
    formData.append('phoneNumber', document.getElementById('phoneNumber').value);
    formData.append('dob', document.getElementById('dob').value);
    formData.append('password', document.getElementById('password').value);
    formData.append('imageFile', document.getElementById('file-upload').files[0]);

    const res = await fetch("/api/v1/auth/signup", {
        method: 'POST',
        body: formData
    })

    if (!res.ok) {

        buttonElement.innerText = "Sign Up"
        return
    }

    window.location = "/static/checkEmail.html"
})