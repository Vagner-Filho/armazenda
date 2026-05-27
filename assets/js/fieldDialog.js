import { handleNewOption } from "./selectOption.js"
function fieldDialogSetup() {
    const dialogEl = document.querySelector("dialog#fieldFormDialog")
    if (dialogEl) {
        dialogEl.showModal()

        const controller = new AbortController()
        const signal = controller.signal
        let isConfirming = false

        function closeFieldForm() {
            if (isConfirming) return
            controller.abort()
            dialogEl.close()
            dialogEl.remove()
            window.closeFieldForm = undefined
        }
        dialogEl.addEventListener('close', closeFieldForm, { signal })
        window.closeFieldForm = closeFieldForm
        const cancelButton = dialogEl.querySelector('.cancel-btn');
        if (cancelButton) {
            cancelButton.addEventListener('click', closeFieldForm, { signal });
        }

        const form = dialogEl.querySelector('form');
        if (form) {
            form.addEventListener('submit', async (event) => {
                if (form.dataset.skipConfirm) {
                    delete form.dataset.skipConfirm
                    return
                }
                event.preventDefault();
                event.stopImmediatePropagation();

                isConfirming = true

                const parentDialog = document.querySelector('dialog#addEntryDialog, dialog#departure-form-dialog');
                if (parentDialog) {
                    parentDialog.isPaused = true;
                    parentDialog.close();
                }
                dialogEl.close()

                const name = form.querySelector('input#field-name')?.value || '';
                const hectares = form.querySelector('input#hectares')?.value || '';

                const result = await window.Swal.fire({
                    title: 'Confirmar cadastro de talhão',
                    html: `<div style="text-align:left"><strong>Nome:</strong> ${name}<br><strong>Hectares:</strong> ${hectares}</div>`,
                    icon: 'question',
                    showCancelButton: true,
                    cancelButtonText: 'Cancelar',
                    cancelButtonColor: '#ff7200',
                    reverseButtons: true,
                    confirmButtonText: 'Confirmar',
                    confirmButtonColor: '#22c55e',
                    theme: 'dark',
                });

                if (result.isConfirmed) {
                    form.dataset.skipConfirm = 'true';
                    form.requestSubmit();
                }

                if (parentDialog) {
                    parentDialog.showModal();
                    parentDialog.isPaused = false;
                }
                dialogEl.showModal()
                isConfirming = false
            }, { capture: true, signal });
        }

        handleNewOption(closeFieldForm, signal)
    }
}

export { fieldDialogSetup }
