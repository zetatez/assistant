package gcal

import (
	"context"
	"time"

	calendar "google.golang.org/api/calendar/v3"
)

func (c *Client) CreateEvent(ctx context.Context, e Event) (Event, error) {
	ev := &calendar.Event{
		Summary:     e.Title,
		Description: e.Description,
		Location:    e.Location,
		Start: &calendar.EventDateTime{
			DateTime: e.Start.Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: e.End.Format(time.RFC3339),
		},
	}

	res, err := c.svc.Events.Insert("primary", ev).Context(ctx).Do()
	if err != nil {
		return Event{}, err
	}

	return Event{
		ID:    res.Id,
		Title: res.Summary,
	}, nil
}

func (c *Client) DeleteEvent(ctx context.Context, id string) error {
	return c.svc.Events.Delete("primary", id).Context(ctx).Do()
}

func (c *Client) GetEvent(ctx context.Context, id string) (Event, error) {
	e, err := c.svc.Events.Get("primary", id).Context(ctx).Do()
	if err != nil {
		return Event{}, err
	}

	return Event{
		ID:    e.Id,
		Title: e.Summary,
	}, nil
}

type EventQuery struct {
	Start time.Time
	End   time.Time
	Query string
}

func (c *Client) ListEvents(ctx context.Context, q EventQuery) ([]Event, error) {
	call := c.svc.Events.List("primary").
		SingleEvents(true).
		OrderBy("startTime").
		Context(ctx)

	if !q.Start.IsZero() {
		call = call.TimeMin(q.Start.Format(time.RFC3339))
	}
	if !q.End.IsZero() {
		call = call.TimeMax(q.End.Format(time.RFC3339))
	}
	if q.Query != "" {
		call = call.Q(q.Query)
	}

	resp, err := call.Do()
	if err != nil {
		return nil, err
	}

	var out []Event
	for _, e := range resp.Items {
		out = append(out, Event{
			ID:    e.Id,
			Title: e.Summary,
		})
	}
	return out, nil
}

func (c *Client) UpdateEvent(
	ctx context.Context,
	eventID string,
	event *calendar.Event,
) (*calendar.Event, error) {
	return c.svc.Events.
		Update("primary", eventID, event).
		Context(ctx).
		Do()
}

func (c *Client) PatchEvent(
	ctx context.Context,
	eventID string,
	updates *calendar.Event,
) (*calendar.Event, error) {
	return c.svc.Events.
		Patch("primary", eventID, updates).
		Context(ctx).
		Do()
}
