import { formatDateToInput } from "./date.js"
import { applyDiscounts } from "./discount.js"
import { removeEmptyFields } from "./form.js"

export function setupEntryForm(payload) {
    const humidityInput = document.querySelector('input#humidity')
    if (humidityInput) {
        humidityInput.addEventListener('input', applyDiscounts)
    }

    const damageInput = document.querySelector('input#damage')
    if (damageInput) {
        damageInput.addEventListener('input', applyDiscounts)
    }

    const impurityInput = document.querySelector('input#impurity')
    if (impurityInput) {
        impurityInput.addEventListener('input', applyDiscounts)
    }

    const grossWeightInput = document.querySelector('input#grossWeight')
    if (grossWeightInput) {
        grossWeightInput.addEventListener('input', applyDiscounts)
    }

    const tareInput = document.querySelector('input#tare')
    if (tareInput) {
        tareInput.addEventListener('input', applyDiscounts)
    }

    const dateVal = formatDateToInput(payload.arrivalDate)

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

        const grossWeightInput = dialogEl.querySelector('input#grossWeight')
        const tareInput = dialogEl.querySelector('input#tare')

        let grossWeightValue = Number(grossWeightInput ? grossWeightInput.value : 0)
        let tareValue = Number(tareInput ? tareInput.value : 0)

        const netWeightInput = dialogEl.querySelector('input#netWeight')
        const rawNetWeightInput = dialogEl.querySelector('input#netWeightRaw')
        if (netWeightInput && rawNetWeightInput) {
            if (netWeightInput.value !== "") {
                netWeightInput.dataset.raw = grossWeightValue - tareValue
                applyDiscounts()
            }

            grossWeightInput.addEventListener('input', (e) => {
                grossWeightValue = Number(e.target.value) ?? 0

                netWeightInput.value = grossWeightValue - tareValue
                rawNetWeightInput.value = grossWeightValue - tareValue
                netWeightInput.dataset.raw = grossWeightValue - tareValue
                applyDiscounts()
            })
            tareInput.addEventListener('input', (e) => {
                tareValue = Number(e.target.value) ?? 0

                netWeightInput.value = grossWeightValue - tareValue
                rawNetWeightInput.value = grossWeightValue - tareValue
                netWeightInput.dataset.raw = grossWeightValue - tareValue
                applyDiscounts()
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

    function setPersonConfig(el) {
        const selectedOption = el.options[el.selectedIndex];
        const humidityDiscount = selectedOption.getAttribute('data-humidity');
        const entrySoyDiscount = selectedOption.getAttribute('data-entry-soy-discount');
        const entryCornDiscount = selectedOption.getAttribute('data-entry-corn-discount');
        const personId = selectedOption.value;
        const personName = selectedOption.textContent.trim();

        sessionStorage.setItem('personConfig', JSON.stringify({
            humidityDiscount: humidityDiscount,
            entrySoyDiscount: entrySoyDiscount,
            entryCornDiscount: entryCornDiscount
        }));

        applyDiscounts();
    }
    const selector = document.getElementById('origin-selector');
    if (selector) {
        selector.addEventListener('change', function(evt) {
            setPersonConfig(evt.target);
        });
        const option = selector.querySelector('option[selected]')

        if (option && option.dataset.humidity) {
            sessionStorage.setItem('personConfig', JSON.stringify({
                humidityDiscount: option.dataset.humidity,
                entrySoyDiscount: option.dataset.entrySoyDiscount,
                entryCornDiscount: option.dataset.entryCornDiscount
            }));
        }
    }
}

