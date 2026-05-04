/**
 * returns a formatted string based on its length (14 for CNPJ, 11 for CPF).
 * returns N/A if the string is null.
 * returns the original string if it is not a valid CPF or CNPJ.
 * @param {string} document - string representing the document to be formatted: cpf or cnpj
 */
export function formatDocument(document) {
	if (!document) return "N/A";

	document = document.replace(/\D/g, "");
	if (document.length === 11) {
		return document.replace(/(\d{3})(\d{3})(\d{3})(\d{2})/, "$1.$2.$3-$4");
	} else if (document.length === 14) {
		return document.replace(/(\d{2})(\d{3})(\d{3})(\d{4})(\d{2})/, "$1.$2.$3/$4-$5");
	} else {
		return document;
	}
}
