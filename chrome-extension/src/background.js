import * as api from "./api.js";

// Add listener for tab activation and update events
chrome.tabs.onActivated.addListener((activeInfo) => {
	console.log("Tab activated:", activeInfo.tabId);
	checkIfBookmarked(activeInfo.tabId);
});

chrome.tabs.onUpdated.addListener((tabId, changeInfo, tab) => {
	// Only run when the page has finished loading
	if (changeInfo.status === "complete") {
		checkIfBookmarked(tabId);
	}
});

async function checkIfBookmarked(tabId) {
	chrome.tabs.sendMessage(
		tabId,
		{ action: "getPageInfo" },
		async function(tab_reponse) {
			if (chrome.runtime.lastError) {
				console.error("Could not connect to page");
				return;
			}

			try {
				// Get the API token and icon color
				const { apiToken, iconColor } = await chrome.storage.sync.get([
					"apiToken",
					"iconColor",
				]);
				if (!apiToken) {
					// Can't check without API token
					return;
				}

				// Check if the URL is bookmarked
				const response = await api.getBookmark(apiToken, tab_reponse.url);

				if (response.ok) {
					// URL is bookmarked
					await setFilledIcon(iconColor);
				} else if (response.status === 404) {
					// URL is not bookmarked
					await setDefaultIcon(iconColor);
				} else {
					// Error occurred
					console.error(
						"Error checking bookmark status:",
						response.statusText,
					);
					await setDefaultIcon(iconColor);
				}
			} catch (error) {
				console.error("Error in checkIfBookmarked:", error);
				// Try to get iconColor if possible, otherwise default
				const { iconColor } = await chrome.storage.sync.get(["iconColor"]);
				await setDefaultIcon(iconColor);
			}
		},
	);
}



async function setFilledIcon(color) {
	await setIcon(color, "icon_filled");
}

async function setDefaultIcon(color) {
	await setIcon(color, "icon");
}

async function setIcon(color, prefix) {
	// If no color or black is specified, use the default png icons
	if (!color || color === "#000000") {
		chrome.action.setIcon({
			path: {
				16: `${prefix}16.png`,
				32: `${prefix}32.png`,
				48: `${prefix}48.png`,
				128: `${prefix}128.png`,
			},
		});
		return;
	}

	try {
		// Helper to generate ImageData for a specific size
		const generateIcon = async (size) => {
			const response = await fetch(`${prefix}${size}.png`);
			const blob = await response.blob();
			const bitmap = await createImageBitmap(blob);
			
			const canvas = new OffscreenCanvas(size, size);
			const ctx = canvas.getContext("2d");
			
			ctx.drawImage(bitmap, 0, 0);
			ctx.globalCompositeOperation = "source-in";
			ctx.fillStyle = color;
			ctx.fillRect(0, 0, size, size);
			
			return ctx.getImageData(0, 0, size, size);
		};

		// Generate icons for standard sizes
		const [img16, img32, img48, img128] = await Promise.all([
			generateIcon(16),
			generateIcon(32),
			generateIcon(48),
			generateIcon(128)
		]);

		chrome.action.setIcon({
			imageData: {
				16: img16,
				32: img32,
				48: img48,
				128: img128
			}
		});
	} catch (e) {
		console.error(`Failed to generate colored icon for ${prefix}, falling back to default`, e);
		chrome.action.setIcon({
			path: {
				16: `${prefix}16.png`,
				32: `${prefix}32.png`,
				48: `${prefix}48.png`,
				128: `${prefix}128.png`,
			},
		});
	}
}
