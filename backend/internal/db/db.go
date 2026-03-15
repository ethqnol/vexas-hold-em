package db

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
)

var Client *firestore.Client

func InitFirestore() {
	ctx := context.Background()

	// connects automatically using google_application_credentials
	// in dev, or app default credentials in prod
	conf := &firebase.Config{ProjectID: "vexasholdem"}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		log.Fatalf("Error initializing Firebase app: %v\n", err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		log.Printf("⚠️ Warning: Could not initialize Firestore client. Please ensure GOOGLE_APPLICATION_CREDENTIALS environment variable is set. Error: %v\n", err)
		return
	}

	Client = client
	log.Println("Successfully connected to Firestore database as Admin.")
}

func CloseFirestore() {
	if Client != nil {
		if err := Client.Close(); err != nil {
			log.Printf("Error closing Firestore client: %v\n", err)
		} else {
			log.Println("Firestore client closed.")
		}
	}
}
