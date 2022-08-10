
const loginForm = document.querySelector('form')
const body = document.querySelector('body')
const errorMessageElement = document.querySelector('.error-message')
const submitButton = loginForm.querySelector('button')

loginForm.addEventListener('submit', async (e) => {
    e.preventDefault()
    submitButton.innerText = "Loading..."

    const formData = new FormData()
    formData.append('email', document.getElementById('email').value);
    formData.append('password', document.getElementById('password').value);

    console.log(formData)

    // send request with axios
    try {
        const res = await fetch("/api/v1/auth/login?", {
            method: 'post',
            body: formData
        })

        if (!res.ok) {
            const error = await res.json()
            errorMessageElement.innerText = '* ' + error.message
            errorMessageElement.style.color = "red"
            submitButton.innerHTML = "Authorize"
            
            throw new Error(error)
        }

        const resData = await res.json()
        console.log(resData)

        const res1 = await fetch(`/api/v1/oauth/authorize/handshake${window.location.search}&userId=${resData.id}`)
        const handshakeData = await res1.json()

        console.log(handshakeData)

        await fetch(handshakeData.redirectLink)
        

        window.history.back()
    } catch(err) {
        console.log(err, "here")
        document.getElementById('email').value = ""
        document.getElementById('password').value = ""
    }
   
   console.log("submmitted")
})