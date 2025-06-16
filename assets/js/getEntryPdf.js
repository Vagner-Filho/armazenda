/**
 * fetches an entry and generates a pdf with the received entry html.
 * if the session cookie is not valid, redirects user to loign
 * @param {HTMLElement} element - The element that contains the entry id as its id
 */
function getEntryPdf(element) {
	fetch(`entry/pdf/${element.id}`)
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
						document.body.append(html.body.firstElementChild)
						function removePdf() {
							const pdf = document.querySelector(`#entry-pdf`)
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

window.getEntryPdf = getEntryPdf;
