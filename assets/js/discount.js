import { formatWeight } from "./weight.js";
import { progressionSync } from "./db/progressionSync.js";

const HUMIDITY_THRESHOLD = 14;
const DAMAGE_THRESHOLD = 8;
const IMPURITY_THRESHOLD = 1;

/**
 * Get the current humidity discount based on progressive tiers
 * @param {number} humidity - Humidity value (optional, reads from DOM if not provided)
 * @param {number} gross - Gross weight (optional)
 * @param {number} tare - Tare weight (optional)
 * @param {number} h_discount - Legacy single discount value (deprecated, for backward compatibility)
 * @returns {Promise<number>} The calculated discount
 */
async function getHumidityDiscount(humidity, gross, tare, h_discount) {
	const humidityInput = humidity ? { value: humidity } : document.querySelector('input#humidity');
	const exceedingHumidity = parseFloat(humidityInput.value) - HUMIDITY_THRESHOLD;

	if (!humidityInput || exceedingHumidity <= 0 || isNaN(exceedingHumidity)) {
		if (humidityInput instanceof HTMLInputElement) {
			const label = humidityInput.previousElementSibling;
			if (label instanceof HTMLElement) {
				label.textContent = "";
			}
		}
		updateHumidityTierUI(null, 0);
		return 0;
	}

	const personConfig = sessionStorage.getItem('personConfig');
	const farmConfig = sessionStorage.getItem('farmConfig');

	let discountValue;

	// Check for legacy single discount (backward compatibility)
	if (h_discount) {
		discountValue = parseFloat(h_discount);
	} else if (personConfig || farmConfig) {
		// Get progression for this person/farm
		const person = personConfig ? JSON.parse(personConfig) : {};
		const farm = farmConfig ? JSON.parse(farmConfig) : {};

		try {
			const progression = await progressionSync.getCurrentProgression(
				person.personConfig || person,
				farm.farmConfig || farm
			);

			if (progression) {
				discountValue = progressionSync.getDiscountForHumidity(
					progression,
					parseFloat(humidityInput.value)
				);
			} else {
				// Fallback: try legacy humidityDiscount
				discountValue = person.humidityDiscount || 0;
			}
			// Update UI with tier info
			updateHumidityTierUI(progression, parseFloat(humidityInput.value));
		} catch (error) {
			console.error('[Discount] Failed to get progression:', error);
			// Fallback to legacy
			discountValue = (person && person.humidityDiscount) ? parseFloat(person.humidityDiscount) : 0;
		}
	} else {
		return 0;
	}

	if (discountValue === undefined || discountValue === null || isNaN(discountValue)) {
		return 0;
	}

	const discount = exceedingHumidity * discountValue;

	const netWeightInput = document.querySelector('input#net_weight');
	if (!netWeightInput || !netWeightInput.dataset.raw && !gross && !tare) {
		return 0;
	}

	const rawNetWeight = gross && tare ? gross - tare : parseFloat(netWeightInput.dataset.raw);
	if (isNaN(rawNetWeight) || rawNetWeight <= 0) {
		return 0;
	}

	if (humidityInput instanceof HTMLInputElement) {
		const label = humidityInput.previousElementSibling;
		if (label instanceof HTMLElement) {
			label.textContent = "Peso descontado: " + ((rawNetWeight * discount) / 100).toFixed(3);
		}
	}

	return (rawNetWeight * discount) / 100;
}

/**
 * Update UI with current tier information
 * @param {Object} progression - The progression object
 * @param {number} humidity - Current humidity value
 */
function updateHumidityTierUI(progression, humidity) {
	const tierInfo = progressionSync.getTierDisplayInfo(progression, humidity);
	const tierDisplay = document.getElementById('humidityTierDisplay');

	if (tierDisplay) {
		if (tierInfo.hasTier) {
			tierDisplay.textContent = `Desconto ${tierInfo.tier.discountValue}`;
			tierDisplay.classList.remove('text-gray-400');
			tierDisplay.classList.add('text-green-600');
		} else {
			tierDisplay.textContent = '';
		}
	}

	// Update progression source indicator
	const sourceDisplay = document.getElementById('humidityProgressionSource');
	if (sourceDisplay && tierInfo.progressionName) {
		let sourceText = tierInfo.progressionName;
		if (tierInfo.isDefault) {
			sourceText += ' (Padrão)';
		}
		sourceDisplay.textContent = sourceText;
		return
	}
	if (sourceDisplay) {
		sourceDisplay.textContent = '';
	}
}

function getDamageDiscount(damage, gross, tare) {
	const damageInput = damage ? { value: damage } : document.querySelector('input#damage');
	const exceedingDamage = parseFloat(damageInput.value) - DAMAGE_THRESHOLD;

	if (!damageInput || isNaN(exceedingDamage) || exceedingDamage <= 0) {
		if (damageInput instanceof HTMLInputElement) {
			const label = damageInput.previousElementSibling;
			if (label instanceof HTMLElement) {
				label.textContent = "";
			}
		}
		return 0;
	}

	const netWeightInput = document.querySelector('input#net_weight');
	if (!netWeightInput || !netWeightInput.dataset.raw && !gross && !tare) {
		return 0;
	}

	const rawNetWeight = gross && tare ? gross - tare : parseFloat(netWeightInput.dataset.raw);
	if (isNaN(rawNetWeight) || rawNetWeight <= 0) {
		return 0;
	}

	if (damageInput instanceof HTMLInputElement) {
		const label = damageInput.previousElementSibling;
		if (label instanceof HTMLElement) {
			label.textContent = "Peso descontado: " + ((rawNetWeight * exceedingDamage) / 100).toFixed(3);
		}
	}

	return (rawNetWeight * exceedingDamage) / 100;
}

