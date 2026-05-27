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
	function formatCpf() {
		// 1. Get the raw value and remove non-digit characters
		let baseValue = cpfInput.value.replace(/\D/g, '');

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
		cpfInput.value = formattedValue;
	}

	cpfInput.addEventListener('input', formatCpf);
	formatCpf(); // Format initial value if present
	document.body.addEventListener('htmx:configRequest', function(evt) {
		if (evt.detail.parameters['cpf']) {
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
	function formatCnpj() {
		// 1. Get the raw value and remove non-digit characters
		let baseValue = cpfInput.value.replace(/\D/g, '');

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
		cpfInput.value = formattedValue;
	}

	cpfInput.addEventListener('input', formatCnpj);
	formatCnpj(); // Format initial value if present
	document.body.addEventListener('htmx:configRequest', function(evt) {
		if (evt.detail.parameters['cnpj']) {
			evt.detail.parameters['cnpj'] = evt.detail.parameters['cnpj'].replace(/\D/g, '');
		}
	});
}

export function setupCepFormatter(cep_selector) {
	/**
					 * Formats the input value to the Brazilian CEP format (00000-000)
					 * as the user types.
					 * @param {HTMLInputElement} inputElement - The input element to format.
					 */
	const cepInput = document.querySelector(cep_selector)
	if (!cepInput) {
		throw Error("input de cep não encontrado");
	}
	function formatCep() {
		// 1. Get the raw value and remove non-digit characters
		let baseValue = cepInput.value.replace(/\D/g, '');

		// 2. Limit to 8 digits (maximum length of a CEP without formatting)
		baseValue = baseValue.substring(0, 8);

		// 3. Apply the formatting mask
		let formattedValue = '';
		if (baseValue.length > 5) {
			// Format: 00000-000
			formattedValue = baseValue.replace(/(\d{5})(\d{1,3})/, '$1-$2');
		} else {
			// Format: 00000 or less
			formattedValue = baseValue;
		}

		// 4. Update the input field's value
		cepInput.value = formattedValue;
	}

	cepInput.addEventListener('input', formatCep);
	formatCep(); // Format initial value if present
	document.body.addEventListener('htmx:configRequest', function(evt) {
		if (evt.detail.parameters['cep']) {
			evt.detail.parameters['cep'] = evt.detail.parameters['cep'].replace(/\D/g, '');
		}
	});
}

export function setupPhoneFormatter(phone_selector) {
	/**
					 * Formats the input value to the Brazilian phone format ((00) 00000-0000)
					 * as the user types.
					 * @param {HTMLInputElement} inputElement - The input element to format.
					 */
	const phoneInput = document.querySelector(phone_selector)
	if (!phoneInput) {
		throw Error("input de telefone não encontrado");
	}
	function formatPhone() {
		// 1. Get the raw value and remove non-digit characters
		let baseValue = phoneInput.value.replace(/\D/g, '');

		// 2. Limit to 11 digits (maximum length of a Brazilian mobile phone without formatting)
		baseValue = baseValue.substring(0, 11);

		// 3. Apply the formatting mask
		let formattedValue = '';
		if (baseValue.length > 6) {
			// Format: (00) 00000-0000
			formattedValue = baseValue.replace(/(\d{2})(\d{5})(\d{1,4})/, '($1) $2-$3');
		} else if (baseValue.length > 2) {
			// Format: (00) 00000
			formattedValue = baseValue.replace(/(\d{2})(\d{1,5})/, '($1) $2');
		} else {
			// Format: 00 or less
			formattedValue = baseValue;
		}

		// 4. Update the input field's value
		phoneInput.value = formattedValue;
	}

	phoneInput.addEventListener('input', formatPhone);
	formatPhone(); // Format initial value if present
	document.body.addEventListener('htmx:configRequest', function(evt) {
		if (evt.detail.parameters['phoneNumber']) {
			evt.detail.parameters['phoneNumber'] = evt.detail.parameters['phoneNumber'].replace(/\D/g, '');
		}
	});
}
