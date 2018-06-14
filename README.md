# GWS Demo Lambda Gateway

API gateway for various demo lambda services for the GWS partner platform.

## Requirements

- golang
- dep

## Adding a service

Lambda services must be added to the `runway/pipeline.json` file. Create a new package in the app package, and add the appropriate Echo router group to `app/app.go`
