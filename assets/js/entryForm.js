import { formatDateToInput } from "./date.js"
import { applyDiscounts } from "./discount.js"
import { removeEmptyFields } from "./form.js"

/**
 * Wrapper for async applyDiscounts
 * @returns {Promise<void>}
 */
async function applyDiscountsAsync() {
    await applyDiscounts();
}

export function setupEntryForm(payload) {
    const humidityInput = document.querySelector('input#humidity')
    if (humidityInput) {
        humidityInput.addEventListener('input', applyDiscountsAsync)
    }

    const damageInput = document.querySelector('input#damage')
    if (damageInput) {
        damageInput.addEventListener('input', applyDiscountsAsync)
    }

    const impurityInput = document.querySelector('input#impurity')
    if (impurityInput) {
        impurityInput.addEventListener('input', applyDiscountsAsync)
    }

    const grossWeightInput = document.querySelector('input#grossWeight')
    if (grossWeightInput) {
        grossWeightInput.addEventListener('input', applyDiscountsAsync)
    }

    const tareInput = document.querySelector('input#tare')
    if (tareInput) {
        tareInput.addEventListener('input', applyDiscountsAsync)
    }

    const dateVal = formatDateToInput(payload && typeof payload === "object" ? payload.arrivalDate : null)

    const dialogEl = document.querySelector("#addEntryDialog")
    if (dialogEl) {
        dialogEl.showModal()
        const arrivalDateEl = dialogEl.querySelector("input#arrival-date")
        if (!!arrivalDateEl) {
            arrivalDateEl.setAttribute("value", dateVal)
        }

        function closeEntryFormDialog() {
            dialogEl.close()
            dialogEl.remove()
        }
        dialogEl.addEventListener('close', closeEntryFormDialog)
        window.closeEntryFormDialog = closeEntryFormDialog
        const cancelButton = dialogEl.querySelector('.cancel-btn');
        if (cancelButton) {
            cancelButton.addEventListener('click', closeEntryFormDialog);
        }

        const grossWeightInput = dialogEl.querySelector('input#grossWeight')
        const tareInput = dialogEl.querySelector('input#tare')

        let grossWeightValue = Number(grossWeightInput ? grossWeightInput.value : 0)
        let tareValue = Number(tareInput ? tareInput.value : 0)

        const netWeightInput = dialogEl.querySelector('input#net_weight')
        const rawNetWeightInput = dialogEl.querySelector('input#net_weight_raw')
        if (netWeightInput && rawNetWeightInput) {
            if (netWeightInput.value !== "") {
                netWeightInput.dataset.raw = grossWeightValue - tareValue
                applyDiscountsAsync()
            }

            grossWeightInput.addEventListener('input', (e) => {
                grossWeightValue = Number(e.target.value) ?? 0

                netWeightInput.value = grossWeightValue - tareValue
                rawNetWeightInput.value = grossWeightValue - tareValue
                netWeightInput.dataset.raw = grossWeightValue - tareValue
                applyDiscountsAsync()
            })
            tareInput.addEventListener('input', (e) => {
                tareValue = Number(e.target.value) ?? 0

                netWeightInput.value = grossWeightValue - tareValue
                rawNetWeightInput.value = grossWeightValue - tareValue
                netWeightInput.dataset.raw = grossWeightValue - tareValue
                applyDiscountsAsync()
            })
        }
        document.body.addEventListener('htmx:configRequest', function(evt) {
            evt.detail.parameters = removeEmptyFields(evt.detail.parameters);
        });
        document.body.addEventListener('htmx:afterRequest', function(evt) {
            if (evt.detail.successful && evt.detail.requestConfig.path === '/entry' && (evt.detail.requestConfig.verb === 'post' || evt.detail.requestConfig.verb === 'put')) {
                closeEntryFormDialog()
            }
        });
        document.body.addEventListener('htmx:afterSwap', function(evt) {
            if (evt.detail.successful && evt.detail.requestConfig.path === '/entry' && (evt.detail.requestConfig.verb === 'post' || evt.detail.requestConfig.verb === 'put')) {
                window.formatEntryListItem('#' + evt.detail.elt.firstElementChild.id);
            }
        });
    }

    const cropInput = document.querySelector('select#crop-selector')
    if (cropInput) {
        cropInput.addEventListener('change', function setProduct(ev) {
            const selectedOption = ev.target.options[ev.target.selectedIndex]
            sessionStorage.setItem('product', selectedOption.dataset.productId)
        })
        const selectedOption = cropInput.options[cropInput.selectedIndex]
        sessionStorage.setItem('product', selectedOption.dataset.productId)
    }

    async function setPersonConfig(el) {
        const selectedOption = el.options[el.selectedIndex];
        const humidityProgressionId = selectedOption.getAttribute('data-humidity-progression-id');
        const entrySoyDiscount = selectedOption.getAttribute('data-entry-soy-discount');
        const entryCornDiscount = selectedOption.getAttribute('data-entry-corn-discount');
        const personId = selectedOption.value;
        const personName = selectedOption.textContent.trim();

        sessionStorage.setItem('personConfig', JSON.stringify({
            humidityProgressionId: humidityProgressionId,
            entrySoyDiscount: entrySoyDiscount,
            entryCornDiscount: entryCornDiscount
        }));

        await applyDiscounts();
    }
    const selector = document.getElementById('origin-selector');
    if (selector) {
        selector.addEventListener('change', function(evt) {
            setPersonConfig(evt.target);
        });
        const option = selector.querySelector('option[selected]')

        if (option && option.dataset.humidityProgressionId) {
            sessionStorage.setItem('personConfig', JSON.stringify({
                humidityProgressionId: option.dataset.humidityProgressionId,
                entrySoyDiscount: option.dataset.entrySoyDiscount,
                entryCornDiscount: option.dataset.entryCornDiscount
            }));
        }
    }
}

