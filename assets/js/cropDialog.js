import { handleNewOption } from "./selectOption.js"
import { setTodayDatetimeInput } from "./date.js"
function cropDialogSetup() {
    const dialogEl = document.querySelector("dialog#cropFormDialog")
    if (dialogEl) {
        setTodayDatetimeInput("input#start-date", true)

        dialogEl.showModal()

        const controller = new AbortController()
        const signal = controller.signal
        let isConfirming = false

        function closeCropForm() {
            if (isConfirming) return
            controller.abort()
            dialogEl.close()
            dialogEl.remove()
            window.closeCropForm = undefined
        }
        dialogEl.addEventListener('close', closeCropForm, { signal })
        window.closeCropForm = closeCropForm
        const cancelButton = dialogEl.querySelector('.cancel-btn');
        if (cancelButton) {
            cancelButton.addEventListener('click', closeCropForm, { signal });
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

                const name = form.querySelector('input#crop-name')?.value || '';
                const grain = form.querySelector('select#grain-selector')?.selectedOptions[0]?.textContent || '';
                const startDate = form.querySelector('input#start-date')?.value || '';

                const result = await window.Swal.fire({
                    title: 'Confirmar cadastro de safra',
                    html: `<div style="text-align:left"><strong>Nome:</strong> ${name}<br><strong>Grão:</strong> ${grain}<br><strong>Data de início:</strong> ${startDate}</div>`,
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

        handleNewOption(closeCropForm, signal)
    }
}

export { cropDialogSetup }
