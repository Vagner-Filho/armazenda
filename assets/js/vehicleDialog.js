import { handleNewOption } from "./selectOption.js"
function vehicleDialogSetup() {
    const dialogEl = document.querySelector("dialog#vehicleFormDialog")
    if (dialogEl) {
        dialogEl.showModal()

        const controller = new AbortController()
        const signal = controller.signal

        function closeVehicleForm() {
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
        handleNewOption(closeVehicleForm, signal)
    }
}

export { vehicleDialogSetup }