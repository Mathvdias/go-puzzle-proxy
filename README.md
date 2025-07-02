Gemini Proxy API with Cache
This project implements a Go API that acts as a proxy for the Google Gemini API, adding a caching layer using PostgreSQL (Supabase) to optimize performance and reduce costs. Additionally, it serves as a centralized point to protect your Gemini API key, preventing it from being exposed directly in your client application (such as a Dart/Flutter app).

🚀 Technologies Used
Go: Programming language for the API backend.

Docker: For containerization and running the application locally.

PostgreSQL (Supabase): Relational database used for caching Gemini API responses.

Gemini API (Google AI): Generative AI service for puzzle creation.

✨ How It Works
Client Application Request: Your Dart application (or any other client) sends a POST request to the Go API (/generate-puzzle) with the desired puzzle parameters (game type, difficulty, topics, language).

Cache Check: The Go API calculates a unique hash based on the request parameters.

Cache Hit: If a response for that hash already exists in the PostgreSQL database (Supabase), the Go API immediately returns the cached response.

Cache Miss: If the response is not in the cache, the Go API:

Constructs the necessary prompt and JSON response schema for the Gemini API.

Calls the Gemini API with your key (which is securely stored on the Go server).

Receives the response from Gemini.

Saves the new response to the PostgreSQL database for future requests (cache).

Returns the response to the client application.

Key Security: Your Gemini API key is never exposed to the client application, remaining secure in the server environment.

📋 Prerequisites
Before you start, make sure you have the following installed and configured:

Go (v1.22 or higher):

Download and install from golang.org/dl.

Verify installation: go version

Docker Desktop:

Download and install from docker.com/products/docker-desktop.

Ensure Docker Desktop is running before attempting to build or run images.

Supabase Account:

You should already have a Supabase account and project configured for the PostgreSQL database.

Gemini API Key:

Make sure you have your Gemini API key.

⚙️ Environment Setup
Clone the Repository (or create the file structure):
If you haven't already, create the folder and file structure as previously provided (main.go, models.go, database.go, gemini_service.go, Dockerfile, go.mod, go.sum, schema.sql, .env.example).

Install Go Dependencies:
Open your terminal in the root folder of your project (go-puzzle-proxy) and run:

go mod tidy

Configure the Supabase PostgreSQL Database:
You need to ensure the cache table exists in your Supabase database.

Access your Supabase project dashboard at app.supabase.com.

In the left sidebar menu, go to "SQL Editor".

Paste the content of the schema.sql file (provided in the project) and execute the query:

CREATE TABLE IF NOT EXISTS cached_puzzles (
    id SERIAL PRIMARY KEY,
    request_hash TEXT UNIQUE NOT NULL,
    request_params JSONB NOT NULL,
    response_data JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

Verify in the "Table Editor" section that the cached_puzzles table was created successfully.

Create and Configure the .env File:
Create a file named .env (no extension) in the root of your project and add your credentials. The .env.example file serves as a template for the required variables.

DATABASE_URL="YOUR_SUPABASE_DATABASE_URL" # Replace with your actual Supabase database connection string
GEMINI_API_KEY="YOUR_GEMINI_API_KEY" # Replace with your actual Gemini API key
PORT="8080"

Attention: This .env file is crucial for local Docker execution.

🏃 How to Run with Docker
Ensure Docker Desktop is running.

Open your terminal in the root folder of your project.

Build the Docker Image for amd64/linux:
This step is crucial to ensure the image is compatible with most environments, even for local testing.

docker build --platform linux/amd64 -t puzzle-proxy-api:local .

puzzle-proxy-api:local is a custom tag for your local image.

Run the Docker Container:
You'll need to pass the environment variables from your .env file to the Docker container and map the port.

docker run -p 8080:8080 \
  --env-file ./.env \
  puzzle-proxy-api:local

-p 8080:8080: Maps port 8080 on your host machine to port 8080 inside the container.

--env-file ./.env: Tells Docker to load environment variables from your local .env file.

puzzle-proxy-api:local: The name and tag of the image you just built.

Your API will be running at http://localhost:8080.

Test the API Locally with curl:
Open another terminal and run:

curl -X POST \
     -H "Content-Type: application/json" \
     -d '{
           "gameType": "crossword",
           "difficulty": "easy",
           "topics": ["animals", "nature"],
           "language": "pt"
         }' \
     http://localhost:8080/generate-puzzle

You should receive a JSON response with the puzzle data.

📱 Updating the Dart Application
In your Dart/Flutter application, you will need to update the _apiUrl in your GeminiService class to point to the URL of your locally running Docker container:

// In file: lib/services/gemini_service.dart

class GeminiService {
  // ... other code parts ...

  // UPDATE THIS URL TO YOUR LOCAL DOCKER API URL
  final String _apiUrl = 'http://localhost:8080/generate-puzzle';

  // ... the rest of your code ...
}

⚖️ License
This project is licensed under the MIT License. See the LICENSE file in the repository root for more details.