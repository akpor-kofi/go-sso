const forgotPasswordForm = document.querySelector('form')
const submitButton = forgotPasswordForm.querySelector('button')

forgotPasswordForm.addEventListener('submit', async (e) => {
    e.preventDefault()
    submitButton.innerText = "Loading..."

    const formData = new FormData()
    formData.append('email', document.getElementById('email').value);

    const res = await fetch("/api/v1/users/forgot-password", {
        method: 'post',
        body: formData
    })

    if (!res.ok) {
        submitButton.innerText = "Send"

        return
    }

    submitButton.innerText = "Sent"
})