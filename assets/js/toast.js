const SUCCESS = 0
const WARNING = 1
const ERROR = 2
const INFO = 3

const successIcon = `<svg id="success-icon" class="shrink-0 size-4 text-teal-500 mt-0.5" xmlns="http://www.w3.org/2000/svg" width="16" height="16"
                fill="currentColor" viewBox="0 0 16 16">
                <path
                    d="M16 8A8 8 0 1 1 0 8a8 8 0 0 1 16 0zm-3.97-3.03a.75.75 0 0 0-1.08.022L7.477 9.417 5.384 7.323a.75.75 0 0 0-1.06 1.06L6.97 11.03a.75.75 0 0 0 1.079-.02l3.992-4.99a.75.75 0 0 0-.01-1.05z">
                </path>
            </svg>`

const infoIcon = `<svg id="info-icon" class="shrink-0 size-4 text-blue-500 mt-0.5" xmlns="http://www.w3.org/2000/svg" width="16" height="16"
                fill="currentColor" viewBox="0 0 16 16">
                <path
                    d="M8 16A8 8 0 1 0 8 0a8 8 0 0 0 0 16zm.93-9.412-1 4.705c-.07.34.029.533.304.533.194 0 .487-.07.686-.246l-.088.416c-.287.346-.92.598-1.465.598-.703 0-1.002-.422-.808-1.319l.738-3.468c.064-.293.006-.399-.287-.47l-.451-.081.082-.381 2.29-.287zM8 5.5a1 1 0 1 1 0-2 1 1 0 0 1 0 2z">
                </path>
            </svg>`

const warningIcon = `<svg id="warning-icon" class="shrink-0 size-4 text-yellow-500 mt-0.5" xmlns="http://www.w3.org/2000/svg" width="16"
                height="16" fill="currentColor" viewBox="0 0 16 16">
                <path
                    d="M16 8A8 8 0 1 1 0 8a8 8 0 0 1 16 0zM8 4a.905.905 0 0 0-.9.995l.35 3.507a.552.552 0 0 0 1.1 0l.35-3.507A.905.905 0 0 0 8 4zm.002 6a1 1 0 1 0 0 2 1 1 0 0 0 0-2z">
                </path>
            </svg>`

const errorIcon = `<svg id="error-icon" class="shrink-0 size-4 text-red-500 mt-0.5" xmlns="http://www.w3.org/2000/svg" width="16" height="16"
                fill="currentColor" viewBox="0 0 16 16">
                <path
                    d="M16 8A8 8 0 1 1 0 8a8 8 0 0 1 16 0zM5.354 4.646a.5.5 0 1 0-.708.708L7.293 8l-2.647 2.646a.5.5 0 0 0 .708.708L8 8.707l2.646 2.647a.5.5 0 0 0 .708-.708L8.707 8l2.647-2.646a.5.5 0 0 0-.708-.708L8 7.293 5.354 4.646z">
                </path>
            </svg>`

const iconMp = new Map([
	[SUCCESS, successIcon],
	[WARNING, warningIcon],
	[ERROR, errorIcon],
	[INFO, infoIcon],
])

class ToastManager {
	makeToast(message, hint, type) {
		const container = document.createElement('div')
		container.classList.add(...["max-w-xs", "bg-white", "border", "border-gray-200", "rounded-xl", "shadow-lg", "fixed"])

		container.setAttribute("style", "top: 8px; right: 8px;")
		container.setAttribute("role", "alert")
		container.setAttribute("tabindex", "-1")
		container.setAttribute("aria-live", "assertive")
		container.setAttribute("aria-atomic", "true")
		container.setAttribute("aria-labelledby", "armazenda-toast")

		const toastBody = document.createElement('div')
		toastBody.classList.add(...["flex", "p-4"])

		const iconContainer = document.createElement('div')
		iconContainer.classList.add('shrink-0')
		iconContainer.setAttribute('id', 'toast-icon-container')
		const literalIcon = iconMp.get(type) ?? infoIcon
		iconContainer.innerHTML = literalIcon

		toastBody.append(iconContainer)

		const messageContainer = document.createElement('div')
		messageContainer.classList.add('ms-3')

		const messageParagraph = document.createElement('p')
		messageParagraph.classList.add(...["text-sm", "text-gray-700"])
		messageParagraph.setAttribute('id', 'armazenda-toast')
		messageParagraph.textContent = message

		messageContainer.append(messageParagraph)

		toastBody.append(messageContainer)
		container.append(toastBody)

		return container
	}

	showToast(toast) {
		if (!this.toastQueue) {
			this.toastQueue = []
		}
		if (this.toastQueue.length > 0) {
			for (let i = 0; i < this.toastQueue.length; i++) {
				this.toastQueue[i].style.top = `${64 * (this.toastQueue.length - i)}px`
			}
		}
		this.toastQueue.push(toast)

		document.body.append(toast)

		const t = setTimeout(() => {
			const toRemove = this.toastQueue.shift()
			if (toRemove) {
				document.body.removeChild(toRemove)
			}
			clearTimeout(t)
		}, 3000)
	}

	constructor(message, hint, type) {
		this.message = message;
		this.hint = hint;
		this.type = type;

		this.makeToast(this.message, this.hint, this.type)
	}
}

const tm = new ToastManager()
document.body.addEventListener("toast", (evt) => {
	const toast = tm.makeToast(evt.detail.Message, evt.detail.Hint, evt.detail.Type)
	tm.showToast(toast)
})
