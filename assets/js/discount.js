const HUMIDITY_THRESHOLD = 14;
const DAMAGE_THRESHOLD = 8;
const IMPURITY_THRESHOLD = 1;

function getHumidityDiscount() {
	const humidityInput = document.querySelector('input#humidity');
	const exceedingHumidity = parseFloat(humidityInput.value) - HUMIDITY_THRESHOLD;

	if (!humidityInput || exceedingHumidity <= 0) {
		return 0;
	}

	const personConfig = sessionStorage.getItem('personConfig');
	if (!personConfig) {
		return 0;
	}

	const person = JSON.parse(personConfig);
	const discount = exceedingHumidity * parseFloat(person.humidityDiscount);
	if (discount === undefined || discount === null || isNaN(discount)) {
		return 0;
	}

	const netWeightInput = document.querySelector('#netWeight');
	if (!netWeightInput || !netWeightInput.dataset.raw) {
		return 0;
	}

	const rawNetWeight = parseFloat(netWeightInput.dataset.raw);
	if (isNaN(rawNetWeight) || rawNetWeight <= 0) {
		return 0;
	}

	return (rawNetWeight * discount) / 100;
}

function getDamageDiscount() {
	const damageInput = document.querySelector('input#damage');
	const exceedingDamage = parseFloat(damageInput.value) - DAMAGE_THRESHOLD;

	if (!damageInput || isNaN(exceedingDamage) || exceedingDamage <= 0) {
		return 0;
	}

	const netWeightInput = document.querySelector('#netWeight');
	if (!netWeightInput || !netWeightInput.dataset.raw) {
		return 0;
	}

	const rawNetWeight = parseFloat(netWeightInput.dataset.raw);
	if (isNaN(rawNetWeight) || rawNetWeight <= 0) {
		return 0;
	}

	return (rawNetWeight * exceedingDamage) / 100;
}

function getImpurityDiscount() {
	const impurityInput = document.querySelector('input#impurity');
	const exceedingImpurity = parseFloat(impurityInput.value) - IMPURITY_THRESHOLD;

	if (!impurityInput || isNaN(exceedingImpurity) || exceedingImpurity <= 0) {
		return 0;
	}

	const netWeightInput = document.querySelector('#netWeight');
	if (!netWeightInput || !netWeightInput.dataset.raw) {
		return 0;
	}

	const rawNetWeight = parseFloat(netWeightInput.dataset.raw);
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
