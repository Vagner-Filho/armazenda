export function setupCepListener(cepFieldId, streetFieldId, neighborhoodFieldId, cityFieldId, stateFieldId) {
    const cepInput = document.getElementById(cepFieldId);
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
                            document.getElementById(streetFieldId).value = data.logradouro;
                            document.getElementById(neighborhoodFieldId).value = data.bairro;
                            document.getElementById(cityFieldId).value = data.localidade;
                            document.getElementById(stateFieldId).value = data.uf;
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
