export function setupCepListener(formId, cepFieldId, streetFieldId, neighborhoodFieldId, cityFieldId, stateFieldId) {
    const parentForm = document.getElementById(formId);
    if (!parentForm) return;

    const cepInput = parentForm.querySelector('#' + cepFieldId);
    if (cepInput) {
        cepInput.addEventListener('input', () => {
            const cep = cepInput.value.replace(/\D/g, '');
            if (cep.length === 8) {
                const fieldset = cepInput.closest('fieldset')
                if (fieldset) {
                    fieldset.classList.toggle('cep-data-loading')
                }
                cepInput.setAttribute('disabled', true)
                fetch(`https://viacep.com.br/ws/${cep}/json/`)
                    .then(response => response.json())
                    .then(data => {
                        if (!data.erro) {
                            parentForm.querySelector('#' + streetFieldId).value = data.logradouro;
                            parentForm.querySelector('#' + neighborhoodFieldId).value = data.bairro;
                            parentForm.querySelector('#' + cityFieldId).value = data.localidade;
                            parentForm.querySelector('#' + stateFieldId).value = data.uf;
                        }
                    })
                    .catch(error => console.error('Error fetching CEP:', error))
                    .finally(() => {
                        if (fieldset) {
                            fieldset.classList.toggle('cep-data-loading')
                        }
                        cepInput.removeAttribute('disabled')
                    })
            }
        });
    }
}
