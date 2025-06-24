package internal

func IsAdmin(email string, admins []string) bool {
	for _, adminEmail := range admins {
		if email == adminEmail {
			return true
		}
	}
	return false
}
