const changePasswordForm = document.querySelector('form')
const changePassordButton = changePasswordForm.querySelector('button')

const params = new URLSearchParams(window.location.search);
const resetToken = params.get("token")

changePasswordForm.addEventListener('submit', async (e) => {
    e.preventDefault()

    const formData = new FormData()
    formData.append('password', document.getElementById('new-password').value)
    formData.append('confirmPassword', document.getElementById('confirm-password').value)

    const res = await fetch(`/api/v1/users/reset-password/${resetToken}`, {
        method: 'PATCH',
        body: formData
    })

    if (!res.ok) {
        const data = await res.text()

        return 
    }

    const data = await res.text()
})
