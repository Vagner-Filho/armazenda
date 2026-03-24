import { handleNewOption } from "./selectOption.js"
function fieldDialogSetup() {
    const dialogEl = document.querySelector("dialog#fieldFormDialog")
    if (dialogEl) {
        dialogEl.showModal()
        function closeFieldForm() {
            dialogEl.close()
            dialogEl.remove()
            window.closeFieldForm = undefined
        }
        dialogEl.addEventListener('close', closeFieldForm)
        window.closeFieldForm = closeFieldForm
        const cancelButton = dialogEl.querySelector('.cancel-btn');
        if (cancelButton) {
            cancelButton.addEventListener('click', closeFieldForm);
        }
        handleNewOption(closeFieldForm)
    }
}

export { fieldDialogSetup }
