package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

const AppDataDirName = "gcalTimeTracker"
const credentialsFileName = "credentials.json"
const tokenFIleName = "token.json"

// Temporary fox for GoLand error. To be fixed in version 2024.1
// https://youtrack.jetbrains.com/issue/GO-12649
//
//goland:noinspection GoBoolExpressions
func main() {
	appDataDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	if runtime.GOOS == "windows" {
		appDataDir = filepath.Join(appDataDir, "AppData", "Local", AppDataDirName)
	} else {
		appDataDir = filepath.Join(appDataDir, "."+AppDataDirName)
	}

	_, err = os.Stat(appDataDir)
	if os.IsNotExist(err) {
		errDir := os.MkdirAll(appDataDir, 0755)
		if errDir != nil {
			fmt.Println("Error creating directory:", err)
			return
		}
		fmt.Println("Directory created successfully")
		fmt.Println("Copy your app credentials.json file to app directory:", appDataDir)
		os.Exit(0)
	} else if err != nil {
		fmt.Println("Error checking directory:", err)
		return
	} else {
		fmt.Println("Directory already exists")
		if _, err := os.Stat(appDataDir + "/" + credentialsFileName); err == nil {
			credentialsFile, err := os.ReadFile(appDataDir + "/" + credentialsFileName)
			if err != nil {
				fmt.Println("Error opening file:", err)
				return
			}
			config, err := google.ConfigFromJSON(credentialsFile, calendar.CalendarScope)
			if err != nil {
				log.Fatalf("Unable to parse client secret file to config: %v", err)
			}
			token, err := getToken(appDataDir, *config)
			if err != nil {
				log.Fatalf("error: %v", err)
			}

			client := config.Client(context.Background(), token)
			srv, err := calendar.NewService(context.Background(), option.WithHTTPClient(client))
			if err != nil {
				log.Fatalf("Unable to create Calendar service: %v", err)
			}
			calendars, err := srv.CalendarList.List().Do()
			if len(calendars.Items) > 0 {
				fmt.Println("Calendars:")
				for _, item := range calendars.Items {
					fmt.Printf("- %s (%s)\n", item.Summary, item.Id)
				}
			} else {
				fmt.Println("No calendars found.")
			}

			os.Exit(0)
		} else if os.IsNotExist(err) {
			fmt.Println("Credentials file does not exist.")
		} else {
			fmt.Println("Error:", err)
		}

	}
}

func getToken(appDataDir string, config oauth2.Config) (token *oauth2.Token, error error) {
	_, err := os.Stat(appDataDir + "/" + tokenFIleName)
	if os.IsNotExist(err) {
		authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
		fmt.Printf("Go to the following link in your browser then type the "+
			"authorization code: \n%v\n", authURL)
		var code string
		if _, err := fmt.Scan(&code); err != nil {
			log.Fatalf("Unable to read authorization code: %v", err)
		}
		token, err := config.Exchange(context.Background(), code)
		if err != nil {
			log.Fatalf("Unable to retrieve token from web: %v", err)
		}
		jsonData, err := json.Marshal(token)
		if err != nil {
			fmt.Println("Unable to serialize: ", token)
		}
		err = os.WriteFile(appDataDir+"/"+tokenFIleName, jsonData, 0644)
		if err != nil {
			log.Fatalf("Error writing to file: %v", err)
		}
		return token, nil

	} else if err != nil {
		return nil, errors.New("error checking token file")
	} else {
		tokenFile, err := os.ReadFile(appDataDir + "/" + tokenFIleName)
		if err != nil {
			return nil, errors.New("error reading token from file")
		}
		var token oauth2.Token
		err = json.Unmarshal(tokenFile, &token)
		if err != nil {
			return nil, errors.New("error deserializing token from file")
		}
		return &token, nil
	}
}
