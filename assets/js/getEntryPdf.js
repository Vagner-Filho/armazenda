/**
 * fetches an entry and triggers event with it as value.
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
						const pdfOptions = {
							margin: 4,
							filename: `romaneio_entrada_${element.id}`,
						}
						window.html2pdf()
							.set(pdfOptions)
							.from(htmlText)
							.save()
					})
			}
		})
}

window.getEntryPdf = getEntryPdf;
