const tabs = document.querySelector("#tabs");
const fieldsets = document.querySelectorAll("fieldset");
const nfeForm = document.querySelector("form#detached-nfe-form");

/**
 * Fieldset group per tab. The pagamento (detPag/tPag) and entrega
 * (<entrega>) fieldsets were introduced by the NF-e emission wiring
 * (docs/sdd/detached_nfe_input_wiring): payment is transaction data, so it
 * shares the Transação tab, and the delivery site belongs with the parties,
 * sharing the Destinatário tab.
 */
const tabFieldsets = {
    dest: ["dest", "entrega-fieldset"],
    transac: ["transac", "pagamento-fieldset"],
    transp: ["transporte-fieldset"],
    trib: ["trib"],
};

if (tabs && fieldsets) {
    /**
     * Reveal the fieldset group of a tab; hides every other fieldset.
     * @param {string} identifier tab button id
     */
    function showTab(identifier) {
        for (const fd of fieldsets) {
            fd.classList.add('hidden');
        }
        for (const fdId of (tabFieldsets[identifier] || [])) {
            const fdst = nfeForm
                ? nfeForm.querySelector(`fieldset#${fdId}`)
                : document.querySelector(`fieldset#${fdId}`);
            if (fdst) {
                fdst.classList.remove('hidden');
            }
        }
    }

    /**
     * @param {PointerEvent} pe
    */
    function handleTabClick(pe) {
        pe.stopPropagation();

        const identifier = pe.currentTarget.id;

        if (tabs.classList.toString().includes(identifier)) {
            return;
        }

        for (const cls of tabs.classList.values()) {
            if (cls.includes('active')) {
                tabs.classList.remove(cls);
                break;
            }
        }

        showTab(identifier);

        tabs.classList.add(`active-${identifier}`);

        const previewBtn = nfeForm ? nfeForm.querySelector('button[type="submit"]#preview-nfe') : null;
        if (previewBtn) {
            identifier === 'trib' ?
            previewBtn.removeAttribute("disabled") :
            previewBtn.setAttribute("disabled", "true");
        }
    }

    // The template ships with `active-dest` on the tabs bar; apply the same
    // reveal on load so co-tabbed fieldsets (entrega, pagamento) are
    // reachable without a first click — their own markup starts hidden.
    const initialTab = Array.from(tabs.classList).find((cls) => cls.startsWith('active-'));
    if (initialTab) {
        showTab(initialTab.replace('active-', ''));
    }

    for (const btn of tabs.querySelectorAll("button")) {
        btn.addEventListener('click', handleTabClick);
    }
}
