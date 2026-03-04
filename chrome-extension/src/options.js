import { Ollama } from "ollama/browser";
import { ollama_schema } from "./schema";
import { description_prompt, tagging_prompt } from "./default_prompts";

document.addEventListener("DOMContentLoaded", function() {
	// Load saved settings
	chrome.storage.sync.get(
		[
			"apiToken",
			"iconColor",
			"ollamaEndpoint",
			"ollamaModel",
			"ollamaPrompt",
			"ollamaTemperature",
			"ollamaSummaryModel",
			"ollamaSummaryPrompt",
			"ollamaSummaryTemperature",
		],
		function(result) {
			console.log(result);
			if (result.apiToken) {
				document.getElementById("tokenInput").value = result.apiToken;
			}

			if (result.iconColor) {
				document.getElementById("iconColor").value = result.iconColor;
			}

			if (result.ollamaEndpoint) {
				document.getElementById("ollamaEndpoint").value =
					result.ollamaEndpoint;
			}

			if (result.ollamaModel) {
				document.getElementById("ollamaModel").value =
					result.ollamaModel;
			}

			if (result.ollamaTemperature) {
				document.getElementById("ollamaTemperature").value =
					result.ollamaTemperature;
			}

			if (result.ollamaPrompt) {
				document.getElementById("ollamaPrompt").value =
					result.ollamaPrompt;
			} else {
				// Set default prompt if not already set
				document.getElementById("ollamaPrompt").value =
					tagging_prompt;
			}


			if (result.ollamaSummaryModel) {
				document.getElementById("ollamaSummaryModel").value =
					result.ollamaSummaryModel;
			}

			if (result.ollamaSummaryTemperature) {
				document.getElementById("ollamaSummaryTemperature").value =
					result.ollamaSummaryTemperature;
			}

			if (result.ollamaSummaryPrompt) {
				document.getElementById("ollamaSummaryPrompt").value =
					result.ollamaSummaryPrompt;
			} else {
				// Set default prompt if not already set
				document.getElementById("ollamaSummaryPrompt").value =
					description_prompt;
			}
		},
	);

	// Save settings
	document
		.getElementById("saveSettings")
		.addEventListener("click", function() {
			const token = document.getElementById("tokenInput").value;
			const iconColor = document.getElementById("iconColor").value;
			const ollamaEndpoint =
				document.getElementById("ollamaEndpoint").value ||
				"http://localhost:11434";
			const ollamaModel = document.getElementById("ollamaModel").value ||
				"lukasmwerner/mark-tagger:1b";
			const ollamaPrompt =
				document.getElementById("ollamaPrompt").value || tagging_prompt;
			const ollamaTemperature = new Number(
				document.getElementById("ollamaTemperature").value,
			) || 0.3;



			const ollamaSummaryModel = document.getElementById("ollamaSummaryModel").value ||
				"qwen3.5:2b";
			const ollamaSummaryPrompt =
				document.getElementById("ollamaSummaryPrompt").value ||
				description_prompt;
			const ollamaSummaryTemperature = new Number(
				document.getElementById("ollamaSummaryTemperature").value,
			) || 0.3;

			chrome.storage.sync.set(
				{
					apiToken: token,
					iconColor: iconColor,
					ollamaEndpoint: ollamaEndpoint,
					ollamaModel: ollamaModel,
					ollamaPrompt: ollamaPrompt,
					ollamaTemperature: ollamaTemperature.valueOf(),
					ollamaSummaryModel: ollamaSummaryModel,
					ollamaSummaryPrompt: ollamaSummaryPrompt,
					ollamaSummaryTemperature: ollamaSummaryTemperature.valueOf(),
				},
				function() {
					updateStatus("Settings saved successfully!");
				},
			);
		});

	// Reset Icon Color
	document
		.getElementById("resetIconColor")
		.addEventListener("click", function() {
			document.getElementById("iconColor").value = "#000000";
		});

	// Test Ollama connection
	document
		.getElementById("testOllama")
		.addEventListener("click", async function() {
			try {
				updateStatus("Testing Ollama connection...");

				const endpoint =
					document.getElementById("ollamaEndpoint").value ||
					"http://localhost:11434";
				const model = document.getElementById("ollamaModel").value ||
					"lukasmwerner/mark-tagger:1b";

				const ollama = new Ollama({
					host: endpoint,
				});

				// Try to get the list of models to verify connection
				const models = await ollama.list();

				// Check if the specified model is available
				const modelExists = models.models.some((m) => m.name === model);

				if (modelExists) {
					updateStatus(
						`Connection successful! Model "${model}" is available.`,
					);
				} else {
					updateStatus(
						`Connection successful, but model "${model}" was not found. Available models: ${models.models.map((m) => m.name).join(", ")
						}`,
					);
				}
			} catch (error) {
				updateStatus(`Error connecting to Ollama: ${error.message}`);
				console.error(error);
			}
		});

	// Test prompt for tag generation
	document
		.getElementById("testPrompt")
		.addEventListener("click", async function() {
			try {
				updateStatus("Testing tag generation...");

				const endpoint =
					document.getElementById("ollamaEndpoint").value ||
					"http://localhost:11434";
				const model = document.getElementById("ollamaModel").value ||
					"llama3.2:3b";
				const promptTemplate =
					document.getElementById("ollamaPrompt").value;
				const ollamaTemperature = new Number(
					document.getElementById("ollamaTemperature").value,
				) || 0.3;

				// Sample data for testing
				const sampleTitle = "Understanding Machine Learning Algorithms";
				const sampleDescription =
					"An overview of common machine learning algorithms including supervised and unsupervised learning approaches, with examples of practical applications.";
				const sampleURL =
					"https://example.com/posts/2022/7/1/Understanding-Machine-Learning";

				// Replace placeholders with sample data
				const prompt = promptTemplate
					.replace(/{{title}}/g, sampleTitle)
					.replace(/{{url}}/g, sampleURL)
					.replace(/{{description}}/g, sampleDescription);

				const ollama = new Ollama({
					host: endpoint,
				});

				const response = await ollama.generate({
					model: model,
					prompt: prompt,
					options: {
						temperature: ollamaTemperature,
					},
					format: ollama_schema,
				});

				let tags = JSON.parse(response.response).tags;
				// Display the results
				const resultsDiv = document.getElementById("promptTestResults");
				resultsDiv.style.display = "block";

				const tagsDiv = document.getElementById("generatedTags");
				if (tags.length > 0) {
					tagsDiv.innerHTML = tags
						.map(
							(tag) =>
								`<span style="display: inline-block; background: #e0e0e0; padding: 4px 8px; border-radius: 4px; margin: 2px; font-size: 14px;">${tag}</span>`,
						)
						.join("");
					updateStatus("Tag generation successful!");
				} else {
					tagsDiv.innerHTML =
						"<p>No tags could be parsed from the response.</p>";
					tagsDiv.innerHTML += "<p><strong>Raw response:</strong> " +
						response.response + "</p>";
					updateStatus(
						"Warning: Could not parse tags from the response.",
					);
				}
			} catch (error) {
				updateStatus(`Error testing prompt: ${error.message}`);
				console.error(error);
			}
		});
});

function updateStatus(message) {
	const status = document.getElementById("status");
	status.textContent = message;
	setTimeout(() => {
		if (status.textContent === message) {
			status.textContent = "";
		}
	}, 5000);
}
