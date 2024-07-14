package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/EyzRyder/Travel-Planner/internal/api/spec"
	"github.com/EyzRyder/Travel-Planner/internal/pgstore"

	openapi_types "github.com/discord-gophers/goapi-gen/types"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Store interface {
	GetParticipant(ctx context.Context, participantID uuid.UUID) (pgstore.Participant, error)
	ConfirmParticipant(ctx context.Context, participantID uuid.UUID) error
	GetParticipants(ctx context.Context, tripID uuid.UUID) ([]pgstore.Participant, error)
	InviteParticipantToTrip(ctx context.Context, params pgstore.InviteParticipantToTripParams) (uuid.UUID, error)

	CreateTrip(context.Context, *pgxpool.Pool, spec.CreateTripRequest) (uuid.UUID, error)
	GetTrip(ctx context.Context, id uuid.UUID) (pgstore.Trip, error)
	UpdateTrip(ctx context.Context, params pgstore.UpdateTripParams) error

	CreateActivity(ctx context.Context, params pgstore.CreateActivityParams) (uuid.UUID, error)
	GetTripActivities(ctx context.Context, tripID uuid.UUID) ([]pgstore.Activity, error)
}

type Mailer interface {
	SendConfirmTripEmailToTripOwner(uuid.UUID) error
	SendTripConfirmedEmails(tripID uuid.UUID) error
	SendTripConfirmedEmail(tripID, participantID uuid.UUID) error
}

type API struct {
	store     Store
	logger    *zap.Logger // us.logger do stander lib tambem serve
	validator *validator.Validate
	pool      *pgxpool.Pool
	mailer    Mailer
}

func NewAPI(pool *pgxpool.Pool, logger *zap.Logger, mailer Mailer) API {
	validator := validator.New(validator.WithRequiredStructEnabled())
	return API{pgstore.New(pool), logger, validator, pool, mailer}
}

// Confirms a participant on a trip.
// (PATCH /participants/{participantId}/confirm)
func (ap *API) PatchParticipantsParticipantIDConfirm(
	w http.ResponseWriter,
	r *http.Request,
	participantID string,
) *spec.Response {
	id, err := uuid.Parse(participantID)
	if err != nil {
		return spec.PatchParticipantsParticipantIDConfirmJSON400Response(
			spec.Error{Message: "invalid uuid passed: " + err.Error()},
		)
	}

	participant, err := ap.store.GetParticipant(r.Context(), id)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return spec.PatchParticipantsParticipantIDConfirmJSON400Response(
				spec.Error{Message: "trip or participant not found"},
			)
		}
		ap.logger.Error(
			"failed to get participant",
			zap.Error(err),
			zap.String("participant_id", participantID),
		)
		return spec.PatchParticipantsParticipantIDConfirmJSON400Response(
			spec.Error{
				Message: "something went wrong, try again",
			})
	}

	if participant.IsConfirmed {
		return spec.PatchParticipantsParticipantIDConfirmJSON400Response(
			spec.Error{
				Message: "participant already confirmed",
			})
	}
	if err := ap.store.ConfirmParticipant(r.Context(), id); err != nil {
		ap.logger.Error(
			"failed to confim participant",
			zap.Error(err),
			zap.String("participant_id", participantID),
		)
		return spec.PatchParticipantsParticipantIDConfirmJSON400Response(
			spec.Error{
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
		return spec.PostTripsJSON400Response(spec.Error{Message: "invalid JSON: " + err.Error()})
	}

	if err := ap.validator.Struct(body); err != nil {
		return spec.PostTripsJSON400Response(spec.Error{Message: "invalid input: " + err.Error()})
	}

	tripID, err := ap.store.CreateTrip(r.Context(), ap.pool, body)

	if err != nil {
		return spec.PostTripsJSON400Response(spec.Error{Message: "failed to create trip, try again"})
	}

	go func() {
		if err := ap.mailer.SendConfirmTripEmailToTripOwner(tripID); err != nil {
			ap.logger.Error(
				"failed to send email on PostTrips",
				zap.Error(err),
				zap.String("trip_id", tripID.String()),
			)
		}
	}()

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
	id, err := uuid.Parse(tripID)
	if err != nil {
		return spec.GetTripsTripIDActivitiesJSON400Response(
			spec.Error{Message: "invalid uuid passed: " + err.Error()},
		)
	}

	activities, err := ap.store.GetTripActivities(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return spec.GetTripsTripIDActivitiesJSON400Response(
				spec.Error{Message: "trip not found"},
			)
		}
		ap.logger.Error(
			"failed to find trip participants",
			zap.Error(err),
			zap.String("trip_id", tripID),
		)
		return spec.GetTripsTripIDActivitiesJSON400Response(
			spec.Error{Message: "something went wrong, try again"},
		)
	}

	var output spec.GetTripActivitiesResponse

	groupedActivites := make(map[string][]pgstore.Activity)

	for _, act := range activities {
		date := act.OccursAt.Time.Format(time.DateOnly)
		groupedActivites[date] = append(groupedActivites[date], act)
	}

	for dateStr, actsOnDate := range groupedActivites {
		var innerActs []spec.GetTripActivitiesResponseInnerArray

		for _, act := range actsOnDate {
			innerActs = append(innerActs,
				spec.GetTripActivitiesResponseInnerArray{
					ID:       act.ID.String(),
					OccursAt: act.OccursAt.Time,
					Title:    act.Title,
				})
		}

		date, _ := time.Parse(time.DateOnly, dateStr)
		output.Activities = append(output.Activities,
			spec.GetTripActivitiesResponseOuterArray{
				Date:       date,
				Activities: innerActs,
			})
	}

	return spec.GetTripsTripIDActivitiesJSON200Response(output)
}

