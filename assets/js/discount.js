const HUMIDITY_THRESHOLD = 14;
const DAMAGE_THRESHOLD = 8;
const IMPURITY_THRESHOLD = 1;

function getHumidityDiscount(humidity, gross, tare, h_discount) {
	const humidityInput = humidity ? { value: humidity } : document.querySelector('input#humidity');
	const exceedingHumidity = parseFloat(humidityInput.value) - HUMIDITY_THRESHOLD;

	if (!humidityInput || exceedingHumidity <= 0) {
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
		const label = humidityInput.nextElementSibling;
		if (label instanceof HTMLLabelElement) {
			label.dataset.discounted = "Peso descontado: " + ((rawNetWeight * discount) / 100).toFixed(2);
		}
	}

	return (rawNetWeight * discount) / 100;
}

function getDamageDiscount(damage, gross, tare) {
	const damageInput = damage ? { value: damage } : document.querySelector('input#damage');
	const exceedingDamage = parseFloat(damageInput.value) - DAMAGE_THRESHOLD;

	if (!damageInput || isNaN(exceedingDamage) || exceedingDamage <= 0) {
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

	return (rawNetWeight * exceedingDamage) / 100;
}

function getImpurityDiscount(impurity, gross, tare) {
	const impurityInput = impurity ? { value: impurity } : document.querySelector('input#impurity');
	const exceedingImpurity = parseFloat(impurityInput.value) - IMPURITY_THRESHOLD;

	if (!impurityInput || isNaN(exceedingImpurity) || exceedingImpurity <= 0) {
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

	return (rawNetWeight * exceedingImpurity) / 100;
}

export function applyDiscounts() {
	const humidityDiscount = getHumidityDiscount();
	const damageDiscount = getDamageDiscount();
	const impurityDiscount = getImpurityDiscount();

	const totalDiscount = humidityDiscount + damageDiscount + impurityDiscount;

	const netWeightInput = document.querySelector('input#netWeight');
	if (!netWeightInput || !netWeightInput.dataset.raw) {
		return;
	}
	const rawNetWeight = parseFloat(netWeightInput.dataset.raw);
	if (isNaN(rawNetWeight) || rawNetWeight <= 0) {
		return;
	}
	const finalNetWeight = rawNetWeight - totalDiscount;
	if (finalNetWeight < 0) {
		netWeightInput.value = '0';
	} else {
		netWeightInput.value = finalNetWeight.toFixed(2);
	}
}
