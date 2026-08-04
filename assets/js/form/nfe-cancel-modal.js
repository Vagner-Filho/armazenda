export function setupNFeCancelModal() {
	const dialogEl = document.querySelector('#nfe-cancel-dialog');
	if (!dialogEl) {
		return;
	}

	dialogEl.addEventListener('close', () => {
		dialogEl.remove();
	});
	dialogEl.showModal();

	// Close the dialog after a successful cancellation. On failure the server
	// responds with 4xx/5xx (no swap) and only a toast is shown, keeping the
	// dialog open so the user can fix the justification or retry.
	document.body.addEventListener('htmx:afterRequest', function handler(evt) {
		const reqPath = evt.detail.requestConfig.path;
		if (
			evt.detail.successful &&
			evt.detail.requestConfig.verb === 'post' &&
			reqPath.includes('/nfe/cancel/')
		) {
			dialogEl.close();
			document.body.removeEventListener('htmx:afterRequest', handler);
		}
	});
}
