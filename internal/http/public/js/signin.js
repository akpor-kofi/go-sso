
const loginForm = document.querySelector('form')
const errorMessageElement = document.querySelector('.error-message')
const submitButton = loginForm.querySelector('button')

loginForm.addEventListener('submit', async (e) => {
    e.preventDefault()
    submitButton.innerText = "Loading..."

    const formData = new FormData()
    formData.append('email', document.getElementById('email').value);
    formData.append('password', document.getElementById('password').value);

    // send request with axios
    try {
        const res = await fetch("/api/v1/auth/login", {
            method: 'post',
            body: formData
        })

        if (!res.ok) {
            const error = await res.json()
            errorMessageElement.innerHTML = `<p class="error-paragraph">* ${error.message}</p>`
            submitButton.innerHTML = "Authorize"
            
            throw new Error(error)
        }

        const resData = await res.json()

        await fetch(`/api/v1/oauth/authorize/handshake${window.location.search}&userId=${resData.id}`, {
            headers: {
                "Access-Control-Allow-Origin": "*"
            }
        })


    } catch(err) {
        document.getElementById('email').value = ""
        document.getElementById('password').value = ""
    }
   
})