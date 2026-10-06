import {
    deriveOriginDestination,
    initCfopSelectors,
    refreshCfopOptions,
    takePendingOriginSelect,
} from "./cfopSelector.js"

/**
 * Wires the "Cadastrar CFOP" dialog: modal lifecycle, client-side
 * Origem/Destino derivation (still user-editable), and the post-create
 * farm-ordered options refresh. Failures keep the dialog open; their pt-br
 * toast comes from the HX-Trigger response header.
 */
function cfopDialogSetup() {
    initCfopSelectors()

    const dialogEl = document.querySelector("dialog#nfeCfopFormDialog")
    if (dialogEl) {
        dialogEl.showModal()

        const controller = new AbortController()
        const signal = controller.signal

        function closeCfopForm() {
            controller.abort()
            dialogEl.close()
            dialogEl.remove()
            window.closeCfopForm = undefined
        }
        dialogEl.addEventListener('close', closeCfopForm, { signal })
        window.closeCfopForm = closeCfopForm
        const cancelButton = dialogEl.querySelector('.cancel-btn')
        if (cancelButton) {
            cancelButton.addEventListener('click', closeCfopForm, { signal })
        }

        const codeInput = dialogEl.querySelector('input[name="code"]')
        const originSelect = dialogEl.querySelector('select[name="originDestination"]')
        if (codeInput && originSelect) {
            codeInput.addEventListener('input', () => {
                const derived = deriveOriginDestination(codeInput.value)
                if (derived) {
                    originSelect.value = derived
                }
            }, { signal })
        }

        const form = dialogEl.querySelector('form')
        if (form) {
            form.addEventListener('htmx:afterRequest', async (event) => {
                if (!event.detail.successful) {
                    // Keep the modal open so the user can fix the values.
                    return
                }

                const code = (codeInput ? codeInput.value : '').trim()
                const targetSelect = takePendingOriginSelect()
                try {
                    await refreshCfopOptions(code, targetSelect)
                } catch (error) {
                    // Refreshing is a convenience; the success toast already
                    // confirmed the registration. Never block the close.
                }
                closeCfopForm()
            }, { signal })
        }
    }
}

export { cfopDialogSetup }
