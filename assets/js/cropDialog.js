import { handleNewOption } from "./selectOption.js"
import { setTodayDatetimeInput } from "./date.js"
function cropDialogSetup() {
    const dialogEl = document.querySelector("dialog#cropFormDialog")
    if (dialogEl) {
        setTodayDatetimeInput("input#start-date", true)

        dialogEl.showModal()
        function closeCropForm() {
            dialogEl.close()
            dialogEl.remove()
            window.closeCropForm = undefined
        }
        dialogEl.addEventListener('close', closeCropForm)
        window.closeCropForm = closeCropForm
        const cancelButton = dialogEl.querySelector('.cancel-btn');
        if (cancelButton) {
            cancelButton.addEventListener('click', closeCropForm);
        }
        handleNewOption(closeCropForm)
    }
}

export { cropDialogSetup }
