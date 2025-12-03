chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
	if (request.action === "getPageInfo") {
		const pageInfo = {
			url: window.location.href,
			title: document.title,
			description: getDescription(),
		};
		sendResponse(pageInfo);
	}
	if (request.action === "getBodyContent") {
		sendResponse(getContent());
	}
});

function getDescription() {
	// Try meta description first
	const metaDesc =
		document.querySelector('meta[property="og:description"]')?.content ||
		document.querySelector('meta[name="description"]')?.content;
	if (metaDesc) return metaDesc;

	// Fallback to first paragraph
	const firstPara = document.querySelector("p")?.textContent;
	return firstPara ? firstPara.substring(0, 200) + "..." : "";
}

function getContent() {
	// Highest priority query to lowest priority
	let queries = ["article", "#content", "main", ".main", "body"];
	for (let i = 0; i < queries.length; i++) {
		let root = document.querySelector(queries[i]);
		if (root != null) {
			return getTextContent(root);
		}
	}
	return ""
}


function getTextContent(root) {
	let walker = document.createTreeWalker(root,
		NodeFilter.SHOW_TEXT,
		null
	);
	let node;
	let textContent = ""
	while (node = walker.nextNode()) {
		try {
			let tagName = node.parentElement.tagName;
			if (tagName == "SCRIPT" || tagName == "STYLE") { continue; }
		} catch (error) { }
		textContent += node.nodeValue + " ";
	}
	return textContent;
}