// Create a trip activity.
// (POST /trips/{tripId}/activities)
func (ap *API) PostTripsTripIDActivities(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	id, err := uuid.Parse(tripID)
	if err != nil {
		return spec.PostTripsTripIDActivitiesJSON400Response(
			spec.Error{Message: "invalid uuid passed: " + err.Error()},
		)
	}

	var body spec.PostTripsTripIDActivitiesJSONBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return spec.PostTripsTripIDActivitiesJSON400Response(
			spec.Error{Message: err.Error()},
		)
	}

	activityID, err := ap.store.CreateActivity(r.Context(),
		pgstore.CreateActivityParams{
			TripID:   id,
			Title:    body.Title,
			OccursAt: pgtype.Timestamp{Time: body.OccursAt, Valid: true},
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return spec.PostTripsTripIDActivitiesJSON400Response(
				spec.Error{Message: "trip not found"},
			)
		}
		ap.logger.Error(
			"failed to find trip participants",
			zap.Error(err),
			zap.String("trip_id", tripID),
		)
		return spec.PostTripsTripIDActivitiesJSON400Response(
			spec.Error{Message: "something went wrong, try again"},
		)
	}

	return spec.PostTripsTripIDActivitiesJSON201Response(
		spec.CreateActivityResponse{ActivityID: activityID.String()},
	)
}

