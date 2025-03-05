import { Ollama } from "ollama/browser";

const ollama_schema = {
  type: "object",
  properties: {
    tags: {
      type: "array",
      items: { type: "string" },
    },
  },
  required: ["tags"],
};

// Initialize Ollama client
let ollamaClient = null;

chrome.runtime.onInstalled.addListener(async () => {
  chrome.contextMenus.create({
    id: "saveWithTags",
    title: "Save with Tags",
    contexts: ["action"],
  });

  // Initialize Ollama client on install
  await initOllamaClient();
});

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
    async function (tab_reponse) {
      if (chrome.runtime.lastError) {
        console.error("Could not connect to page");
        return;
      }

      try {
        // Get the API token
        const { apiToken } = await chrome.storage.sync.get(["apiToken"]);
        if (!apiToken) {
          // Can't check without API token
          return;
        }

        // Check if the URL is bookmarked
        const response = await fetch(
          `http://localhost:1990/api/bookmarks?url=${encodeURIComponent(
            tab_reponse.url,
          )}`,
          {
            method: "GET",
            headers: {
              Authorization: `Bearer ${apiToken}`,
            },
          },
        );

        console.log("Response status:", response.status);

        if (response.ok) {
          // URL is bookmarked
          setFilledIcon();
        } else if (response.status === 404) {
          // URL is not bookmarked
          setDefaultIcon();
        } else {
          // Error occurred
          console.error("Error checking bookmark status:", response.statusText);
          setDefaultIcon();
        }
      } catch (error) {
        console.error("Error in checkIfBookmarked:", error);
        setDefaultIcon();
      }
    },
  );
}

async function initOllamaClient() {
  try {
    const { ollamaEndpoint, ollamaModel } = await chrome.storage.sync.get([
      "ollamaEndpoint",
      "ollamaModel",
    ]);

    // Default to local Ollama instance and a general model if not configured
    const endpoint = ollamaEndpoint || "http://localhost:11434";
    const model = ollamaModel || "gemma:7b";

    ollamaClient = new Ollama({
      host: endpoint,
    });

    console.log("Ollama client initialized");
  } catch (error) {
    console.error("Error initializing Ollama client:", error);
  }
}

chrome.action.onClicked.addListener((tab) => {
  quickSave(tab);
});

chrome.contextMenus.onClicked.addListener(async (info, tab) => {
  if (info.menuItemId === "saveWithTags") {
    // First get the page info
    try {
      const response = await chrome.tabs.sendMessage(tab.id, {
        action: "getPageInfo",
      });
      // Store the data temporarily
      await chrome.storage.local.set({ tempPageInfo: response });

      // Open the popup with the tab id
      chrome.windows.create({
        url: "popup.html",
        type: "popup",
        width: 400,
        height: 500,
      });
    } catch (error) {
      console.error("Error getting page info:", error);
    }
  }
});

async function generateTagsWithOllama(title, description) {
  if (!ollamaClient) {
    await initOllamaClient();

    if (!ollamaClient) {
      console.error("Failed to initialize Ollama client");
      return [];
    }
  }

  try {
    const { ollamaModel, ollamaPrompt } = await chrome.storage.sync.get([
      "ollamaModel",
      "ollamaPrompt",
    ]);
    const model = ollamaModel || "gemma:7b";

    // Use custom prompt if available, otherwise use default
    let promptTemplate =
      ollamaPrompt ||
      `Generate 3 or more relevant tags for this content.
      return as JSON

      Title: {{title}}
      Description: {{description}}`;

    // Replace placeholders with actual content
    const prompt = promptTemplate
      .replace(/{{title}}/g, title)
      .replace(/{{description}}/g, description);

    const response = await ollamaClient.generate({
      model: model,
      prompt: prompt,
      options: {
        temperature: 0.3,
      },
      format: ollama_schema,
    });

    return JSON.parse(response.response).tags;
  } catch (error) {
    console.error("Error generating tags with Ollama:", error);
    return [];
  }
}

async function quickSave(tab) {
  try {
    // Check for token
    const { apiToken, enableAutoTagging } = await chrome.storage.sync.get([
      "apiToken",
      "enableAutoTagging",
    ]);
    if (!apiToken) {
      chrome.runtime.openOptionsPage();
      return;
    }

    // Get page info from content script
    chrome.tabs.sendMessage(
      tab.id,
      { action: "getPageInfo" },
      async function (response) {
        if (chrome.runtime.lastError) {
          console.error("Could not connect to page");
          return;
        }

        // Generate tags using Ollama only if auto-tagging is enabled
        let autoTags = [];
        if (enableAutoTagging !== false) {
          // Default to enabled if not set
          autoTags = await generateTagsWithOllama(
            response.title,
            response.description,
          );
        }

        // Prepare bookmark data
        const bookmark = {
          Url: response.url,
          Title: response.title,
          Description: response.description,
          Tags: autoTags || [], // Use generated tags
        };

        try {
          // Send to API
          const apiResponse = await fetch(
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

          if (!apiResponse.ok) {
            throw new Error(`HTTP error! status: ${apiResponse.status}`);
          }

          // Show success badge
          chrome.action.setBadgeText({ text: "✓" });
          chrome.action.setBadgeBackgroundColor({ color: "#4CAF50" });
          setTimeout(() => {
            chrome.action.setBadgeText({ text: "" });
          }, 2000);
        } catch (error) {
          console.error("Error:", error);
          // Show error badge
          chrome.action.setBadgeText({ text: "!" });
          chrome.action.setBadgeBackgroundColor({ color: "#F44336" });
          setTimeout(() => {
            chrome.action.setBadgeText({ text: "" });
          }, 2000);
        }
      },
    );
  } catch (error) {
    console.error("Error:", error);
  }
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
