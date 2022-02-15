package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/spf13/cobra"
)

// releaseCmd represents the release command
// API doc: https://docs.github.com/cn/rest/reference/releases
var releaseCmd = &cobra.Command{
	Use:   "release <command>",
	Short: "Manage GitHub releases",
}

func init() {
	releaseCmd.AddCommand(NewCmdList())
	rootCmd.AddCommand(releaseCmd)
}

func NewCmdList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [<repo>]",
		Short: "List releases in a repository",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				println(cmd.UsageString())
				return nil
			}
			return runList(fmt.Sprintf(`https://api.github.com/repos/%v/releases`, args[0]))
		},
	}
	return cmd
}

/**
格式为：
[
    {
        "tag_name": "v2.1.1",
        "name": "2.1.1 - Improvements and Fixes",
        "published_at": "2022-02-06T14:24:56Z",
        "assets": [
            {
                "name": "Heroic-2.1.1.AppImage",
                "browser_download_url": "https://github.com/Heroic-Games-Launcher/HeroicGamesLauncher/releases/download/v2.1.1/Heroic-2.1.1.AppImage"
            }
        ]
    }
]

*/
type ReleaseAsset struct {
	Name string `json:"name"`
	Url  string `json:"browser_download_url"`
}

type ReleaseData struct {
	Name   string         `json:"tag_name"`
	Date   string         `json:"published_at"`
	Assets []ReleaseAsset `json:"assets"`
}

func runList(repo string) error {
	request, err := http.NewRequestWithContext(context.Background(), "GET", repo, nil)
	if err != nil {
		panic(err.Error())
	}

	request.Header.Add("User-Agent", `'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36 Edg/87.0.664.66'`)

	response, err := http.DefaultClient.Do(request)

	if err != nil {
		panic(err.Error())
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		panic(err.Error())
	}
	jsonData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err.Error())
	}

	var data []ReleaseData

	err = json.Unmarshal(jsonData, &data)

	if err != nil {
		panic(err.Error())
	}

	for i := 0; i < len(data); i++ {
		release := data[i]
		assets := release.Assets
		for j := 0; j < len(assets); j++ {
			asset := assets[j]
			str := fmt.Sprintf("%v %v %v %v", release.Name, release.Date, asset.Name, asset.Url)
			fmt.Println(str)
		}
	}
	return err
}
