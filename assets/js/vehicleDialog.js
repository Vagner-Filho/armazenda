import { handleNewOption } from "./selectOption.js"
function vehicleDialogSetup() {
    const dialogEl = document.querySelector("dialog#vehicleFormDialog")
    if (dialogEl) {
        dialogEl.showModal()

        function closeVehicleForm() {
            dialogEl.close()
            dialogEl.remove()
            window.closeVehicleForm = undefined
        }
        dialogEl.addEventListener('close', closeVehicleForm)
        window.closeVehicleForm = closeVehicleForm
        handleNewOption(closeVehicleForm)
    }
}

export { vehicleDialogSetup }
