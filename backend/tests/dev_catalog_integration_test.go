//go:build integration

package tests

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	eventsdomain "reserveflow/backend/internal/modules/events/domain"
)

func TestPublicCatalogOnlyReturnsBookableKudaGoEvents(t *testing.T) {
	ctx := context.Background()
	server, deps, cleanup := newIntegrationServerWithDeps(t, ctx)
	defer cleanup()

	publicEvent := seedBookableKudaGoEvent(t, ctx, deps.DB, "Public KudaGo")
	seedEventFixture(t, ctx, deps.DB, seedEventOptions{
		Source:             eventsdomain.SourceManual,
		BookingMode:        eventsdomain.BookingModeReserveFlowManaged,
		Bookable:           true,
		IncludeCoordinates: true,
		Layout:             layoutPtr(sampleStoredLayout(1, 2)),
		Title:              "Manual event",
	})
	seedEventFixture(t, ctx, deps.DB, seedEventOptions{
		Source:             eventsdomain.SourceTimepad,
		BookingMode:        eventsdomain.BookingModeReserveFlowManaged,
		Bookable:           true,
		IncludeCoordinates: true,
		Layout:             layoutPtr(sampleStoredLayout(1, 2)),
		Title:              "Timepad event",
	})
	hiddenEvent := seedEventFixture(t, ctx, deps.DB, seedEventOptions{
		Source:             eventsdomain.SourceKudaGo,
		BookingMode:        eventsdomain.BookingModeExternalLinkOnly,
		Bookable:           false,
		IncludeCoordinates: true,
		Title:              "Hidden KudaGo",
	})

	var events listResponse[eventCatalogResponse]
	doJSON(t, http.MethodGet, server.URL+"/api/v1/events", "", nil, http.StatusOK, &events)
	require.Len(t, events.Items, 1)
	require.Equal(t, publicEvent.EventID, events.Items[0].ID)
	require.Equal(t, "kudago", events.Items[0].Source)
	require.Equal(t, "reserveflow_managed", events.Items[0].BookingMode)

	var filtered listResponse[eventCatalogResponse]
	doJSON(t, http.MethodGet, server.URL+"/api/v1/events?source=manual", "", nil, http.StatusOK, &filtered)
	require.Empty(t, filtered.Items)

	var bookable listResponse[eventCatalogResponse]
	doJSON(t, http.MethodGet, server.URL+"/api/v1/events?bookingMode=bookable", "", nil, http.StatusOK, &bookable)
	require.Len(t, bookable.Items, 1)
	require.Equal(t, publicEvent.EventID, bookable.Items[0].ID)

	var mapResponse struct {
		Events []eventMapResponse `json:"events"`
	}
	doJSON(t, http.MethodGet, server.URL+"/api/v1/events/map", "", nil, http.StatusOK, &mapResponse)
	require.Len(t, mapResponse.Events, 1)
	require.Equal(t, publicEvent.EventID, mapResponse.Events[0].ID)

	var detail eventDetailResponse
	doJSON(t, http.MethodGet, server.URL+"/api/v1/events/"+hiddenEvent.EventID, "", nil, http.StatusOK, &detail)
	require.Equal(t, hiddenEvent.EventID, detail.ID)
	require.Equal(t, "external_link_only", detail.BookingMode)
	require.Len(t, detail.Sessions, 1)
	require.False(t, detail.Sessions[0].IsBookable)
}
