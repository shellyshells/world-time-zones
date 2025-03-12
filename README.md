# World Time Zones
A comprehensive web application built with Go that provides real-time information about different countries, their time zones, and development statistics.

## Prerequisites

- Git
- Go (version 1.16 or later)
- A text editor or IDE (e.g., Visual Studio Code, GoLang, or Sublime Text)

## File Tree Structure 

```
📦 
HDR23-24_Statistical_Annex_HDI_Table - HDI.csv
README.md
data
part-1.geojson
part-2.geojson
part-3.geojson
part-4.geojson
part-5.geojson
part-6.geojson
part-7.geojson
part-8.geojson
part-9.geojson
documentation.pdf
favorites.json
go.mod
main.go
src
config.go
handlers.go
main.go
models.go
services.go
storage.go
utils.go
static
css
about.css
error.css
favorites.css
home.css
map.css
images
1.png
2.png
3.png
4.png
js
about.js
favorites.js
home.js
map.js
templates
about.html
error.html
favorites.html
home.html
map.html
```

## Setup and Installation

1. Clone the repository:
   ```
   git clone https://github.com/shellyshells/somebodythatiusedtoknow.git
   cd somebodythatiusedtoknow
   ```

2. Install dependencies:
   ```
   go mod tidy
   ```

## Running the Application

1. Start the server:
   ```
   go run main.go
   ```

2. You should see the message: "Server is running on http://localhost:8080"

## Testing the Application

Open a web browser and enter http://localhost:8080/

## Troubleshooting

- Ensure that port 8080 is not being used by another application.
- If you encounter any errors, make sure you have the latest version of Go installed or restart your computer.
- In the case of 'Exit Status 1', restart the IDE, ensure multiple terminals are not running simultaneously or spam 'go run main.go'. 

*Note: The loading of the timezones takes a few seconds.*