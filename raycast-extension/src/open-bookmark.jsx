import { useState } from "react";
import { Icon, List, Detail, Action, ActionPanel, openExtensionPreferences, getPreferenceValues } from "@raycast/api";
import { useFetch } from "@raycast/utils";

export default function Command() {
  const apiKey = getPreferenceValues().apikey;
  if (!apiKey) {
    return (
      <Detail
        markdown="API key incorrect. Please update it in extension preferences and try again."
        actions={
          <ActionPanel>
            <Action title="Open Extension Preferences" onAction={openExtensionPreferences} />
          </ActionPanel>
        }
      />
    );
  }

  const [searchText, setSearchText] = useState("");
  const [showingDetail, setShowDetail] = useState(false);
  const { isLoading, data } = useFetch(`http://localhost:1990/api/bookmarks/search?q=${searchText}`, {
    headers: {
      Authorization: `Bearer ${apiKey}`,
    },
    execute: searchText != "",
    // to make sure the screen isn't flickering when the searchText changes
    keepPreviousData: true,
  });

  return (
    <List
      isLoading={isLoading}
      searchText={searchText}
      onSearchTextChange={setSearchText}
      throttle
      isShowingDetail={showingDetail}
    >
      {searchText === "" ? (
        <List.EmptyView icon={Icon.Bookmark} title="Search for bookmarks" />
      ) : (
        (data || []).map((item) => (
          <List.Item
            key={item.Url}
            title={item.Title}
            /*subtitle={showingDetail ? "" : new URL(item.Url).hostname}*/
            accessories={[{ text: showingDetail ? null : new URL(item.Url).hostname }]}
            detail={
              <List.Item.Detail
                markdown={item.Description}
                metadata={
                  <List.Item.Detail.Metadata>
                    <List.Item.Detail.Metadata.Label title="Title" text={item.Title} />
                    <List.Item.Detail.Metadata.Link title="URL" text={new URL(item.Url).hostname} target={item.Url} />
                    <List.Item.Detail.Metadata.TagList title="Tags">
                      {(item.Tags || []).map((tag) => (
                        <List.Item.Detail.Metadata.TagList.Item text={tag} key={tag} />
                      ))}
                    </List.Item.Detail.Metadata.TagList>
                  </List.Item.Detail.Metadata>
                }
              />
            }
            actions={
              <ActionPanel>
                <Action.OpenInBrowser url={item.Url} shortcut={{ modifiers: ["cmd"], key: "enter" }} />
                <Action.CopyToClipboard title="Copy URL" content={item.Url} />
                <Action
                  title="Toggle Detail"
                  onAction={() => setShowDetail(!showingDetail)}
                  shortcut={{ modifiers: ["cmd"], key: "d" }}
                />
              </ActionPanel>
            }
          />
        ))
      )}
    </List>
  );
}
