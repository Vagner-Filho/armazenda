export function closeModal(selector) {
	const dialogEl = document.querySelector(selector)
	if (dialogEl) {
		dialogEl.close()
		dialogEl.remove()
	}
}

export function setupCpfFormatter(cpf_selector) {
	/**
					 * Formats the input value to the Brazilian CPF format (XXX.XXX.XXX-XX)
					 * as the user types.
					 * @param {HTMLInputElement} inputElement - The input element to format.
					 */
	const cpfInput = document.querySelector(cpf_selector)
	if (!cpfInput) {
		throw Error("input de cpf não encontrado");
	}
	cpfInput.addEventListener('input', function handleCpfInput() {
		// 1. Get the raw value and remove non-digit characters
		let baseValue = this.value.replace(/\D/g, '');

		// 2. Limit to 11 digits (maximum length of a CPF without formatting)
		baseValue = baseValue.substring(0, 11);

		// 3. Apply the formatting mask
		let formattedValue = '';
		if (baseValue.length > 9) {
			// Format: XXX.XXX.XXX-XX
			formattedValue = baseValue.replace(/(\d{3})(\d{3})(\d{3})(\d{2})/, '$1.$2.$3-$4');
		} else if (baseValue.length > 6) {
			// Format: XXX.XXX.XXX
			formattedValue = baseValue.replace(/(\d{3})(\d{3})(\d{3})/, '$1.$2.$3');
		} else if (baseValue.length > 3) {
			// Format: XXX.XXX
			formattedValue = baseValue.replace(/(\d{3})(\d{3})/, '$1.$2');
		} else {
			// Format: XXX
			formattedValue = baseValue;
		}

		// 4. Update the input field's value
		this.value = formattedValue;
	})
	document.body.addEventListener('htmx:configRequest', function(evt) {
		if (evt.detail.parameters.has('cpf')) {
			evt.detail.parameters['cpf'] = evt.detail.parameters['cpf'].replace(/\D/g, '');
		}
	});
}

export function setupCnpjFormatter(cpf_selector) {
	/**
					 * Formats the input value to the Brazilian CNPJ format (XX.XXX.XXX/XXXX-XX)
					 * as the user types.
					 * @param {HTMLInputElement} inputElement - The input element to format.
					 */
	const cpfInput = document.querySelector(cpf_selector)
	if (!cpfInput) {
		throw Error("input de cnpj não encontrado");
	}
	cpfInput.addEventListener('input', function handleCpfInput() {
		// 1. Get the raw value and remove non-digit characters
		let baseValue = this.value.replace(/\D/g, '');

		// 2. Limit to 11 digits (maximum length of a CNPJ without formatting)
		baseValue = baseValue.substring(0, 14);

		// 3. Apply the formatting mask
		let formattedValue = '';
		if (baseValue.length > 12) {
			// Format: XXX.XXX.XXX-XX
			formattedValue = baseValue.replace(/(\d{2})(\d{3})(\d{3})(\d{4})(\d{2})/, '$1.$2.$3/$4-$5');
		} else if (baseValue.length > 6) {
			// Format: XXX.XXX.XXX
			formattedValue = baseValue.replace(/(\d{2})(\d{3})(\d{3})/, '$1.$2.$3');
		} else if (baseValue.length > 3) {
			// Format: XXX.XXX
			formattedValue = baseValue.replace(/(\d{2})(\d{3})/, '$1.$2');
		} else {
			// Format: XXX
			formattedValue = baseValue;
		}

		// 4. Update the input field's value
		this.value = formattedValue;
	})
	document.body.addEventListener('htmx:configRequest', function(evt) {
		if (evt.detail.parameters.has('cnpj')) {
			evt.detail.parameters['cnpj'] = evt.detail.parameters['cnpj'].replace(/\D/g, '');
		}
	});
}
