import { Ollama } from "ollama/browser";
import { ollama_schema } from "./schema";

async function generateTagsWithOllama(url, title, description, currentTags) {
	try {
		// Show loading indicator
		document.getElementById("tagsLoading").style.display = "flex";

		// Get Ollama settings
		const {
			ollamaEndpoint,
			ollamaModel,
			ollamaPrompt,
			ollamaTemperature: temperature,
		} = await chrome.storage.sync.get([
			"ollamaEndpoint",
			"ollamaModel",
			"ollamaPrompt",
			"ollamaTemperature",
		]);

		const endpoint = ollamaEndpoint || "http://localhost:11434";
		const model = ollamaModel || "lukasmwerner/mark-tagger:1b";

		// Use custom prompt if available, otherwise use default
		let promptTemplate = ollamaPrompt ||
			`Generate 3 or more relevant tags for this content.
      return as JSON

      Title: {{title}}
	  URL: {{url}}
      Description: {{description}}`;

		// Replace placeholders with actual content
		const prompt = promptTemplate
			.replace(/{{title}}/g, title)
			.replace(/{{url}}/g, url)
			.replace(/{{description}}/g, description)
			.replace(/{{tags}}/g, JSON.stringify(currentTags));

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
			format: ollama_schema,
		});

		// Hide loading indicator
		document.getElementById("tagsLoading").style.display = "none";

		console.log(response);
		let tags = JSON.parse(response.response).tags;

		tags = tags.map(v => v.toLowerCase())
		return tags;
	} catch (error) {
		// Hide loading indicator
		document.getElementById("tagsLoading").style.display = "none";

		console.error("Error generating tags with Ollama:", error);
		updateStatus(`Error generating tags: ${error.message}`);
		return [];
	}
}

export {
	generateTagsWithOllama,
}
