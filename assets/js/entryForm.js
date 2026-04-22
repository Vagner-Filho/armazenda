import { formatDateToInput } from "./date.js"
import { applyDiscounts } from "./discount.js"
import { removeEmptyFields } from "./form.js"
import { progressionSync } from "./db/progressionSync.js"

/**
 * Wrapper for async applyDiscounts
 * @returns {Promise<void>}
 */
async function applyDiscountsAsync() {
    await applyDiscounts();
}

export async function setupEntryForm(payload, entry_id) {
    const controller = new AbortController()
    const signal = controller.signal

    const dateVal = formatDateToInput(payload && typeof payload === "object" ? payload.arrivalDate : null)

    const dialogEl = document.querySelector("#addEntryDialog")
    if (dialogEl) {
        dialogEl.showModal()
        const arrivalDateEl = dialogEl.querySelector("input#arrival-date")
        if (arrivalDateEl) {
            arrivalDateEl.setAttribute("value", dateVal)
        }
        const cropSelector = dialogEl.querySelector("select#crop-selector")
        const lastUsedCrop = localStorage.getItem("last_used_crop");
        if (cropSelector && lastUsedCrop && !entry_id) {
            for (const option of cropSelector.options) {
                if (option.value === lastUsedCrop) {
                    option.selected = true;
                    break;
                }
            }
        }

        function closeEntryFormDialog() {
            controller.abort()
            dialogEl.close()
            dialogEl.remove()
            window.closeEntryFormDialog = undefined
        }
        dialogEl.addEventListener('close', closeEntryFormDialog, { signal })
        window.closeEntryFormDialog = closeEntryFormDialog
        const cancelButton = dialogEl.querySelector('.cancel-btn');
        if (cancelButton) {
            cancelButton.addEventListener('click', closeEntryFormDialog, { signal });
        }

        const grossWeightInput = dialogEl.querySelector('input#grossWeight')
        const tareInput = dialogEl.querySelector('input#tare')
        const humidityInput = dialogEl.querySelector('input#humidity')
        const damageInput = dialogEl.querySelector('input#damage')
        const impurityInput = dialogEl.querySelector('input#impurity')

        if (humidityInput) {
            humidityInput.addEventListener('input', applyDiscountsAsync, { signal })
        }
        if (damageInput) {
            damageInput.addEventListener('input', applyDiscountsAsync, { signal })
        }
        if (impurityInput) {
            impurityInput.addEventListener('input', applyDiscountsAsync, { signal })
        }
        /*if (grossWeightInput) {
            grossWeightInput.addEventListener('input', applyDiscountsAsync, { signal })
        }
        if (tareInput) {
            tareInput.addEventListener('input', applyDiscountsAsync, { signal })
        }*/

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
            }, { signal })
            tareInput.addEventListener('input', (e) => {
                tareValue = Number(e.target.value) ?? 0

                netWeightInput.value = grossWeightValue - tareValue
                rawNetWeightInput.value = grossWeightValue - tareValue
                netWeightInput.dataset.raw = grossWeightValue - tareValue
                applyDiscountsAsync()
            }, { signal })
        }
        document.body.addEventListener('htmx:configRequest', function(evt) {
            evt.detail.parameters = removeEmptyFields(evt.detail.parameters);
            const crop = evt.detail.parameters.crop;
            if (!isNaN(crop)) {
                localStorage.setItem("last_used_crop", crop);
            } else {
                localStorage.removeItem("last_used_crop");
            }
        }, { signal });
        document.body.addEventListener('htmx:afterRequest', function(evt) {
            if (evt.detail.successful && evt.detail.requestConfig.path === '/entry' && (evt.detail.requestConfig.verb === 'post' || evt.detail.requestConfig.verb === 'put')) {
                closeEntryFormDialog()
            }
        }, { signal });

        const putEntryPattern = new RegExp(/entry\/[0-9]+$/gm);
        document.body.addEventListener('htmx:afterSwap', function(evt) {
            if (evt.detail.successful && (evt.detail.requestConfig.path === '/entry' || putEntryPattern.test(evt.detail.requestConfig.path))) {
                if (evt.detail.requestConfig.verb === 'post') {
                    window.formatEntryListItem('#' + evt.detail.elt.firstElementChild.id);
                } else if (evt.detail.requestConfig.verb === 'put') {
                    window.formatEntryListItem('#' + evt.detail.elt.id);
                }
            }
        }, { signal });

        // Crop selector
        const cropInput = dialogEl.querySelector('select#crop-selector')
        if (cropInput) {
            cropInput.addEventListener('change', function setProduct(ev) {
                const selectedOption = ev.target.options[ev.target.selectedIndex]
                sessionStorage.setItem('product', selectedOption.dataset.productId)
            }, { signal })
            const selectedOption = cropInput.options[cropInput.selectedIndex]
            sessionStorage.setItem('product', selectedOption.dataset.productId)
        }

        // Origin selector
        const selector = dialogEl.querySelector('#origin-selector');
        if (selector) {
            selector.addEventListener('change', function(evt) {
                setPersonConfig(evt.target);
            }, { signal });
            const option = selector.querySelector('option[selected]')

            if (option && option.dataset.humidityProgressionId) {
                // Fetch progression and get first tier threshold
                let humidityThreshold = 14; // Default
                const progressionId = option.dataset.humidityProgressionId;
                if (progressionId) {
                    try {
                        const progression = await progressionSync.getProgression(parseInt(progressionId));
                        if (progression && progression.tiers && progression.tiers.length > 0) {
                            // Get the minimum threshold from all tiers
                            const thresholds = progression.tiers.map(t => parseFloat(t.thresholdHumidity));
                            humidityThreshold = Math.min(...thresholds);
                        }
                    } catch (error) {
                        console.error('[EntryForm] Failed to get progression threshold:', error);
                    }
                }

                sessionStorage.setItem('personConfig', JSON.stringify({
                    humidityProgressionId: progressionId,
                    entrySoyDiscount: option.dataset.entrySoyDiscount,
                    entryCornDiscount: option.dataset.entryCornDiscount,
                    humidityThreshold: humidityThreshold
                }));
            }
        }

        async function setPersonConfig(el) {
            const selectedOption = el.options[el.selectedIndex];
            const humidityProgressionId = selectedOption.getAttribute('data-humidity-progression-id');
            const entrySoyDiscount = selectedOption.getAttribute('data-entry-soy-discount');
            const entryCornDiscount = selectedOption.getAttribute('data-entry-corn-discount');
            //const personId = selectedOption.value;
            //const personName = selectedOption.textContent.trim();

            // Fetch progression and get first tier threshold
            let humidityThreshold = 14; // Default
            if (humidityProgressionId) {
                try {
                    const progression = await progressionSync.getProgression(parseInt(humidityProgressionId));
                    if (progression && progression.tiers && progression.tiers.length > 0) {
                        // Get the minimum threshold from all tiers
                        const thresholds = progression.tiers.map(t => parseFloat(t.thresholdHumidity));
                        humidityThreshold = Math.min(...thresholds);
                    }
                } catch (error) {
                    console.error('[EntryForm] Failed to get progression threshold:', error);
                }
            }

            sessionStorage.setItem('personConfig', JSON.stringify({
                humidityProgressionId: humidityProgressionId,
                entrySoyDiscount: entrySoyDiscount,
                entryCornDiscount: entryCornDiscount,
                humidityThreshold: humidityThreshold
            }));

            await applyDiscounts();
        }
    }
}
