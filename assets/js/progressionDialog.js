import { removeEmptyFields } from "./form.js"

function progressionDialogSetup() {
    const dialogEl = document.querySelector("dialog#progressionFormDialog")
    if (!dialogEl) return

    dialogEl.showModal()

    const controller = new AbortController()
    const signal = controller.signal

    function closeProgressionForm() {
        controller.abort()
        dialogEl.close()
        dialogEl.remove()
        window.closeProgressionForm = undefined
    }
    dialogEl.addEventListener('close', closeProgressionForm, { signal })
    window.closeProgressionForm = closeProgressionForm
    const cancelButton = dialogEl.querySelector('.cancel-btn');
    if (cancelButton) {
        cancelButton.addEventListener('click', closeProgressionForm, { signal });
    }

    // Clean empty fields before HTMX submit
    dialogEl.addEventListener('htmx:configRequest', function (evt) {
        evt.detail.parameters = removeEmptyFields(evt.detail.parameters)
    }, { signal })

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
    }, { signal })

    // Initialize tier visualization
    initTierVisualization(signal)

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
                updateTierVisualization()
            }, { signal })

            // Update visualization after adding row
            updateTierVisualization()
        }, { signal })
    }

    // Setup remove buttons for existing rows
    container.querySelectorAll('.remove-tier-btn').forEach(function (btn) {
        btn.addEventListener('click', function () {
            const row = btn.closest('.tier-row')
            removeTierRow(row, container)
            updateTierVisualization()
        }, { signal })
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

// Initialize tier visualization
function initTierVisualization(signal) {
    const container = document.getElementById('tiers-container')
    if (!container) return

    // Initial render
    updateTierVisualization()

    // Watch for changes in tier inputs with debounce
    container.addEventListener('input', debounce(updateTierVisualization, 150), { signal })
}

// Collect and sort tiers from inputs
function getTiersFromInputs() {
    const rows = document.querySelectorAll('.tier-row')
    const tiers = []

    rows.forEach(row => {
        const humidityInput = row.querySelector('input[name*="thresholdHumidity"]')
        const discountInput = row.querySelector('input[name*="discountValue"]')

        if (humidityInput && discountInput) {
            const humidity = parseFloat(humidityInput.value)
            const discount = parseFloat(discountInput.value)

            if (!isNaN(humidity) && !isNaN(discount)) {
                tiers.push({ humidity, discount })
            }
        }
    })

    // Sort by humidity ascending
    return tiers.sort((a, b) => a.humidity - b.humidity)
}

// Calculate blue color based on position in 0-100 range
// Whiteish (low humidity) to deep blue (high humidity)
function getBlueColorForPosition(position) {
    // position is 0-100, normalize to 0-1
    const ratio = Math.min(Math.max(position / 100, 0), 1)

    // HSL: Hue=210 (blue)
    // Start: Lightness=96%, Saturation=10% (almost white)
    // End: Lightness=35%, Saturation=85% (deep blue)
    const saturation = 10 + (ratio * 75)   // 10% to 85%
    const lightness = 96 - (ratio * 61)    // 96% to 35%

    return `hsl(210, ${saturation}%, ${lightness}%)`
}

// Render the visualization with fixed 0-100 range
function updateTierVisualization() {
    const tiers = getTiersFromInputs()
    const segmentsContainer = document.querySelector('.progression-segments')
    const discountLabelsContainer = document.querySelector('.progression-discount-labels')
    const thresholdLabelsContainer = document.querySelector('.progression-threshold-labels')

    if (!segmentsContainer || !discountLabelsContainer || !thresholdLabelsContainer) return

    // Clear existing
    segmentsContainer.innerHTML = ''
    discountLabelsContainer.innerHTML = ''
    thresholdLabelsContainer.innerHTML = ''

    if (tiers.length === 0) {
        // No tiers yet - show placeholder
        segmentsContainer.innerHTML = '<div class="flex items-center justify-center h-full text-white/40 text-sm w-full">Adicione faixas para visualizar</div>'
        return
    }

    const firstTier = tiers[0]

    // === SEGMENT 1: 0% to first tier threshold (No discount) ===
    if (firstTier.humidity > 0) {
        const noDiscountWidth = firstTier.humidity
        const noDiscountColor = getBlueColorForPosition(firstTier.humidity / 2) // Middle of the segment

        // Create segment
        const segment = document.createElement('div')
        segment.className = 'progression-segment'
        segment.style.width = noDiscountWidth + '%'
        segment.style.backgroundColor = noDiscountColor
        segmentsContainer.appendChild(segment)

        // Discount label (above bar) - centered in segment
        const discountLabel = document.createElement('div')
        discountLabel.className = 'progression-discount-label'
        discountLabel.style.left = (noDiscountWidth / 2) + '%'
        discountLabel.textContent = 'Sem desc.'
        discountLabelsContainer.appendChild(discountLabel)

        // Threshold label (below bar) - at the boundary
        const thresholdLabel = document.createElement('div')
        thresholdLabel.className = 'progression-threshold-label'
        thresholdLabel.style.left = noDiscountWidth + '%'
        thresholdLabel.textContent = firstTier.humidity + '%'
        thresholdLabelsContainer.appendChild(thresholdLabel)
    }

    // === MIDDLE SEGMENTS: Between tiers ===
    for (let i = 0; i < tiers.length - 1; i++) {
        const currentTier = tiers[i]
        const nextTier = tiers[i + 1]
        const width = nextTier.humidity - currentTier.humidity
        const midPoint = currentTier.humidity + (width / 2)
        const color = getBlueColorForPosition(midPoint)

        // Create segment
        const segment = document.createElement('div')
        segment.className = 'progression-segment'
        segment.style.width = width + '%'
        segment.style.backgroundColor = color
        segmentsContainer.appendChild(segment)

        // Discount label (above bar) - centered in segment
        const discountLabel = document.createElement('div')
        discountLabel.className = 'progression-discount-label'
        discountLabel.style.left = (currentTier.humidity + width / 2) + '%'
        discountLabel.textContent = currentTier.discount + '%'
        discountLabelsContainer.appendChild(discountLabel)

        // Threshold label (below bar) - at the boundary
        const thresholdLabel = document.createElement('div')
        thresholdLabel.className = 'progression-threshold-label'
        thresholdLabel.style.left = nextTier.humidity + '%'
        thresholdLabel.textContent = nextTier.humidity + '%'
        thresholdLabelsContainer.appendChild(thresholdLabel)
    }

    // === FINAL SEGMENT: From last tier to 100% ===
    const lastTier = tiers[tiers.length - 1]
    const finalWidth = 100 - lastTier.humidity
    const finalMidPoint = lastTier.humidity + (finalWidth / 2)
    const finalColor = getBlueColorForPosition(finalMidPoint)

    // Create segment
    const finalSegment = document.createElement('div')
    finalSegment.className = 'progression-segment'
    finalSegment.style.width = finalWidth + '%'
    finalSegment.style.backgroundColor = finalColor
    segmentsContainer.appendChild(finalSegment)

    // Discount label for last tier (above bar) - centered in final segment
    const finalDiscountLabel = document.createElement('div')
    finalDiscountLabel.className = 'progression-discount-label'
    finalDiscountLabel.style.left = (lastTier.humidity + finalWidth / 2) + '%'
    finalDiscountLabel.textContent = lastTier.discount + '%'
    discountLabelsContainer.appendChild(finalDiscountLabel)

    // === FIXED LABELS: 0% and 100% ===
    // 0% label (first)
    const zeroLabel = document.createElement('div')
    zeroLabel.className = 'progression-threshold-label first'
    zeroLabel.textContent = '0%'
    thresholdLabelsContainer.appendChild(zeroLabel)

    // 100% label (last)
    const hundredLabel = document.createElement('div')
    hundredLabel.className = 'progression-threshold-label last'
    hundredLabel.textContent = '100%'
    thresholdLabelsContainer.appendChild(hundredLabel)
}

// Utility: debounce function
function debounce(func, wait) {
    let timeout
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout)
            func(...args)
        }
        clearTimeout(timeout)
        timeout = setTimeout(later, wait)
    }
}

export { progressionDialogSetup }
