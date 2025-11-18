async function createBookmark(apiKey, {
	url,
	title,
	description,
	tags,
}) {
	const bookmark = {

		Url: url,
		Title: title,
		Description: description,
		Tags: tags,
	}

	const response = await fetch(
		"http://localhost:1990/api/bookmarks",
		{
			method: "POST",
			headers: {
				"Content-Type": "application/json",
				Authorization: `Bearer ${apiKey}`,
			},
			body: JSON.stringify(bookmark),
		},
	);


	return response;
}


async function updateBookmark(apiKey, original_url, {
	url,
	title,
	description,
	tags,
}) {
	const bookmark = {

		Url: url,
		Title: title,
		Description: description,
		Tags: tags,
	}

	const response = await fetch(
		`http://localhost:1990/api/bookmarks?url=${encodeURIComponent(
			original_url,
		)
		}`,
		{
			method: "PATCH",
			headers: {
				"Content-Type": "application/json",
				Authorization: `Bearer ${apiKey}`,
			},
			body: JSON.stringify(bookmark),
		},
	);


	return response;
}

async function getBookmark(apiKey, url) {
	const response = await fetch(
		`http://localhost:1990/api/bookmarks?url=${encodeURIComponent(
			url,
		)
		}`,
		{
			method: "GET",
			headers: {
				Authorization: `Bearer ${apiKey}`,
			},
		},
	);
	return response;
}

export {
	getBookmark,
	updateBookmark,
	createBookmark
}
