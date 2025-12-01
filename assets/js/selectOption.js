/**
 * Handles the selection of a newly added option in a select element after an HTMX POST request.
 * Automatically selects the last option in the select element and optionally closes a dialog.
 * @param {Function} closeDialogCallback - Optional callback function to close a dialog after selection
 */
function handleNewOption(closeDialogCallback) {
    document.addEventListener("htmx:afterSwap", (evt) => {
        if (evt.detail.requestConfig.verb === 'post' && evt.detail.xhr.status === 201 && evt.target instanceof HTMLSelectElement) {
            const target = evt.target
            if (target.children[0].value === '-1') {
                target.children[0].remove()
            }
            for (const c of target.children) {
                if (c instanceof HTMLOptionElement) {
                    if (c.selected) {
                        c.selected = false;
                        break;
                    }
                }
            }
            target.children[target.children.length - 1].selected = true;
            if (closeDialogCallback) {
                closeDialogCallback()
            }
        }
    })

}

export { handleNewOption }
