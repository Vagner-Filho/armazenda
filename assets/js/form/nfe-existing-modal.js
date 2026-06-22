export function setupNFeExistingModal() {
	const dialogEl = document.querySelector('#nfe-existing-dialog');
	if (dialogEl) {
		dialogEl.addEventListener('close', () => {
			dialogEl.remove();
		});
		dialogEl.showModal();
	}
}
