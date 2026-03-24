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

        function closeDepartureForm() {
            dialogEl.close()
            dialogEl.remove()
        }

        dialogEl.addEventListener('close', closeDepartureForm)
        window.closeDepartureForm = closeDepartureForm
        const cancelButton = dialogEl.querySelector('.cancel-btn');
        if (cancelButton) {
            cancelButton.addEventListener('click', closeDepartureForm);
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
            })
            tareInput.addEventListener('input', (e) => {
                tareValue = Number(e.target.value) ?? 0
                netWeightInput.value = grossWeightValue - tareValue
            })
        }

        document.body.addEventListener('htmx:configRequest', function(evt) {
            evt.detail.parameters = removeEmptyFields(evt.detail.parameters);
        });

        document.body.addEventListener('htmx:afterRequest', function(evt) {
            if (evt.detail.successful && evt.detail.requestConfig.path === '/departure' && (evt.detail.requestConfig.verb === 'post' || evt.detail.requestConfig.verb === 'put')) {
                closeDepartureForm()
            }
        });

        document.body.addEventListener('htmx:afterSwap', function(evt) {
            if (evt.detail.successful && evt.detail.requestConfig.path === '/departure' && (evt.detail.requestConfig.verb === 'post' || evt.detail.requestConfig.verb === 'put')) {
                window.formatDepartureListItem('#' + evt.detail.elt.firstElementChild.id);
            }
        });
    }
}
