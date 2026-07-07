function handleNFeResponse(evt) {
    const reqPath = evt.detail.requestConfig.path;
    if (evt.detail.successful && reqPath.includes('/nfe/build/') && evt.detail.requestConfig.verb === 'post') {
        const xml = evt.detail.xhr.responseText;
        const blob = new Blob([xml], { type: 'application/xml' });
        const url = URL.createObjectURL(blob);
        const link = document.querySelector('#nfe-download-link');
        if (link) {
            link.href = url;
            const departureId = Number(reqPath.slice(reqPath.lastIndexOf('/')));
            link.download = 'nfe_' + departureId + '.xml';
        }
        const resultEl = document.querySelector('#nfe-result');
        if (resultEl) {
            resultEl.classList.remove('hidden');
        }
    }
}

// Wires the "Usar taxa padrão" checkbox to the three tax rate inputs.
// When checked, the inputs are readonly (showing the product config defaults).
// When unchecked, the inputs become editable so the user can override.
function setupTaxRateToggle(dialogEl) {
    const checkbox = dialogEl.querySelector('#useDefaultTaxRates');
    if (!checkbox) return;
    const rateInputs = [
        dialogEl.querySelector('#icmsRate'),
        dialogEl.querySelector('#pisRate'),
        dialogEl.querySelector('#cofinsRate'),
    ].filter(Boolean);
    if (rateInputs.length === 0) return;

    function applyState() {
        const useDefault = checkbox.checked;
        rateInputs.forEach(input => {
            input.readOnly = useDefault;
        });
    }

    checkbox.addEventListener('change', applyState, { signal: dialogEl._nfeAbortSignal });
    applyState();
}

export function setupNFeForm() {
    if (!window.__nfeFormListenerAdded) {
        document.addEventListener('htmx:afterRequest', handleNFeResponse);
        window.__nfeFormListenerAdded = true;
    }
    const dialogEl = document.querySelector('#nfe-emit-dialog');

    const controller = new AbortController()
    const signal = controller.signal
    dialogEl._nfeAbortSignal = signal

    function closeNFeFormDialog() {
        if (dialogEl.isPaused) return
        controller.abort()
        dialogEl.close()
        dialogEl.remove()
    }

    if (dialogEl) {
        dialogEl.showModal();
        dialogEl.addEventListener('close', closeNFeFormDialog, { signal })
        setupTaxRateToggle(dialogEl);
    }
}
