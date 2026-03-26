import { handleNewOption } from "./selectOption.js"
import { setTodayDatetimeInput } from "./date.js"
function cropDialogSetup() {
    const dialogEl = document.querySelector("dialog#cropFormDialog")
    if (dialogEl) {
        setTodayDatetimeInput("input#start-date", true)

        dialogEl.showModal()

        const controller = new AbortController()
        const signal = controller.signal

        function closeCropForm() {
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
        handleNewOption(closeCropForm, signal)
    }
}

export { cropDialogSetup }