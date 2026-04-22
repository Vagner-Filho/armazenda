import { formatDateToInput } from "./date.js"
import { formatWeight, setupWeightInput } from "./weight.js"
import { removeEmptyFields } from "./form.js"

export function setupDepartureForm(payload) {
    const dateVal = formatDateToInput(payload && typeof payload === "object" ? payload.departureDate : null)

    const dialogEl = document.querySelector("#departure-form-dialog")
    if (dialogEl) {
        dialogEl.showModal()

        const departureDateEl = dialogEl.querySelector("input#departure-input")
        if (!!departureDateEl) {
            departureDateEl.setAttribute("value", dateVal)
        }

        const controller = new AbortController()
        const signal = controller.signal

        function closeDepartureForm() {
            controller.abort()
            dialogEl.close()
            dialogEl.remove()
            window.closeDepartureForm = undefined
        }

        dialogEl.addEventListener('close', closeDepartureForm, { signal })
        window.closeDepartureForm = closeDepartureForm
        const cancelButton = dialogEl.querySelector('.cancel-btn');
        if (cancelButton) {
            cancelButton.addEventListener('click', closeDepartureForm, { signal });
        }

        const grossWeightInput = dialogEl.querySelector('input#grossWeight')
        const tareInput = dialogEl.querySelector('input#tare')

        // Setup weight formatting first so dataset.raw is populated
        setupWeightInput(grossWeightInput, signal)
        setupWeightInput(tareInput, signal)

        let grossWeightValue = parseFloat(grossWeightInput?.dataset.raw || 0)
        let tareValue = parseFloat(tareInput?.dataset.raw || 0)

        const netWeightInput = dialogEl.querySelector('input#netWeight')
        if (netWeightInput) {
            grossWeightInput.addEventListener('input', () => {
                grossWeightValue = parseFloat(grossWeightInput.dataset.raw || 0)
                tareValue = parseFloat(tareInput.dataset.raw || 0)
                const net = grossWeightValue - tareValue;
                netWeightInput.dataset.raw = net;
                netWeightInput.value = formatWeight(net);
            }, { signal })
            tareInput.addEventListener('input', () => {
                grossWeightValue = parseFloat(grossWeightInput.dataset.raw || 0)
                tareValue = parseFloat(tareInput.dataset.raw || 0)
                const net = grossWeightValue - tareValue;
                netWeightInput.dataset.raw = net;
                netWeightInput.value = formatWeight(net);
            }, { signal })

            netWeightInput.value = formatWeight(grossWeightValue - tareValue);
        }

        document.body.addEventListener('htmx:configRequest', function(evt) {
            evt.detail.parameters = removeEmptyFields(evt.detail.parameters);
            if (grossWeightInput && grossWeightInput.dataset.raw !== undefined) {
                evt.detail.parameters.grossWeight = grossWeightInput.dataset.raw;
            }
            if (tareInput && tareInput.dataset.raw !== undefined) {
                evt.detail.parameters.tare = tareInput.dataset.raw;
            }
        }, { signal });

        document.body.addEventListener('htmx:afterRequest', function(evt) {
            if (evt.detail.successful && evt.detail.requestConfig.path === '/departure' && (evt.detail.requestConfig.verb === 'post' || evt.detail.requestConfig.verb === 'put')) {
                closeDepartureForm()
            }
        }, { signal });

        const putDeparturePattern = new RegExp(/departure\/[0-9]+$/gm);
        document.body.addEventListener('htmx:afterSwap', function(evt) {
            if (evt.detail.successful && (evt.detail.requestConfig.path === '/departure' || putDeparturePattern.test(evt.detail.requestConfig.path))) {
                if (evt.detail.requestConfig.verb === 'post') {
                    window.formatDepartureListItem('#' + evt.detail.elt.firstElementChild.id);
                } else if (evt.detail.requestConfig.verb === 'put') {
                    window.formatDepartureListItem('#' + evt.detail.elt.id);
                }
            }
        }, { signal });
    }
}
