import { formatDateToInput } from "./date.js";

export function setupPersonCND(date, expDateSelector) {
    if (date) {
        const expDateInput = document.querySelector(expDateSelector ?? "input#expDate");
        if (expDateInput) {
            const expDate = formatDateToInput(date, true);
            expDateInput.setAttribute("value", expDate);
            if (new Date(expDate) < new Date()) {
                const label = expDateInput.nextElementSibling
                label.innerHTML += `<small class="text-red-500 font-semibold">o certificado expirou!</small>`;
            }
        }
    }

    let cndFieldsets = document.querySelectorAll('fieldset#pessoa-cnd-fieldset');
    if (!cndFieldsets || cndFieldsets.length === 0) {
        cndFieldsets = document.querySelectorAll('fieldset#farm-cnd-fieldset');
    }
    for (const fdset of cndFieldsets.values()) {
        const addMetaBtn = fdset.querySelector('#addMetaKeyValue');
        const metaContainer = fdset.querySelector('#meta-container');

        if (addMetaBtn && metaContainer) {
            const metaRow = metaContainer.querySelector('#meta-row');
            addMetaBtn.addEventListener('click', function handleAddClick() {
                const newMetaRow = metaRow.cloneNode(true);
                for (const c of newMetaRow.childNodes) {
                    if (c instanceof HTMLInputElement) {
                        c.value = "";
                        if (c.id === 'meta-key') {
                            c.setAttribute("name", `meta[${metaContainer.childElementCount}].key`)
                        } else if (c.id === 'meta-value') {
                            c.setAttribute("name", `meta[${metaContainer.childElementCount}].value`)
                        }
                    }
                    if (c instanceof HTMLButtonElement) {
                        c.classList.remove('invisible');
                        c.addEventListener('click', function handleMinusClick(evt) {
                            evt.target.parentElement.remove();
                        })
                    }
                }

                metaContainer.append(newMetaRow);
            });
        }
    }
}
