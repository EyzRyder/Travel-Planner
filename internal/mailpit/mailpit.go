package mailpit

import (
	"context"
	"fmt"
	"time"

    "github.com/EyzRyder/Travel-Planner/internal/pgstore"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/wneessen/go-mail"
)

type Store interface {
    GetTrip(context.Context, uuid.UUID) (pgstore.Trip, error)
    GetParticipant(ctx context.Context, participantID uuid.UUID) (pgstore.Participant, error)
	GetParticipants(ctx context.Context, tripID uuid.UUID) ([]pgstore.Participant, error)
}

type Mailpit struct{
    store Store
}

func NewMailpit(pool *pgxpool.Pool) Mailpit{
    return Mailpit{store: pgstore.New(pool)}
}

func (mp Mailpit) SendConfirmTripEmailToTripOwner(tripID uuid.UUID) error {
    ctx := context.Background()
    trip, err := mp.store.GetTrip(ctx,tripID)
    if err != nil {
        return fmt.Errorf("mailpit: failed to get trip for SendConfirmTripEmailToTripOwner: %w", err)
    }

    msg := mail.NewMsg()

	if err := msg.From("mailpit@journey.com"); err != nil {
		return fmt.Errorf("mailpit: failed to set 'From' in email SendConfirmTripEmailToTripOwner: %w", err)
	}

	if err := msg.To(trip.OwnerEmail); err != nil {
		return fmt.Errorf("mailpit: failed to set 'to' in email SendConfirmTripEmailToTripOwner: %w", err)
	}

	msg.Subject("Confirme sua viagem")
	msg.SetBodyString(mail.TypeTextPlain, fmt.Sprintf(`
		Olá, %s!

		A sua viagem para %s que começa no dia %s precisa ser confirmada.
		Clique no botão abaixo para confirmar.
		`,
		trip.OwnerName, trip.Destination, trip.StartsAt.Time.Format(time.DateOnly),
	))

	client, err := mail.NewClient("mailpit", mail.WithTLSPortPolicy(mail.NoTLS), mail.WithPort(1025))
	if err != nil {
		return fmt.Errorf("mailpit: failed create email client SendConfirmTripEmailToTripOwner: %w", err)
	}

	if err := client.DialAndSend(msg); err != nil {
		return fmt.Errorf("mailpit: failed send email client SendConfirmTripEmailToTripOwner: %w", err)
	}

    return nil
}

func (mp Mailpit) SendTripConfirmedEmails(tripID uuid.UUID) error {
    participants, err := mp.store.GetParticipants(context.Background(), tripID)
	if err != nil {
		return err
	}

	c, err := mail.NewClient("mailpit", mail.WithTLSPortPolicy(mail.NoTLS), mail.WithPort(1025))
	if err != nil {
		return err
	}

	for _, p := range participants {
		msg := mail.NewMsg()
		if err := msg.From("mailpit@journey.com"); err != nil {
			return err
		}

		if err := msg.To(p.Email); err != nil {
			return err
		}

		msg.Subject("Confirme sua viagem")
		msg.SetBodyString(mail.TypeTextPlain, "Você deve confirmar sua viagem")

		if err := c.DialAndSend(msg); err != nil {
			return err
		}
	}

	return nil
}

func (mp Mailpit) SendTripConfirmedEmail(tripID, participantID uuid.UUID) error {
	ctx := context.Background()
	participant, err := mp.store.GetParticipant(ctx, participantID)
	if err != nil {
		return err
	}

	msg := mail.NewMsg()
	if err := msg.From("mailpit@journey.com"); err != nil {
		return err
	}

	if err := msg.To(participant.Email); err != nil {
		return err
	}

	msg.Subject("Confirme sua viagem")
	msg.SetBodyString(mail.TypeTextPlain, "Você deve confirmar sua viagem")

	c, err := mail.NewClient("mailpit", mail.WithTLSPortPolicy(mail.NoTLS), mail.WithPort(1025))
	if err != nil {
		return err
	}

	if err := c.DialAndSend(msg); err != nil {
		return err
	}

	return nil
}
