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

export function setupNFeForm() {
    if (!window.__nfeFormListenerAdded) {
        document.addEventListener('htmx:afterRequest', handleNFeResponse);
        window.__nfeFormListenerAdded = true;
    }
    const dialogEl = document.querySelector('#nfe-emit-dialog');
    if (dialogEl) {
        dialogEl.showModal();
    }
}
