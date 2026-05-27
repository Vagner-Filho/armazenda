import { handleNewOption } from "./selectOption.js"
function vehicleDialogSetup() {
    const dialogEl = document.querySelector("dialog#vehicleFormDialog")
    if (dialogEl) {
        dialogEl.showModal()

        const controller = new AbortController()
        const signal = controller.signal
        let isConfirming = false

        function closeVehicleForm() {
            if (isConfirming) return
            controller.abort()
            dialogEl.close()
            dialogEl.remove()
            window.closeVehicleForm = undefined
        }
        dialogEl.addEventListener('close', closeVehicleForm, { signal })
        window.closeVehicleForm = closeVehicleForm
        const cancelButton = dialogEl.querySelector('.cancel-btn');
        if (cancelButton) {
            cancelButton.addEventListener('click', closeVehicleForm, { signal });
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

                const plate = form.querySelector('input#vehicle-plate')?.value || '';
                const name = form.querySelector('input#vehicle-name')?.value || '';

                const result = await window.Swal.fire({
                    title: 'Confirmar cadastro de veículo',
                    html: `<div style="text-align:left"><strong>Placa:</strong> ${plate}${name ? '<br><strong>Nome:</strong> ' + name : ''}</div>`,
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

        handleNewOption(closeVehicleForm, signal)
    }
}

export { vehicleDialogSetup }
