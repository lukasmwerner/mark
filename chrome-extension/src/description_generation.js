import { Ollama } from "ollama/browser";
import { description_schema } from "./schema"

async function generateDescription(title, url, content) {
	try {
		// Show loading indicator
		document.getElementById("descLoading").style.display = "flex";

		// Get Ollama settings
		const {
			ollamaEndpoint,
			ollamaSummaryModel: ollamaModel,
			ollamaSummaryPrompt: ollamaPrompt,
			ollamaSummaryTemperature: temperature,
		} = await chrome.storage.sync.get([
			"ollamaEndpoint",
			"ollamaSummaryModel",
			"ollamaSummaryPrompt",
			"ollamaSummaryTemperature",
		]);
		console.log({ ollamaEndpoint, ollamaModel, ollamaPrompt, temperature })

		const endpoint = ollamaEndpoint || "http://localhost:11434";
		const model = ollamaModel || "gemma3:4b";

		// Use custom prompt if available, otherwise use default
		let promptTemplate = ollamaPrompt ||
			`Please generate a short and simple description of the the following document in the style of a website description.
Document: {{content}}`;

		// Replace placeholders with actual content
		const prompt = promptTemplate
			.replace(/{{title}}/g, title)
			.replace(/{{url}}/g, url)
			.replace(/{{content}}/g, content);

		// Initialize Ollama client
		const ollamaClient = new Ollama({
			host: endpoint,
		});

		const response = await ollamaClient.generate({
			model: model,
			prompt: prompt,
			options: {
				temperature: temperature || 0.3,
			},
			format: description_schema,
		});

		// Hide loading indicator
		document.getElementById("descLoading").style.display = "none";

		let desc = JSON.parse(response.response).description;

		return desc;
	} catch (error) {
		// Hide loading indicator
		document.getElementById("descLoading").style.display = "none";

		console.error("Error generating description with Ollama:", error);
		updateStatus(`Error generating description: ${error.message}`);
		return [];
	}
}

export {
	generateDescription,
}
