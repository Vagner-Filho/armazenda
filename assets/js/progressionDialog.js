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
    dialogEl.addEventListener('htmx:configRequest', function(evt) {
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
        addBtn.addEventListener('click', function() {
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
            row.querySelector('.remove-tier-btn').addEventListener('click', function() {
                removeTierRow(row, container)
                updateTierVisualization()
            }, { signal })

            // Update visualization after adding row
            updateTierVisualization()
        }, { signal })
    }

    // Setup remove buttons for existing rows
    container.querySelectorAll('.remove-tier-btn').forEach(function(btn) {
        btn.addEventListener('click', function() {
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
    container.querySelectorAll('.tier-row').forEach(function(r, i) {
        r.querySelectorAll('input').forEach(function(input) {
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

// Minimum pixel widths for segments
const MIN_FIRST_SEGMENT_WIDTH = 80;  // For "Sem Desconto" text
const MIN_TIER_SEGMENT_WIDTH = 25;   // Minimum between tiers
const MIN_LAST_SEGMENT_WIDTH = 25;   // From last tier to 100%

// Base scale: pixels per 1% humidity difference
const BASE_PIXELS_PER_PERCENT = 3;

// Calculate pixel positions for all markers
// Returns { positions: [{humidity, pixelPosition}], totalWidth }
function calculateMarkerPositions(tiers) {
    // Always include 0 and 100 in markers
    if (tiers.length === 0) {
        // No tiers: just 0% to 100% with minimum width for "Sem Desconto"
        const segmentWidth = Math.max(100 * BASE_PIXELS_PER_PERCENT, MIN_FIRST_SEGMENT_WIDTH);
        return {
            positions: [
                { humidity: 0, pixelPosition: 0 },
                { humidity: 100, pixelPosition: segmentWidth }
            ],
            totalWidth: segmentWidth
        };
    }
    
    // Create markers: [0, tier1, tier2, ..., tierN, 100]
    const markers = [0, ...tiers.map(t => t.humidity), 100];
    
    // Calculate segment widths (humidity differences)
    const segments = [];
    for (let i = 0; i < markers.length - 1; i++) {
        segments.push({
            startHumidity: markers[i],
            endHumidity: markers[i + 1],
            humidityDiff: markers[i + 1] - markers[i]
        });
    }
    
    // Calculate base pixel widths (proportional)
    let segmentWidths = segments.map((seg, index) => {
        const baseWidth = seg.humidityDiff * BASE_PIXELS_PER_PERCENT;
        
        // Apply minimum widths
        if (index === 0) {
            // First segment (0 to first tier)
            return Math.max(baseWidth, MIN_FIRST_SEGMENT_WIDTH);
        } else if (index === segments.length - 1) {
            // Last segment (last tier to 100)
            return Math.max(baseWidth, MIN_LAST_SEGMENT_WIDTH);
        } else {
            // Middle segments (tier to tier)
            return Math.max(baseWidth, MIN_TIER_SEGMENT_WIDTH);
        }
    });
    
    // Calculate cumulative positions
    const positions = [];
    let currentPixel = 0;
    
    for (let i = 0; i < markers.length; i++) {
        positions.push({
            humidity: markers[i],
            pixelPosition: currentPixel
        });
        
        if (i < segmentWidths.length) {
            currentPixel += segmentWidths[i];
        }
    }
    
    const totalWidth = currentPixel;
    
    return { positions, totalWidth };
}

// Render the visualization with dynamic pixel-based layout
function updateTierVisualization() {
    const tiers = getTiersFromInputs()
    const segmentsContainer = document.querySelector('.progression-segments')
    const discountLabelsContainer = document.querySelector('.progression-discount-labels')
    const thresholdLabelsContainer = document.querySelector('.progression-threshold-labels')
    const leaderLinesContainer = document.querySelector('.progression-leader-lines')
    const barContainer = document.querySelector('.progression-bar')

    if (!segmentsContainer || !discountLabelsContainer || !thresholdLabelsContainer || !leaderLinesContainer || !barContainer) return

    // Clear existing
    segmentsContainer.innerHTML = ''
    discountLabelsContainer.innerHTML = ''
    thresholdLabelsContainer.innerHTML = ''
    leaderLinesContainer.innerHTML = ''

    // Set bar width to 100% of container
    barContainer.style.width = '100%';
    
    // Calculate pixel positions (empty array if no tiers)
    const { positions, totalWidth } = calculateMarkerPositions(tiers);
    
    // Helper to convert pixel position to percentage
    const getPercentPosition = (humidity) => {
        const pos = positions.find(p => p.humidity === humidity);
        if (!pos) return 0;
        return (pos.pixelPosition / totalWidth) * 100;
    };
    
    let hasStretched = false;

    // === ALWAYS SHOW: 0% to first tier (or 100% if no tiers) ===
    const firstTier = tiers.length > 0 ? tiers[0] : null;
    const firstTierPercent = firstTier ? getPercentPosition(firstTier.humidity) : 100;
    const zeroPercent = 0;
    const firstSegmentWidth = firstTierPercent - zeroPercent;

    // Create segment (transparent, shows gradient underneath)
    const segment = document.createElement('div')
    segment.className = 'progression-segment'
    segment.style.width = firstSegmentWidth + '%'
    if (firstTier) {
        segment.setAttribute('title', `0% → ${firstTier.humidity}%: Sem desconto`)
    } else {
        segment.setAttribute('title', `0% → 100%: Sem desconto`)
    }
    segmentsContainer.appendChild(segment)

    // Discount label (above bar) - centered in segment - ALWAYS SHOW
    const discountLabel = document.createElement('div')
    discountLabel.className = 'progression-discount-label'
    discountLabel.style.left = (zeroPercent + firstSegmentWidth / 2) + '%'
    discountLabel.textContent = 'Sem desc.'
    discountLabelsContainer.appendChild(discountLabel)

    // Leader line at first tier boundary (if exists)
    if (firstTier) {
        const leaderLine = document.createElement('div')
        leaderLine.className = 'progression-leader-line'
        leaderLine.style.left = firstTierPercent + '%'
        leaderLinesContainer.appendChild(leaderLine)

        // Threshold label with link line inside
        const thresholdLabel = createThresholdLabel(
            firstTier.humidity,
            'default',
            firstTierPercent
        )
        thresholdLabelsContainer.appendChild(thresholdLabel)
    }

    // === MIDDLE SEGMENTS: Between tiers ===
    for (let i = 0; i < tiers.length - 1; i++) {
        const currentTier = tiers[i]
        const nextTier = tiers[i + 1]
        const currentPercent = getPercentPosition(currentTier.humidity);
        const nextPercent = getPercentPosition(nextTier.humidity);
        const segmentWidth = nextPercent - currentPercent;

        // Create segment (transparent, shows gradient underneath)
        const segment = document.createElement('div')
        segment.className = 'progression-segment'
        segment.style.width = segmentWidth + '%'
        segment.setAttribute('title', `${currentTier.humidity}% → ${nextTier.humidity}%: ${currentTier.discount}% desconto`)
        segmentsContainer.appendChild(segment)

        // Discount label (above bar) - centered in segment
        const discountLabel = document.createElement('div')
        discountLabel.className = 'progression-discount-label'
        discountLabel.style.left = (currentPercent + segmentWidth / 2) + '%'
        discountLabel.textContent = currentTier.discount + '%'
        discountLabelsContainer.appendChild(discountLabel)

        // Leader line at the boundary
        const leaderLine = document.createElement('div')
        leaderLine.className = 'progression-leader-line'
        leaderLine.style.left = nextPercent + '%'
        leaderLinesContainer.appendChild(leaderLine)

        // Threshold label with link line inside
        const thresholdLabel = createThresholdLabel(
            nextTier.humidity,
            'default',
            nextPercent
        )
        thresholdLabelsContainer.appendChild(thresholdLabel)
    }

    // === FINAL SEGMENT: From last tier to 100% (or full bar if no tiers) ===
    const lastTier = tiers.length > 0 ? tiers[tiers.length - 1] : null;
    const lastTierPercent = lastTier ? getPercentPosition(lastTier.humidity) : 0;
    const hundredPercent = 100;
    const finalSegmentWidth = hundredPercent - lastTierPercent;

    // Create segment (transparent, shows gradient underneath)
    const finalSegment = document.createElement('div')
    finalSegment.className = 'progression-segment'
    finalSegment.style.width = finalSegmentWidth + '%'
    if (lastTier) {
        finalSegment.setAttribute('title', `${lastTier.humidity}% → 100%: ${lastTier.discount}% desconto`)
    } else {
        finalSegment.setAttribute('title', `0% → 100%: Sem desconto`)
    }
    segmentsContainer.appendChild(finalSegment)

    // Discount label for last tier (above bar) - only if tiers exist
    if (lastTier) {
        const finalDiscountLabel = document.createElement('div')
        finalDiscountLabel.className = 'progression-discount-label'
        finalDiscountLabel.style.left = (lastTierPercent + finalSegmentWidth / 2) + '%'
        finalDiscountLabel.textContent = lastTier.discount + '%'
        discountLabelsContainer.appendChild(finalDiscountLabel)
    }

    // === FIXED LABELS: 0% and 100% - ALWAYS SHOW ===
    // 0% label (first) - horizontal, no leader line needed
    const zeroLabel = document.createElement('div')
    zeroLabel.className = 'progression-threshold-label first'
    zeroLabel.style.left = '0%'
    zeroLabel.textContent = '0%'
    thresholdLabelsContainer.appendChild(zeroLabel)

    // 100% label (last) - horizontal, add leader line at the end
    const lastLeaderLine = document.createElement('div')
    lastLeaderLine.className = 'progression-leader-line'
    lastLeaderLine.style.left = '100%'
    leaderLinesContainer.appendChild(lastLeaderLine)

    const hundredLabel = document.createElement('div')
    hundredLabel.className = 'progression-threshold-label last'
    hundredLabel.style.left = '96%'  // Slight offset to keep visible
    hundredLabel.textContent = '100%'
    thresholdLabelsContainer.appendChild(hundredLabel)

    // Update explanation text
    updateExplanationText(tiers)
}

// Update the explanation text based on current tiers
function updateExplanationText(tiers) {
    const explanationContainer = document.getElementById('progression-explanation')
    if (!explanationContainer) return

    // Hide if no tiers
    if (tiers.length === 0) {
        explanationContainer.classList.add('hidden')
        return
    }

    // Show container
    explanationContainer.classList.remove('hidden')

    // Build explanation text
    let text = ''
    const numTiers = tiers.length

    if (numTiers === 1) {
        const tier = tiers[0]
        text = `Esta progressão possui 1 faixa. Umidade até ${tier.humidity}% não tem desconto. A partir de ${tier.humidity}% aplica-se ${tier.discount}% de desconto.`
    } else {
        text = `Esta progressão possui ${numTiers} faixas. `
        
        // First part - no discount
        text += `Umidade até ${tiers[0].humidity}% não tem desconto. `
        
        // Middle parts - tier ranges
        for (let i = 0; i < tiers.length - 1; i++) {
            const currentTier = tiers[i]
            const nextTier = tiers[i + 1]
            text += `Entre ${currentTier.humidity}% e ${nextTier.humidity}% aplica-se ${currentTier.discount}% de desconto. `
        }
        
        // Last part - final tier
        const lastTier = tiers[tiers.length - 1]
        text += `A partir de ${lastTier.humidity}% aplica-se ${lastTier.discount}% de desconto.`
    }

    explanationContainer.textContent = text
}

// Create a threshold label element with link line inside
function createThresholdLabel(humidity, strategy, percentPosition) {
    const label = document.createElement('div')
    label.className = 'progression-threshold-label'
    
    // Position using percentage
    label.style.left = percentPosition + '%'
    
    // Create link line inside the label
    const linkLine = document.createElement('div')
    linkLine.className = 'label-link-line'
    label.appendChild(linkLine)
    
    // Create text span
    const textSpan = document.createElement('span')
    textSpan.className = 'label-text'
    textSpan.textContent = humidity + '%'
    label.appendChild(textSpan)
    
    return label
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