// Confirm a trip and send e-mail invitations.
// (GET /trips/{tripId}/confirm)
func (ap *API) GetTripsTripIDConfirm(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	id, err := uuid.Parse(tripID)
	if err != nil {
		return spec.GetTripsTripIDConfirmJSON400Response(
			spec.Error{Message: "invalid uuid passed: " + err.Error()},
		)
	}

	trip, err := ap.store.GetTrip(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return spec.GetTripsTripIDConfirmJSON400Response(
				spec.Error{Message: "trip not found"},
			)
		}
		ap.logger.Error(
			"failed to get trip by id",
			zap.Error(err),
			zap.String("trip_id", tripID),
		)
		return spec.GetTripsTripIDConfirmJSON400Response(
			spec.Error{Message: "something went wrong, try again"},
		)
	}

	if trip.IsConfirmed {
		return spec.GetTripsTripIDConfirmJSON400Response(
			spec.Error{Message: "trip already confirmed"},
		)
	}

	if err := ap.store.UpdateTrip(r.Context(), pgstore.UpdateTripParams{
		Destination: trip.Destination,
		EndsAt:      trip.EndsAt,
		StartsAt:    trip.StartsAt,
		IsConfirmed: trip.IsConfirmed,
		ID:          id,
	}); err != nil {
		ap.logger.Error(
			"failed to update trip",
			zap.Error(err),
			zap.String("trip_id", tripID),
		)
		return spec.GetTripsTripIDConfirmJSON400Response(
			spec.Error{Message: "something went wrong, try again"},
		)
	}

	go func() {
		if err := ap.mailer.SendTripConfirmedEmails(id); err != nil {
			ap.logger.Error("failed to send trip confirmed email", zap.Error(err))
		}
	}()

	return spec.GetTripsTripIDConfirmJSON204Response(nil)
}

// Invite someone to the trip.
// (POST /trips/{tripId}/invites)
func (ap *API) PostTripsTripIDInvites(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	id, err := uuid.Parse(tripID)
	if err != nil {
		return spec.PostTripsTripIDInvitesJSON400Response(
			spec.Error{Message: "invalid uuid passed: " + err.Error()},
		)
	}

	var body spec.PostTripsTripIDInvitesJSONBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return spec.PostTripsTripIDInvitesJSON400Response(
			spec.Error{Message: err.Error()},
		)
	}

	participantID, err := ap.store.InviteParticipantToTrip(r.Context(), pgstore.InviteParticipantToTripParams{
		TripID: id,
		Email:  string(body.Email),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return spec.PostTripsTripIDInvitesJSON400Response(spec.Error{Message: "trip not found"})
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return spec.PostTripsTripIDInvitesJSON400Response(spec.Error{Message: "participant already invited"})
			}
		}
		ap.logger.Error(
			"failed to invite participant to trip",
			zap.Error(err),
			zap.String("trip_id", tripID),
			zap.String("participant_email", string(body.Email)),
		)
		return spec.PostTripsTripIDInvitesJSON400Response(spec.Error{Message: "something went wrong, try again"})
	}

	go func() {
		if err := ap.mailer.SendTripConfirmedEmail(id, participantID); err != nil {
			ap.logger.Error(
				"failed to send trip confirmed email",
				zap.Error(err),
				zap.String("participant_id", participantID.String()),
				zap.String("trip_id", tripID),
			)
		}
	}()

	return spec.PostTripsTripIDInvitesJSON201Response(nil)

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
	id, err := uuid.Parse(tripID)
	if err != nil {
		return spec.GetTripsTripIDParticipantsJSON400Response(
			spec.Error{Message: "invalid uuid passed: " + err.Error()},
		)
	}

	participants, err := ap.store.GetParticipants(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return spec.GetTripsTripIDParticipantsJSON400Response(
				spec.Error{Message: "trip not found"},
			)
		}
		ap.logger.Error(
			"failed to find trip participants",
			zap.Error(err),
			zap.String("trip_id", tripID),
		)
		return spec.GetTripsTripIDParticipantsJSON400Response(
			spec.Error{Message: "something went wrong, try again"},
		)
	}

	var output spec.GetTripParticipantsResponse

	output.Participants = make([]spec.GetTripParticipantsResponseArray, len(participants))

	for i, p := range participants {
		var name string
		parsedEmail, err := mail.ParseAddress(p.Email)
		if err == nil {
			addr := parsedEmail.Address
			name = addr[:strings.Index(addr, "@")]
		}
		output.Participants[i] = spec.GetTripParticipantsResponseArray{
			Email:       openapi_types.Email(p.Email),
			ID:          p.ID.String(),
			IsConfirmed: p.IsConfirmed,
			Name:        &name,
		}
	}

	return spec.GetTripsTripIDParticipantsJSON200Response(output)
}
