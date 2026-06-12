package core

import (
	"log"
	"os"
)

var SupabaseURL string
var SupabaseKey string

func InitSupabase() {
	SupabaseURL = os.Getenv("SUPABASE_URL")
	SupabaseKey = os.Getenv("SUPABASE_KEY")

	if SupabaseURL == "" {
		log.Fatal("SUPABASE_URL environment variable is required")
	}
	if SupabaseKey == "" {
		log.Fatal("SUPABASE_KEY environment variable is required")
	}

	log.Println("Supabase API configuration initialized successfully")
}
