document.addEventListener("DOMContentLoaded", function() {
	["chip", "a.chip"].forEach((selec) => {
		const chips = document.querySelectorAll(selec);
		chips.forEach((chip) => {
			let fullUrl = '';
			let title = '';
			if (selec == "chip") {
				fullUrl = chip.textContent.trim();
				title = chip.title;
			} else {
				fullUrl = chip.href;
				title = chip.textContent.trim();
			}
			try {
				const url = new URL(fullUrl);
				const domain = url.hostname.replace("www.", "");
				if (title == "") {
					chip.textContent = domain;
				} else {
					chip.textContent = title;
				}
				if (selec == "chip") {
					chip.addEventListener("click", () => {
						window.open(fullUrl, "_self");
					});
				}

				const faviconUrl =
					`https://www.google.com/s2/favicons?sz=64&domain=${domain}`;

				const img = new Image();
				img.onload = function() {
					const style = document.createElement("style");
					if (selec == "chip") {
						style.textContent = `
                    chip[data-domain="${domain}"]::before {
                        background-image: url(${faviconUrl});
                        background-color: transparent;
                    }
                `;
					} else {
						style.textContent = `
                    a.chip[data-domain="${domain}"]::before {
                        background-image: url(${faviconUrl});
                        background-color: transparent;
                    }
                `;
					}
					document.head.appendChild(style);
					chip.setAttribute("data-domain", domain);
					chip.classList.add("favicon-loaded");
				};
				img.onerror = function() {
					console.log(`Favicon not found for ${domain}, using fallback`);
				};
				img.src = faviconUrl;
			} catch (e) {
				console.error("Invalid URL in chip:", fullUrl);
			}
		});
	})
});
