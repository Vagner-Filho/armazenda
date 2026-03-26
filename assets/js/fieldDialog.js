import { handleNewOption } from "./selectOption.js"
function fieldDialogSetup() {
    const dialogEl = document.querySelector("dialog#fieldFormDialog")
    if (dialogEl) {
        dialogEl.showModal()

        const controller = new AbortController()
        const signal = controller.signal

        function closeFieldForm() {
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
        handleNewOption(closeFieldForm, signal)
    }
}

export { fieldDialogSetup }