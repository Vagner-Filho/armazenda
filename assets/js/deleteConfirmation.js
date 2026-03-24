/**
 * Initializes delete confirmation buttons that use data-delete-text attribute.
 * Shows a SweetAlert2 confirmation dialog before triggering htmx delete.
 */

function initDeleteButtons() {
	document.querySelectorAll('button[hx-trigger="confirmed"]').forEach(button => {
		if (button.dataset.deleteInitialized) return;
		button.dataset.deleteInitialized = 'true';
		button.addEventListener('click', async (event) => {
			event.preventDefault();
			const text = button.dataset.deleteText;
			if (!window.Swal || !window.htmx) {
				console.error('Swal or htmx not loaded');
				return;
			}
			const result = await window.Swal.fire({
				title: 'Excluir',
				text: text,
				showCancelButton: true,
				cancelButtonText: 'Cancelar',
				cancelButtonColor: '#ff7200',
				reverseButtons: true,
				confirmButtonText: 'Excluir',
				confirmButtonColor: '#22c55e'
			});
			if (result.isConfirmed) {
				window.htmx.trigger(button, 'confirmed');
			}
		});
	});
}

document.addEventListener('DOMContentLoaded', initDeleteButtons);
document.addEventListener('htmx:afterSwap', initDeleteButtons);
