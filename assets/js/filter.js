async function setupClearFilter(clearElLocator, filterTriggerEl) {
    const clearEl = document.querySelector(clearElLocator)
    if (!clearEl) return

    // Derive the reset endpoint from the clear button's own hx-get so the
    // helper works for any page sharing this filter pattern.
    const resetPath = clearEl.getAttribute('hx-get') || clearEl.getAttribute('data-hx-get')

    // Derive the form selector from the "Filtrar" button's form attribute,
    // falling back to the id embedded in the clear button's hx-target.
    const triggerEl = filterTriggerEl ? document.querySelector(filterTriggerEl) : null
    let formSelector = null
    if (triggerEl) {
        const formId = triggerEl.getAttribute('form')
        if (formId) formSelector = `#${formId}`
    }
    if (!formSelector) {
        const hxTarget = clearEl.getAttribute('hx-target') || clearEl.getAttribute('data-hx-target') || ''
        const m = hxTarget.match(/#([A-Za-z0-9_-]+)/)
        if (m) formSelector = `#${m[1]}`
    }

    let pendingClear = false
    clearEl.addEventListener('click', () => { pendingClear = true })

    document.body.addEventListener('htmx:afterSwap', (evt) => {
        if (!pendingClear) return
        if (!(evt.detail.successful &&
              (!resetPath || evt.detail.requestConfig.path === resetPath) &&
              evt.detail.requestConfig.verb === 'get')) return

        pendingClear = false
        const filterEl = formSelector ? document.querySelector(formSelector) : null
        if (!filterEl) return

        // Defer so HTMX finishes processing the swapped form, then trigger
        // its hx-post the same way a real submit does.
        setTimeout(() => {
            if (window.htmx) {
                window.htmx.process(filterEl)
                window.htmx.trigger(filterEl, 'submit')
            } else {
                filterEl.requestSubmit()
            }
        }, 0)
    })
}

export { setupClearFilter }
