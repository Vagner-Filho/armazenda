export function closeModal(selector) {
	const dialogEl = document.querySelector(selector)
	if (dialogEl) {
		dialogEl.close()
		dialogEl.remove()
	}
}
