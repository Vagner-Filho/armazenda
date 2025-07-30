const HUMIDITY_THRESHOLD = 14;

export function getHumidityDiscount() {
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
