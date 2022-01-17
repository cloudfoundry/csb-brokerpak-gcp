package environment

import "os"

type GCPMetadata struct {
	Project     string
	Credentials string
}

func ReadGCPMetadata() GCPMetadata {
	//project := os.Getenv("GOOGLE_PROJECT")
	//credentials := os.Getenv("GOOGLE_CREDENTIALS")

	return GCPMetadata{
		Project:     os.Getenv("GOOGLE_PROJECT"),
		Credentials: os.Getenv("GOOGLE_CREDENTIALS"),
	}
}
