package api

import (
	"context"
	"encoding/json"
	"errors"
	"journey/internal/api/spec"
	"journey/internal/pgstore"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type store interface {
    CreateTrip(context.Context,*pgxpool.Pool, spec.CreateTripRequest)(uuid.UUID,error)
	GetParticipant(ctx context.Context, id uuid.UUID) (pgstore.Participant, error)
	ConfirmParticipant(ctx context.Context, id uuid.UUID) error
}

type API struct {
	store     store
	logger    *zap.Logger // us.logger do stander lib tambem serve
	validator *validator.Validate
    pool *pgxpool.Pool
}

func NewAPI(pool *pgxpool.Pool, logger *zap.Logger) API {
	validator := validator.New(validator.WithRequiredStructEnabled())
	return API{pgstore.New(pool), logger, validator,pool}
}

// Confirms a participant on a trip.
// (PATCH /participants/{participantId}/confirm)
func (ap *API) PatchParticipantsParticipantIDConfirm(w http.ResponseWriter, r *http.Request, participantID string) *spec.Response {
	id, err := uuid.Parse(participantID)
	if err != nil {
		return spec.PatchParticipantsParticipantIDConfirmJSON400Response(spec.Error{Message: "uuid invalido"})
	}

	participant, err := ap.store.GetParticipant(r.Context(), id)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return spec.PatchParticipantsParticipantIDConfirmJSON400Response(spec.Error{Message: "participante não encontrado"})
		}
		ap.logger.Error("failed to get participant", zap.Error(err), zap.String("participant_id", participantID))
		return spec.PatchParticipantsParticipantIDConfirmJSON400Response(spec.Error{
			Message: "something went wrong, try again",
		})
	}

	if participant.IsConfirmed {
		return spec.PatchParticipantsParticipantIDConfirmJSON400Response(spec.Error{
			Message: "participant already confirmed",
		})
	}
	if err := ap.store.ConfirmParticipant(r.Context(), id); err != nil {
		ap.logger.Error("failed to confim participant", zap.Error(err), zap.String("participant_id", participantID))
		return spec.PatchParticipantsParticipantIDConfirmJSON400Response(spec.Error{
			Message: "something went wrong, try again",
		})
	}

	return spec.PatchParticipantsParticipantIDConfirmJSON204Response(nil)
}

// Create a new trip
// (POST /trips)
func (ap *API) PostTrips(w http.ResponseWriter, r *http.Request) *spec.Response {
	var body spec.CreateTripRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return spec.PostTripsJSON400Response(spec.Error{Message: "invalid JSON"})
	}

	if err := ap.validator.Struct(body); err != nil {
		return spec.PostTripsJSON400Response(spec.Error{Message: "invalid input: " + err.Error()})
	}

    tripID,err := ap.store.CreateTrip(r.Context(),ap.pool,body)

    if err != nil {
		return spec.PostTripsJSON400Response(spec.Error{Message: "failed to create trip, try again"})
    }

	return spec.PostTripsJSON201Response(spec.CreateTripResponse{TripID: tripID.String()})
}

// Get a trip details.
// (GET /trips/{tripId})
func (ap *API) GetTripsTripID(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	panic("not implemented") // TODO: Implement
}

// Update a trip.
// (PUT /trips/{tripId})
func (ap *API) PutTripsTripID(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	panic("not implemented") // TODO: Implement
}

// Get a trip activities.
// (GET /trips/{tripId}/activities)
func (ap *API) GetTripsTripIDActivities(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	panic("not implemented") // TODO: Implement
}

// Create a trip activity.
// (POST /trips/{tripId}/activities)
func (ap *API) PostTripsTripIDActivities(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	panic("not implemented") // TODO: Implement
}

// Confirm a trip and send e-mail invitations.
// (GET /trips/{tripId}/confirm)
func (ap *API) GetTripsTripIDConfirm(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	panic("not implemented") // TODO: Implement
}

// Invite someone to the trip.
// (POST /trips/{tripId}/invites)
func (ap *API) PostTripsTripIDInvites(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	panic("not implemented") // TODO: Implement
}

// Get a trip links.
// (GET /trips/{tripId}/links)
func (ap *API) GetTripsTripIDLinks(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	panic("not implemented") // TODO: Implement
}

// Create a trip link.
// (POST /trips/{tripId}/links)
func (ap *API) PostTripsTripIDLinks(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	panic("not implemented") // TODO: Implement
}

// Get a trip participants.
// (GET /trips/{tripId}/participants)
func (ap *API) GetTripsTripIDParticipants(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	panic("not implemented") // TODO: Implement
}
