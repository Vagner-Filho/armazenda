
# Project Overview

This project is a web application named "armazenda", written in Go. It appears to be a farm management system, with features related to crops, fields, vehicles, entries, departures, and reports. The application uses the Gin web framework for routing and handling HTTP requests. The frontend is built with HTML templates and HTMX to handle requests a HTML fragment swapping and styled with Tailwind CSS.

# Building and Running

## Running the application

The project uses `air` for live reloading during development. To run the application, use the following command:

```bash
air
```

This will start the web server on port 8100.

## Building the application

To build the application, use the following command:

```bash
go build -o ./tmp/main .
```

# Development Conventions

## Backend

The backend is written in Go and follows a modular structure. The code is organized into the following directories:

*   `model`: Contains the data models and database logic.
*   `router`: Defines the HTTP routes and their handlers.
*   `service`: Implements the business logic of the application.
*   `entity`: Contains the data structures used throughout the application.
*   `view`: Contains the logic for rendering the HTML templates.

## Frontend

The frontend is built with HTML templates, HTMX to handle Ajax Requests and HTML Fragment Replacement and styled with Tailwind CSS. The templates are located in the `templates` directory, and the Tailwind CSS configuration is in `tailwind.config.js`.

## Database

The application uses a PostgreSQL database. The database connection is managed in `model/armazenda_database/database.go`. The database schema is initialized in the `main` function.
