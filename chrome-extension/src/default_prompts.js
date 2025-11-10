export let tagging_prompt = `Generate 3-5 relevant tags for this content. Return only the tags as a JSON array of strings. No additional explanation needed.

Title: {{title}}
URL: {{url}}
Description: {{description}}`
export let description_prompt = `URL: {{url}}
Content:
{{content}}

Title:
{{title}}

Please write a short google search style description for the following webpage. Avoid describing the overall website and focus on the content on the page. Avoid phrases like 'news article about'. Return only the description as in the json object as a string. No additional explanation needed.`;
