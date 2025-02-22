const isTimeZeroValue = (date) => date === '0001-01-01T00:00:00Z' || !date
export function formatDateToInput(date, dateOnly) {
	let dateVal = date
	const isZeroValue = isTimeZeroValue(dateVal)

	try {
		dateVal = isZeroValue ? new Date() : new Date(dateVal)
	} catch {
		dateVal = new Date()
	}

	const [YYYY, MM, DD] = [
		isZeroValue ? dateVal.getFullYear() : dateVal.getUTCFullYear(),
		(isZeroValue ? dateVal.getMonth() : dateVal.getUTCMonth()) + 1,
		isZeroValue ? dateVal.getDate() : dateVal.getUTCDate(),
	]

	if (dateOnly) {
		return `${YYYY}-${MM < 10 ? '0' + MM : MM}-${DD < 10 ? '0' + DD : DD}`
	}

	const [HH, mm] = [
		isZeroValue ? dateVal.getHours() : dateVal.getUTCHours(),
		isZeroValue ? dateVal.getMinutes() : dateVal.getUTCMinutes()
	]

	return `${YYYY}-${MM < 10 ? '0' + MM : MM}-${DD < 10 ? '0' + DD : DD}T${HH < 10 ? '0' + HH : HH}:${mm < 10 ? '0' + mm : mm}`
}

export function formatDateToDisplay(date) {
	const isZeroValue = isTimeZeroValue(date)
	return Intl.DateTimeFormat(
		'pt-BR', {
		day: '2-digit',
		month: 'short',
		year: 'numeric',
		hour: '2-digit',
		minute: '2-digit',
		timeZone: isZeroValue ? undefined : 'UTC'
	})
		.format(isZeroValue ? new Date() : new Date(date))
}

export function setTodayDatetimeInput(selector, dateOnly) {
	const itemRow = document.querySelector(selector)
	if (itemRow) {
		itemRow.setAttribute("value", formatDateToInput(null, dateOnly))
	}
}
