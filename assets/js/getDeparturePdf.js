import { formatDateToDisplay } from "./date.js"
import { formatWeight } from "./weight.js"

/**
 * fetches a departure and generates a pdf with the received html departure.
 * if the session cookie is not valid, redirects user to loign
 * @param {HTMLElement} element - The element that contains the departure id as its id
 */
function getDeparturePdf(element) {
	fetch(`departure/pdf/${element.id}`)
		.then((res) => {
			if (res.status == 401) {
				const ev = new CustomEvent("toast", { bubbles: true, detail: { Message: "Sua sessão expirou. Você será redirecionado ao login", Hint: "", Type: 1 } })
				document.body.dispatchEvent(ev)
				setTimeout(() => {
					window.location.href = '/'
				}, 3000)
			} else {
				res.text()
					.then((htmlText) => {
						const parser = new DOMParser();
						const html = parser.parseFromString(htmlText, "text/html")

						const date = html.querySelector('#emission-date')
						date.textContent = formatDateToDisplay(date.textContent)

						const gross = html.querySelector('#grossWeight')
						gross.textContent = formatWeight(gross.textContent)

						const tare = html.querySelector('#tare')
						tare.textContent = formatWeight(tare.textContent)

						const netWeight = html.querySelector('#netWeight')
						netWeight.textContent = formatWeight(netWeight.textContent)
						document.body.append(html.body.firstElementChild)
						function removePdf() {
							const pdf = document.querySelector(`#departure-pdf`)
							if (pdf) {
								pdf.remove()
							}
						}
						window.addEventListener('afterprint', removePdf)
						window.print()
						window.removeEventListener('afterprint', removePdf)
					})
			}
		})
}

window.getDeparturePdf = getDeparturePdf;
