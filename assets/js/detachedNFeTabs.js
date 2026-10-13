const tabs = document.querySelector("#tabs");
const fieldsets = document.querySelectorAll("fieldset");
const nfeForm = document.querySelector("form#detached-nfe-form");
if (tabs && fieldsets) {
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

        for (const fd of fieldsets) {
            fd.classList.add('hidden');
        }

        if (identifier === 'transp') {
            const transpFdst = nfeForm.querySelector('fieldset#transporte-fieldset');
            if (transpFdst) {
                transpFdst.classList.remove('hidden');
            }
        } else {
            const targetFdst = nfeForm.querySelector(`fieldset#${identifier}`);
            if (targetFdst) {
                targetFdst.classList.remove('hidden');
            }
        }

        tabs.classList.add(`active-${identifier}`);

        const previewBtn = nfeForm.querySelector('button[type="submit"]#preview-nfe');
        if (previewBtn) {
            identifier === 'trib' ?
            previewBtn.removeAttribute("disabled") :
            previewBtn.setAttribute("disabled", "true");
        }
    }
    for (const btn of tabs.querySelectorAll("button")) {
        btn.addEventListener('click', handleTabClick);
    }
}
