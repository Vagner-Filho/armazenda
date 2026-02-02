import { formatWeight } from "./weight.js";

const HUMIDITY_THRESHOLD = 14;
const DAMAGE_THRESHOLD = 8;
const IMPURITY_THRESHOLD = 1;

function getHumidityDiscount(humidity, gross, tare, h_discount) {
	const humidityInput = humidity ? { value: humidity } : document.querySelector('input#humidity');
	const exceedingHumidity = parseFloat(humidityInput.value) - HUMIDITY_THRESHOLD;

	if (!humidityInput || exceedingHumidity <= 0) {
		if (humidityInput instanceof HTMLInputElement) {
			const label = humidityInput.previousElementSibling;
			if (label instanceof HTMLElement) {
				label.textContent = "";
			}
		}
		return 0;
	}

	const personConfig = sessionStorage.getItem('personConfig');
	if (!personConfig && !h_discount) {
		return 0;
	}

	const person = h_discount ? { humidityDiscount: h_discount } : JSON.parse(personConfig);
	const discount = exceedingHumidity * parseFloat(person.humidityDiscount);
	if (discount === undefined || discount === null || isNaN(discount)) {
		return 0;
	}

	const netWeightInput = document.querySelector('input#netWeight');
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
			label.textContent = "Peso descontado: " + ((rawNetWeight * discount) / 100).toFixed(2);
		}
	}

	return (rawNetWeight * discount) / 100;
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

	const netWeightInput = document.querySelector('input#netWeight');
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
			label.textContent = "Peso descontado: " + ((rawNetWeight * exceedingDamage) / 100).toFixed(2);
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

	const netWeightInput = document.querySelector('input#netWeight');
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
			label.textContent = "Peso descontado: " + ((rawNetWeight * exceedingImpurity) / 100).toFixed(2);
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

export function applyDiscounts() {
	const hasGrossValue = document.querySelector('input#grossWeight')?.value
	const hasTareValue = document.querySelector('input#tare')?.value
	if (!hasGrossValue || !hasTareValue) {
		return
	}
	const humidityDiscount = getHumidityDiscount();
	const damageDiscount = getDamageDiscount();
	const impurityDiscount = getImpurityDiscount();

	const totalDiscount = humidityDiscount + damageDiscount + impurityDiscount;

	const netWeightInput = document.querySelector('input#netWeight');
	const rawNetWeightInput = document.querySelector('input#netWeightRaw');
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
