package main

import "kubecloud/cmd"

func main() {
	cmd.Execute()
	// zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// var serverConfig internal.Config

	// flag.IntVar(&serverConfig.Port, "port", 8080, "Server port")
	// flag.StringVar(&serverConfig.StorageDir, "storage", ".", "Storage directory")
	// flag.StringVar(&serverConfig.JWTSecret, "secret", "", "JWT Secret")
	// flag.IntVar(&serverConfig.AccessTokenExpiry, "access-token-expiry", 15, "Access Token Expiry date in minutes")
	// flag.IntVar(&serverConfig.RefreshTokenExpiry, "refresh-token-expiry", 24, "Refresh Token Expiry date in hours")
	// flag.Parse()

	// if serverConfig.JWTSecret == "" {
	// 	// Generate a random secret if not provided
	// 	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	// 	secret := make([]byte, 32)
	// 	for i := range secret {
	// 		secret[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	// 	}
	// 	serverConfig.JWTSecret = string(secret)
	// 	log.Warn().Msg("No JWT secret provided. Generated a random secret. This is not secure for production use.")
	// }

	// server, err := app.NewApp(serverConfig)
	// if err != nil {
	// 	log.Fatal().Err(err).Msg("Error creating server")
	// }

}
