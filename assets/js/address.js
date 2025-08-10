export function setupCepListener(cepFieldId, streetFieldId, neighborhoodFieldId, cityFieldId, stateFieldId) {
    const cepInput = document.getElementById(cepFieldId);
    if (cepInput) {
        cepInput.addEventListener('input', () => {
            const cep = cepInput.value.replace(/\D/g, '');
            if (cep.length === 8) {
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
                    .catch(error => console.error('Error fetching CEP:', error));
            }
        });
    }
}