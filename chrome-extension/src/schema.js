export const ollama_schema = {
	type: "object",
	properties: {
		tags: {
			type: "array",
			items: { type: "string" },
		},
	},
	required: ["tags"],
};

export const description_schema = {
	type: "object",
	properties: {
		description: {
			type: "string",
		}
	},
	required: ["description"]
};
