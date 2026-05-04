/**
 * Removes empty fields from an object.
 * @param {Object} obj - The object from which to remove empty fields.
 * @param {Array} exceptKeys - An array of keys that should not be removed even if their values are empty.
*/
export function removeEmptyFields(obj, exceptKeys = []) {
	if (typeof obj !== 'object' || obj === null || obj === undefined) {
		throw new TypeError('Expected an object');
	}
	return Object.fromEntries(
		Object.entries(obj).filter(([key, value]) => value !== '' && value !== null && value !== undefined || exceptKeys.includes(key))
	);
}
