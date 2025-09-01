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
				// Get the API token
				const { apiToken } = await chrome.storage.sync.get([
					"apiToken",
				]);
				if (!apiToken) {
					// Can't check without API token
					return;
				}

				// Check if the URL is bookmarked
				const response = await fetch(
					`http://localhost:1990/api/bookmarks?url=${encodeURIComponent(
						tab_reponse.url,
					)
					}`,
					{
						method: "GET",
						headers: {
							Authorization: `Bearer ${apiToken}`,
						},
					},
				);

				if (response.ok) {
					// URL is bookmarked
					setFilledIcon();
				} else if (response.status === 404) {
					// URL is not bookmarked
					setDefaultIcon();
				} else {
					// Error occurred
					console.error(
						"Error checking bookmark status:",
						response.statusText,
					);
					setDefaultIcon();
				}
			} catch (error) {
				console.error("Error in checkIfBookmarked:", error);
				setDefaultIcon();
			}
		},
	);
}



function setFilledIcon() {
	chrome.action.setIcon({
		path: {
			16: "icon_filled16.png",
			48: "icon_filled48.png",
			128: "icon_filled128.png",
		},
	});
}

function setDefaultIcon() {
	chrome.action.setIcon({
		path: {
			16: "icon16.png",
			48: "icon48.png",
			128: "icon128.png",
		},
	});
}
