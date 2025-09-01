import { generateTagsWithOllama } from "./tag_generation";

function createTagElement(tag) {
	const tagElement = document.createElement("span");
	tagElement.className = "tag";

	const tagText = document.createTextNode(tag);
	const removeButton = document.createElement("button");
	removeButton.innerHTML = "&times;";
	removeButton.addEventListener("click", () => removeTag(tag, tagElement));

	tagElement.appendChild(tagText);
	tagElement.appendChild(removeButton);

	return tagElement;
}

function addTag(tag) {
	if (tags.has(tag)) return;

	tags.add(tag);
	const tagElement = createTagElement(tag);
	const container = document.getElementById("tagsContainer");
	container.insertBefore(tagElement, document.getElementById("tagInput"));
}

function removeTag(tag, element) {
	tags.delete(tag);
	element.remove();
}

function updateStatus(message) {
	const status = document.getElementById("status");
	status.textContent = message;
}


let pageInfo = null;
let tags = new Set();

async function initializeDetailedSave() {
	try {
		const tabs = await chrome.tabs.query({ active: true, currentWindow: true });
		if (!tabs[0]) return;
		const response = await chrome.tabs.sendMessage(tabs[0].id, {
			action: "getPageInfo",
		});
		pageInfo = response;
		document.getElementById("title").value = pageInfo.title;
		document.getElementById("description").value = pageInfo.description;
	} catch (error) {
		updateStatus("Error: Could not load page data");
		console.error(error);
	}

	// Handle tag input
	const tagInput = document.getElementById("tagInput");
	tagInput.addEventListener("keydown", (e) => {
		if (e.key === "Enter" && e.target.value.trim()) {
			e.preventDefault();
			addTag(e.target.value.trim());
			e.target.value = "";
		}
	});

	// Handle manual tag generation
	document
		.getElementById("generateTags")
		.addEventListener("click", async () => {
			if (!pageInfo) {
				updateStatus("Error: Page information not available");
				return;
			}

			const generatedTags = await generateTagsWithOllama(
				pageInfo.title,
				pageInfo.url,
				document.getElementById("description").value,
			);

			if (generatedTags && generatedTags.length > 0) {
				// Add each generated tag
				generatedTags.forEach((tag) => addTag(tag));
				updateStatus(`Generated ${generatedTags.length} tags`);
			} else {
				updateStatus("No tags could be generated");
			}
		});

	// Handle save
	document
		.getElementById("saveBookmark")
		.addEventListener("click", async () => {
			try {
				const { apiToken } = await chrome.storage.sync.get([
					"apiToken",
				]);
				if (!apiToken) {
					updateStatus(
						"Please set your API token in extension options",
					);
					return;
				}

				if (!pageInfo) {
					updateStatus("Error: Page information not available");
					return;
				}

				const bookmark = {
					Url: pageInfo.url,
					Title: document.getElementById("title").value,
					Description: document.getElementById("description").value,
					Tags: Array.from(tags),
				};

				const response = await fetch(
					"http://localhost:1990/api/bookmarks",
					{
						method: "POST",
						headers: {
							"Content-Type": "application/json",
							Authorization: `Bearer ${apiToken}`,
						},
						body: JSON.stringify(bookmark),
					},
				);

				if (!response.ok) {
					throw new Error(`HTTP error! status: ${response.status}`);
				}

				updateStatus("Saved successfully!");
				setTimeout(() => window.close(), 1000);
			} catch (error) {
				updateStatus(`Error: ${error.message}`);
			}
		});
}

document.addEventListener("DOMContentLoaded", function() {
	console.log("dom loaded!")
	initializeDetailedSave()
});
