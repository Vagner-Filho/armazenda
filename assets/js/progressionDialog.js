import { removeEmptyFields } from "./form.js"

function progressionDialogSetup() {
    const dialogEl = document.querySelector("dialog#progressionFormDialog")
    if (!dialogEl) return

    dialogEl.showModal()

    function closeProgressionForm() {
        dialogEl.close()
        dialogEl.remove()
        window.closeProgressionForm = undefined
    }
    dialogEl.addEventListener('close', closeProgressionForm)
    window.closeProgressionForm = closeProgressionForm
    const cancelButton = dialogEl.querySelector('.cancel-btn');
    if (cancelButton) {
        cancelButton.addEventListener('click', closeProgressionForm);
    }

    // Clean empty fields before HTMX submit
    dialogEl.addEventListener('htmx:configRequest', function (evt) {
        evt.detail.parameters = removeEmptyFields(evt.detail.parameters)
    })

    // Close dialog on successful create/update
    document.body.addEventListener('htmx:afterRequest', function handler(evt) {
        if (evt.detail.successful &&
            (evt.detail.requestConfig.verb === 'post' || evt.detail.requestConfig.verb === 'put') &&
            evt.detail.requestConfig.path.startsWith('/progressao')) {
            closeProgressionForm()
            // Refresh the progression table on the config page
            const tableEl = document.getElementById('progression-table')
            if (tableEl) {
                htmx.ajax('GET', '/farm/config/progressions', { target: '#progression-table', swap: 'outerHTML' })
            }
            document.body.removeEventListener('htmx:afterRequest', handler)
        }
    })

    // Add tier row
    const addBtn = dialogEl.querySelector('#add-tier-btn')
    const container = dialogEl.querySelector('#tiers-container')
    if (addBtn && container) {
        addBtn.addEventListener('click', function () {
            const index = container.querySelectorAll('.tier-row').length
            const row = document.createElement('div')
            row.className = 'tier-row flex items-center gap-2'
            row.innerHTML = `
                <div class="flex-1 flex flex-col-reverse">
                    <input type="number" step="0.01" name="tiers[${index}].thresholdHumidity" class="peer bg-white/10 border-white/20 text-white placeholder-white/50" required placeholder="14">
                    <label class="peer-focus:text-white text-white/70 transition-all duration-300 ease-linear px-2 text-xs font-medium uppercase tracking-wider">Umidade (%)</label>
                </div>
                <div class="flex-1 flex flex-col-reverse">
                    <input type="number" step="0.01" name="tiers[${index}].discountValue" class="peer bg-white/10 border-white/20 text-white placeholder-white/50" required placeholder="1.7">
                    <label class="peer-focus:text-white text-white/70 transition-all duration-300 ease-linear px-2 text-xs font-medium uppercase tracking-wider">Desconto</label>
                </div>
                <button type="button" class="remove-tier-btn icon-btn text-red-400/70 hover:text-red-400 mt-4" title="Remover faixa">
                    <iconify-icon icon="ic:baseline-remove-circle" width="20" height="20"></iconify-icon>
                </button>
            `
            container.appendChild(row)
            row.querySelector('.remove-tier-btn').addEventListener('click', function () {
                removeTierRow(row, container)
            })
        })
    }

    // Setup remove buttons for existing rows
    container.querySelectorAll('.remove-tier-btn').forEach(function (btn) {
        btn.addEventListener('click', function () {
            const row = btn.closest('.tier-row')
            removeTierRow(row, container)
        })
    })
}

function removeTierRow(row, container) {
    if (container.querySelectorAll('.tier-row').length <= 1) return
    row.remove()
    // Re-index remaining rows
    container.querySelectorAll('.tier-row').forEach(function (r, i) {
        r.querySelectorAll('input').forEach(function (input) {
            input.name = input.name.replace(/tiers\[\d+\]/, `tiers[${i}]`)
        })
    })
}

export { progressionDialogSetup }
