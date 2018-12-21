// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

var cfgFile string
var spreadsheetId string
var wSDeckURL string
var wantJSON bool

const output = "decks.json"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "fumi",
	Short: "wsdeck to google sheets",
	Long: `Export wsdeck to google sheets
Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		var sheet = &sheets.Spreadsheet{}

		// Prints the names and majors of students in a sample spreadsheet:
		// https://docs.google.com/spreadsheets/d/1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms/edit
		// spreadsheetId := "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms"
		// resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()

		fmt.Printf("%#v\n", sheet)
		decks := GetDecks(wSDeckURL)
		if wantJSON {
			log.Println("Export Json")
			var buffer bytes.Buffer
			out, err := os.Create(output)
			if err != nil {
				fmt.Printf(err.Error())
			}
			defer out.Close()

			b, err := json.Marshal(decks)
			if err != nil {
				log.Fatalf("Error on toJson")
			}
			json.Indent(&buffer, b, "", "\t")
			buffer.WriteTo(out)
		} else {
			b, err := ioutil.ReadFile("credentials.json")
			if err != nil {
				log.Fatalf("Unable to read client secret file: %v", err)
			}

			// If modifying these scopes, delete your previously saved token.json.
			config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets")
			if err != nil {
				log.Fatalf("Unable to parse client secret file to config: %v", err)
			}
			client := getClient(config)

			srv, err := sheets.New(client)
			if err != nil {
				log.Fatalf("Unable to retrieve Sheets client: %v", err)
			}

			if spreadsheetId == "" {
				rb := &sheets.Spreadsheet{
					Properties: &sheets.SpreadsheetProperties{
						Title: "wsdeck export",
					},
				}
				resp, err := srv.Spreadsheets.Create(rb).Do()
				if err != nil {
					log.Fatalf("Unable to create sheet: %v", err)
				}
				sheet = resp
				spreadsheetId = sheet.SpreadsheetId
			} else {
				resp, err := srv.Spreadsheets.Get(spreadsheetId).Do()
				if err != nil {
					log.Fatalf("Unable to retrieve sheet: %v", err)
				}
				sheet = resp
			}

			for _, deck := range decks {
				var rangeSheet = fmt.Sprintf("'%v'!A1:M42", deck.Name)
				firstLine := []interface{}{"Code", "Color", "Amount"}
				level0 := [][]interface{}{}
				level1 := [][]interface{}{}
				level2 := [][]interface{}{}
				level3 := [][]interface{}{}
				cx := [][]interface{}{}
				var valueR sheets.ValueRange

				log.Println(deck.Name)
				var addSheet = &sheets.Request{AddSheet: &sheets.AddSheetRequest{Properties: &sheets.SheetProperties{Title: deck.Name}}}
				var request = &sheets.BatchUpdateSpreadsheetRequest{Requests: []*sheets.Request{addSheet}}
				_, err := srv.Spreadsheets.BatchUpdate(spreadsheetId, request).Do()
				if err != nil {
					log.Fatalf("error on sheet creation for %v deck \t %v", deck.Name, err)
				}

				valueR.Values = append(valueR.Values, firstLine)
				for _, card := range deck.Cards {
					value := []interface{}{card.ID, card.Color, card.Amount}
					switch card.Level {
					case 0:
						level0 = append(level0, value)
					case 1:
						level1 = append(level1, value)
					case 2:
						level2 = append(level2, value)
					case 3:
						level3 = append(level3, value)
					default:
						cx = append(cx, value)
					}
				}
				valueR.Values = append(valueR.Values, []interface{}{"Level 0", "", ""})
				valueR.Values = append(valueR.Values, level0...)
				valueR.Values = append(valueR.Values, []interface{}{"Level 1", "", ""})
				valueR.Values = append(valueR.Values, level1...)
				valueR.Values = append(valueR.Values, []interface{}{"Level 2", "", ""})
				valueR.Values = append(valueR.Values, level2...)
				valueR.Values = append(valueR.Values, []interface{}{"Level 3", "", ""})
				valueR.Values = append(valueR.Values, level3...)
				valueR.Values = append(valueR.Values, []interface{}{"CX", "", ""})
				valueR.Values = append(valueR.Values, cx...)

				_, err = srv.Spreadsheets.Values.Update(spreadsheetId, rangeSheet, &valueR).ValueInputOption("RAW").Do()
				if err != nil {
					log.Fatalf("Error when adding value to %v deck : \t %v", deck.Name, err)
				}

			}

		}

	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.Flags().StringVar(&spreadsheetId, "id", "", "spreadsheetId to put export")
	rootCmd.Flags().StringVar(&wSDeckURL, "ws", "", "wsdeck url")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolVarP(&wantJSON, "json", "j", false, "Export to JSon")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".fumi" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".fumi")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