function getImpurityDiscount(impurity, gross, tare) {
	const impurityInput = impurity ? { value: impurity } : document.querySelector('input#impurity');
	const exceedingImpurity = parseFloat(impurityInput.value) - IMPURITY_THRESHOLD;

	if (!impurityInput || isNaN(exceedingImpurity) || exceedingImpurity <= 0) {
		if (impurityInput instanceof HTMLInputElement) {
			const label = impurityInput.previousElementSibling;
			if (label instanceof HTMLElement) {
				label.textContent = "";
			}
		}
		return 0;
	}

	const netWeightInput = document.querySelector('input#net_weight');
	if (!netWeightInput || !netWeightInput.dataset.raw && !gross && !tare) {
		return 0;
	}

	const rawNetWeight = gross && tare ? gross - tare : parseFloat(netWeightInput.dataset.raw);
	if (isNaN(rawNetWeight) || rawNetWeight <= 0) {
		return 0;
	}

	if (impurityInput instanceof HTMLInputElement) {
		const label = impurityInput.previousElementSibling;
		if (label instanceof HTMLElement) {
			label.textContent = "Peso descontado: " + ((rawNetWeight * exceedingImpurity) / 100).toFixed(3);
		}
	}

	return (rawNetWeight * exceedingImpurity) / 100;
}

function getSoyDiscount(weightAfterQualityDiscount) {
	const personConfig = sessionStorage.getItem('personConfig');
	if (!personConfig) {
		return 0;
	}

	const person = JSON.parse(personConfig);
	const discount = parseFloat(person.entrySoyDiscount);
	if (discount === undefined || discount === null || isNaN(discount)) {
		return 0;
	}

	let rawNetWeight = weightAfterQualityDiscount;
	if (isNaN(rawNetWeight)) {
		rawNetWeight = parseFloat(weightAfterQualityDiscount);
		if (isNaN(rawNetWeight) || rawNetWeight <= 0) {
			return 0;
		}
	}

	return (rawNetWeight * discount) / 100;
}

function getCornDiscount(weightAfterQualityDiscount) {
	const personConfig = sessionStorage.getItem('personConfig');
	if (!personConfig) {
		return 0;
	}

	const person = JSON.parse(personConfig);
	const discount = parseFloat(person.entryCornDiscount);
	if (discount === undefined || discount === null || isNaN(discount)) {
		return 0;
	}

	let rawNetWeight = weightAfterQualityDiscount;
	if (isNaN(rawNetWeight)) {
		rawNetWeight = parseFloat(weightAfterQualityDiscount);
		if (isNaN(rawNetWeight) || rawNetWeight <= 0) {
			return 0;
		}
	}

	return (rawNetWeight * discount) / 100;
}

const productTypeDiscountHandler = {
	1: getCornDiscount,
	2: getSoyDiscount
}

/**
 * Apply all discounts and calculate final net weight
 * @returns {Promise<void>}
 */
export async function applyDiscounts() {
	const hasGrossValue = document.querySelector('input#grossWeight')?.value
	const hasTareValue = document.querySelector('input#tare')?.value
	if (!hasGrossValue || !hasTareValue) {
		return
	}
	const humidityDiscount = await getHumidityDiscount();
	const damageDiscount = getDamageDiscount();
	const impurityDiscount = getImpurityDiscount();

	const totalDiscount = humidityDiscount + damageDiscount + impurityDiscount;

	const netWeightInput = document.querySelector('input#net_weight');
	const rawNetWeightInput = document.querySelector('input#net_weight_raw');
	if (!netWeightInput || !netWeightInput.dataset.raw) {
		return;
	}
	const rawNetWeight = parseFloat(netWeightInput.dataset.raw);
	if (isNaN(rawNetWeight) || rawNetWeight <= 0) {
		return;
	}

	const netWeightAfterQualityDiscount = rawNetWeight - totalDiscount;

	const productType = Number(sessionStorage.getItem("product"));
	const getEntryDiscount = productTypeDiscountHandler[productType]
	const entryProductDiscount = getEntryDiscount ? getEntryDiscount(netWeightAfterQualityDiscount) : 0;

	const finalNetWeight = netWeightAfterQualityDiscount - entryProductDiscount;

	if (finalNetWeight < 0) {
		netWeightInput.value = '0';
		rawNetWeightInput.value = '0';
	} else {
		rawNetWeightInput.value = formatWeight(rawNetWeight.toFixed(2));
		netWeightInput.value = formatWeight(finalNetWeight.toFixed(2));
	}

	const entryTaxField = document.querySelector('input#entryTax')
	entryTaxField.value = formatWeight(entryProductDiscount.toFixed(2));
}
