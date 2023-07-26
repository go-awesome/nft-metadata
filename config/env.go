package config

import "os"

var (
	NFT_URL = GetEnvVariable("NFT_URL", "https://ipfs.io")

	AWS_REGION = GetEnvVariable("AWS_REGION", "us-east-2")

	TABLE_NAME = GetEnvVariable("TABLE_NAME", "nft_metadata")
)

func GetEnvVariable(key string, defaultString string) string {
	var getKey = os.Getenv(key)
	if len(getKey) > 0 {
		return getKey
	}
	return defaultString
}
