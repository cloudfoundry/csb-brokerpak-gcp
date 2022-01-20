package environment

import "os"

type GCPMetadata struct {
	Project     string
	Credentials string
}

func ReadGCPMetadata() GCPMetadata {
	return GCPMetadata{
		Project:     os.Getenv("GOOGLE_PROJECT"),
		Credentials: os.Getenv("GOOGLE_CREDENTIALS"),
	}
}
