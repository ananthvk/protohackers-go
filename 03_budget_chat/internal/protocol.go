package internal

import (
	"fmt"
	"strings"
)

func isAlpha(c rune) bool {
	return ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z')
}
func isNumeric(c rune) bool {
	return ('0' <= c && c <= '9')
}

// validateName checks if the name is valid, returns true if it's valid, false otherwise
func validateName(name string) bool {
	// Alphanumeric & atleast one character
	foundAtleastOneCharacter := false
	for _, ch := range name {
		if !(isAlpha(ch) || isNumeric(ch)) {
			return false
		}
		foundAtleastOneCharacter = foundAtleastOneCharacter || isAlpha(ch)
	}
	return foundAtleastOneCharacter
}

func formatBroadcast(name, msg string) string {
	return fmt.Sprintf("[%s] %s", name, msg)
}

func formatUserList(users []string) string {
	usersList := strings.Join(users, ", ")
	return fmt.Sprintf("* The room contains: %s", usersList)
}

func formatNotification(username string, message string) string {
	return fmt.Sprintf("* %s: %s", username, message)
}

func formatGreeting() string {
	return "Welcome to budgetchat !! Please provide a name to continue..."
}
