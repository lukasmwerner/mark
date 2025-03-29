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
