// Handles all the business logic for the twitter bot.
package main

import (
	"log"
	"os"
)

// Starts the bot, creating the twitter connection as well as performing
// all the actions until an fatal exception occurs.
func startBot() {
	createConnection(config.TwitterAccess)

	if unfollowAll {
		unfollowAllFromUserAndExit(config.TwitterName)
	}

	for {
		followEntries := readFromFile("follows.csv")

		// Unfollow all previous followed users
		if clean {
			cleanFollowListAndExit(followEntries)
		}

		unfollowOldUsers(followEntries)
		userIDsFollowed := getMapOfFollowedUsers(config.TwitterName)
		followEntries = followNewUsers(followEntries, userIDsFollowed)
		writeListOfFollowsToFile(followEntries)
	}
}

// Cleans the provided list of user entities, meaning that all of them
// will be unfollowed.
func cleanFollowListAndExit(userEntities []UserEntity) {
	log.Printf("Unfollowing all followed users in list")
	for index, element := range userEntities {
		unfollow(element.ScreenName)
		log.Printf("[%d] Unfollowed: %s", index, element.ScreenName)
	}

	writeListOfFollowsToFile([]UserEntity{})
	os.Exit(3)
}

// Unfollows previous users that we followed.
// Users are considered old if they are older than the configured hours.
func unfollowOldUsers(userEntities []UserEntity) {
	log.Printf("Checking if anyone needs to be unfollowed")
	for index, element := range userEntities {
		if element.FollowedTimestamp < makeTimestampHoursBeforeNow(followHours) {
			if len(userEntities) > index {
				unfollow(element.ScreenName)
				userEntities = remove(userEntities, index)
				log.Printf("[%d] Unfollowed: %s", index, element.ScreenName)
			}
		} else {
			log.Printf("[%d] user %s isn't due for unfollow yet", index, element.ScreenName)
		}
	}
}

// Unfollows all users of a provided twitterName.
// Recursively calls itself until finished.
func unfollowAllFromUserAndExit(twitterName string) {
	users := listFollows(twitterName)
	if len(users) < 1 {
		log.Printf("User %s doesn't follow any more users. Exiting because work is done.", twitterName)
		os.Create("follows.csv")
		os.Exit(3)
	} else {
		for index, element := range users {
			unfollow(element)
			log.Printf("[%d] Unfollowed: %s", index, element)
		}

		unfollowAllFromUserAndExit(twitterName)
	}
}

// Follows all the users in the provided list of user entities.
func followNewUsers(userEntities []UserEntity, userIDsFollowed map[int64]bool) []UserEntity {
	log.Printf("\nSearching for new users to follow")
	users := searchTweets(randomElementFromSlice(config.Interests), opsBeforeSleep)
	for index, element := range users {
		if !userIDsFollowed[element.UserID] {
			userEntities = append(userEntities, UserEntity{
				ScreenName:        element.ScreenName,
				FollowedTimestamp: makeTimestamp(),
			})
			follow(element.ScreenName)
			log.Printf("[%d] followed: %s", index, element.ScreenName)
		} else {
			log.Printf("User %s is already followed, skipping that user", element.ScreenName)
		}
	}

	return userEntities
}
