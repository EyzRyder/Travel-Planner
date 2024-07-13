# Planejador de Viagens

## Refrences
[Documentation](https://nlw-journey.apidocumentation.com/reference#tag/links/get/trips/{tripId}/links)

[Figma Design](https://www.figma.com/community/file/1392276515495389646https://www.figma.com/community/file/1392276515495389646)

## Requisites
- Docker;
- Go;

## Technologies used
- Go
- Docker
- PostgreSQL

## Setup
- Clone the repository;
```bash
  git clone https://github.com/EyzRyder/Travel-Planner
  cd Travel-Planner
```
- Install dependencies;
```bash
  go mod download && go mod verify
```
- Install dependencies;
```bash
  go generate ./...
```
- there are 2 options setingup aplication
  - Setup App and DB at once
    ```bash
      docker compose up
     ```
  - Set DB and run with go
    ```bash
      docker compose start db
      docker compose start mailer
      # Optianaly setup pgadmin
      # docker compose start pgadmin
      go run ./cmd/journey/journey.go
    ```
  - Test it! (I personally recommend testing with [Hoppscotch](https://hoppscotch.io/)).

## HTTP

### Trips

#### GET `/trips/{tripId}/confirm`

Confirm a trip and send e-mail invitations.​

- Path Parameters `tripId Required string uuid`

- Response
  - 204 - Default Response
  - 400 - Bad request
  ```json
  {
  "message": "…"
  }
  ```

#### POST `/trips`

Create a new trip​

- Request body
  ```json
  {
  "destination": "...", // Required string min: 4
  "starts_at": "2017-07-21T17:32:28Z", //Required string date-time
  "ends_at":"2017-07-21T17:32:28Z", //Required string date-time
  "emails_to_invite":["...","..."], //Required array string[]
  "owner_name":"...", // Required string
  "owner_email":"..." // Required string email
  }
  ```
- Response
  - 201 - Default Response
  ```json
  {
  "tripId": "123e4567-e89b-12d3-a456-426614174000"
  }
  ```
  - 400 - Bad request
  ```json
  {
  "message": "…"
  }
  ```

#### GET `/trips/{tripId}`

Get a trip details.​

- Path Parameters `tripId Required string uuid`

- Response
  - 200 - Default Response
    ```json
       {
        "trip": {
          "id": "123e4567-e89b-12d3-a456-426614174000",
          "destination": "…",
          "starts_at": "2024-07-12T22:07:42.948Z",
          "ends_at": "2024-07-12T22:07:42.948Z",
          "is_confirmed": true
        }
    }
    ```
  - 400 - Bad request
    ```json
    {
    "message": "…"
    }
    ```

#### PUT `/trips/{tripId}`

Update a trip.​

- Path Parameters `tripId Required string uuid`

- Request body
  ```json
  {
  "destination": "...", // Required string min: 4
  "starts_at": "2017-07-21T17:32:28Z", //Required string date-time
  "ends_at":"2017-07-21T17:32:28Z", //Required string date-time
  }
  ```
- Response
  - 204 - Default Response
  - 400 - Bad request
  ```json
  {
  "message": "…"
  }
  ```
### Participants

#### PATCH `/participants/{participantId}/confirm`

Confirms a participant on a trip.​

- Path Parameters `participantId Required string uuid`

- Response
  - 204 - Default Response
  - 400 - Bad request
  ```json
  {
  "message": "…"
  }
  ```

#### PATCH `/trips/{tripId}/invites`

Invite someone to the trip.​

- Path Parameters `tripId Required string uuid`

- Request
```json
{
  "email":"...",// Required string email
}
```
- Response
  - 201 - Default Response
  - 400 - Bad request
  ```json
  {
  "message": "…"
  }
  ```

#### GET `/trips/{tripId}/participants`

Get a trip participants.​

- Path Parameters `tripId Required string uuid`

- Response
  - 200 - Default Response
  ```json
  {
  "participants": [
    {
      "id": "…",
      "name": "…",
      "email": "hello@example.com",
      "is_confirmed": true
    }
  ]
  }
  ```
  - 400 - Bad request
  ```json
  {
  "message": "…"
  }
  ```

### Activities

#### POST `/trips/{tripId}/activities`

Create a trip activity.​

- Path Parameters `tripId Required string uuid`

- Request
```json
  {
    "occurs_at":"2017-07-21T17:32:28Z", //Required string date-time
    "title":"", // Required string
  }
```

- Response
  - 201 - Default Response
  ```json
  {
      "activityId": "123e4567-e89b-12d3-a456-426614174000"
  }
  ```
  - 400 - Bad request
  ```json
  {
  "message": "…"
  }
  ```

#### GET `/trips/{tripId}/activities`

Get a trip activities.​
This route will return all the dates between the trip starts_at and ends_at dates, even those without activities.

- Path Parameters `tripId Required string uuid`

- Response
  - 200 - Default Response
  ```json
  {
      "activities": [
    {
      "date": "2024-07-12T22:19:46.706Z",
      "activities": [
        {
          "id": "123e4567-e89b-12d3-a456-426614174000",
          "title": "…",
          "occurs_at": "2024-07-12T22:19:46.706Z"
        }
      ]
    }
  ]
  }
  ```
  - 400 - Bad request
  ```json
  {
  "message": "…"
  }
  ```

### Links

#### POST `/trips/{tripId}/links`

Create a trip link.​

- Path Parameters `tripId Required string uuid`

- Request
```json
  {
    "title":"...", // Required string
    "url":"...", // Required string url
  }
```

- Response
  - 201 - Default Response
  ```json
    {
    "linkId": "123e4567-e89b-12d3-a456-426614174000"
    }
  ```
  - 400 - Bad request
  ```json
    {
    "message": "…"
    }
  ```

#### GET `/trips/{tripId}/links`

Get a trip links.​

- Path Parameters `tripId Required string uuid`

- Response
  - 200 - Default Response
    ```json
      {
        "links": [
          {
            "id": "123e4567-e89b-12d3-a456-426614174000",
            "title": "…",
            "url": "https://example.com"
          }
        ]
      }
    ```
  - 400 - Bad request
    ```json
    {
    "message": "…"
    }
    ```
