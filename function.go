package function

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/storage"
	"github.com/vmihailenco/msgpack" // Import MessagePack package

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

func init() {
	functions.HTTP("TransferFile", transferFile)
}

type ErrorData struct {
	Message string `msgpack:"message"`
}

type Response struct {
	Status string `msgpack:"status"`
}

func sendResponse(w http.ResponseWriter, message string, status int) {
	errorData := &ErrorData{Message: message}

	// Serialize the error response to MessagePack
	errorBytes, err := msgpack.Marshal(errorData)
	if err != nil {
		log.Printf("Failed to marshal error data: %v", err)
		http.Error(w, "Failed to create error response", http.StatusInternalServerError)
		return
	}

	// Respond with the MessagePack error response
	w.Header().Set("Content-Type", "application/msgpack")
	w.WriteHeader(status) // Use an appropriate HTTP status code
	w.Write(errorBytes)
}

func sendError(w http.ResponseWriter, message string) {
	sendResponse(w, message, http.StatusBadRequest)
}

func sendStatus(w http.ResponseWriter, message string) {
	sendResponse(w, message, http.StatusOK)
}

// TransferFile transfers a file from one Google Cloud Storage bucket to another.
func transferFile(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Extract source and destination bucket names from the query parameters.
	sourceBucket := r.URL.Query().Get("sourceBucket")
	destinationBucket := r.URL.Query().Get("destinationBucket")

	test := os.Getenv("TEST_SOURCE_BUCKET")
	prod := os.Getenv("PROD_SOURCE_BUCKET")

	if sourceBucket != test && sourceBucket != prod {
		sendError(w, "Invalid sourceBucket")
		return
	}

	if sourceBucket == "" || destinationBucket == "" {
		sendError(w, "Both sourceBucket and destinationBucket query parameters are required")
		return
	}

	// Initialize the Google Cloud Storage client
	client, err := storage.NewClient(ctx)
	if err != nil {
		sendError(w, "Failed to create client")
		return
	}
	defer client.Close()

	// Get a handle to the source and destination buckets
	source := client.Bucket(sourceBucket)
	destination := client.Bucket(destinationBucket)

	// Specify the file to transfer (you can change this as needed)
	sourceObject := source.Object("source-file.txt")
	destinationObject := destination.Object("destination-file.txt")

	// Copy the file from the source bucket to the destination bucket
	_, err = destinationObject.CopierFrom(sourceObject).Run(ctx)
	if err != nil {
		sendError(w, fmt.Sprintf("Failed to copy file: %v", err))
		return
	}

	sendStatus(w, "OK")
}
