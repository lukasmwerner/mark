/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/PuerkitoBio/goquery"
	"github.com/lukasmwerner/mark/store"
	"github.com/spf13/cobra"
)

var tags []string
var title string
var description string

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a bookmark",
	Long: `Adds a new bookmark to the bookmark manager. 

> NOTICE: This method may call out to the network to gather more info about the page

Example:
mark add [--tags list,of,seperated,tags] url`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires a url")
		}
		if _, err := url.Parse(args[0]); err != nil {
			return errors.Join(fmt.Errorf("unable to parse url: %s", args[0]), err)
		}
		return nil
	},

	Run: func(cmd *cobra.Command, args []string) {
		link, _ := url.Parse(args[0])

		fmt.Println("link", link)
		fmt.Println("tags", tags)

		db, err := store.Open()
		if err != nil {
			log.Fatalln("error occured in opening db: ", err.Error())
			return
		}

		defer db.Close()


		if title == "" {
			title = fetchTitle(link)
		}

		bm := store.Bookmark{
			Url:         link.String(),
			Tags:        tags,
			Title:       title,
			Description: description,
		}

		id, err := store.InsertBookmark(db, bm)
		if err != nil {
			log.Fatalln("unable to save bookmark: ", err.Error())
			return
		}

		err = store.InsertTagsAndAssociate(db, id, bm.Tags)
		if err != nil {
			log.Fatalln("unable to save tags: ", err.Error())
			return
		}


	},
}

func init() {
	rootCmd.AddCommand(addCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// addCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// addCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	addCmd.Flags().StringSliceVar(&tags, "tags", []string{}, "Tags for bookmark")
	addCmd.Flags().StringVarP(&title, "title", "t", "", "Overrides the title from the scraper")
	addCmd.Flags().StringVarP(&description, "description", "d", "", "Sets the link's description")
}



func fetchTitle(u *url.URL) string {
	resp, err := http.Get(u.String())
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return ""
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return ""
	}

	return doc.Find("title").Text()
}
