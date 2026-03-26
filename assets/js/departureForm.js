import { formatDateToInput } from "./date.js"
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

        let grossWeightValue = Number(grossWeightInput ? grossWeightInput.value : 0)
        let tareValue = Number(tareInput ? tareInput.value : 0)

        const netWeightInput = dialogEl.querySelector('input#netWeight')
        if (netWeightInput) {
            grossWeightInput.addEventListener('input', (e) => {
                grossWeightValue = Number(e.target.value) ?? 0
                netWeightInput.value = grossWeightValue - tareValue
            }, { signal })
            tareInput.addEventListener('input', (e) => {
                tareValue = Number(e.target.value) ?? 0
                netWeightInput.value = grossWeightValue - tareValue
            }, { signal })
        }

        document.body.addEventListener('htmx:configRequest', function(evt) {
            evt.detail.parameters = removeEmptyFields(evt.detail.parameters);
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