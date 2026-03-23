package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/logitrack/core/internal/model"
	"github.com/logitrack/core/internal/repository"
)

type ShipmentService struct {
	repo         repository.ShipmentRepository
	branchRepo   repository.BranchRepository
	customerRepo repository.CustomerRepository
	commentSvc   *CommentService
}

func NewShipmentService(
	repo repository.ShipmentRepository,
	branchRepo repository.BranchRepository,
	customerRepo repository.CustomerRepository,
	commentSvc *CommentService,
) *ShipmentService {
	return &ShipmentService{repo: repo, branchRepo: branchRepo, customerRepo: customerRepo, commentSvc: commentSvc}
}

func (s *ShipmentService) upsertParties(shipment model.Shipment) {
	if shipment.SenderDNI != "" {
		s.customerRepo.Upsert(model.Customer{
			DNI:     shipment.SenderDNI,
			Name:    shipment.SenderName,
			Phone:   shipment.SenderPhone,
			Email:   shipment.SenderEmail,
			Address: shipment.Origin,
		})
	}
	if shipment.RecipientDNI != "" {
		s.customerRepo.Upsert(model.Customer{
			DNI:     shipment.RecipientDNI,
			Name:    shipment.RecipientName,
			Phone:   shipment.RecipientPhone,
			Email:   shipment.RecipientEmail,
			Address: shipment.Destination,
		})
	}
}

// locationToBranchID converts a city string (from API requests) to a branch ID.
// Falls back to the city string itself if no branch is found.
func (s *ShipmentService) locationToBranchID(city string) string {
	if b, ok := s.branchRepo.GetByCity(city); ok {
		return b.ID
	}
	return city
}

func (s *ShipmentService) Create(req model.CreateShipmentRequest) (model.Shipment, error) {
	if strings.TrimSpace(req.Origin.City) == "" || strings.TrimSpace(req.Origin.Province) == "" {
		return model.Shipment{}, fmt.Errorf("origin city and province are required")
	}
	if strings.TrimSpace(req.Destination.City) == "" || strings.TrimSpace(req.Destination.Province) == "" {
		return model.Shipment{}, fmt.Errorf("destination city and province are required")
	}
	now := time.Now().UTC()
	currentLocation := s.locationToBranchID(req.Origin.City)
	if b, ok := s.branchRepo.GetByID(req.ReceivingBranchID); ok {
		currentLocation = b.ID
	}
	shipment := model.Shipment{
		TrackingID:          generateTrackingID(),
		SenderName:          req.SenderName,
		SenderPhone:         req.SenderPhone,
		SenderEmail:         req.SenderEmail,
		SenderDNI:           req.SenderDNI,
		Origin:              req.Origin,
		RecipientName:       req.RecipientName,
		RecipientPhone:      req.RecipientPhone,
		RecipientEmail:      req.RecipientEmail,
		RecipientDNI:        req.RecipientDNI,
		Destination:         req.Destination,
		WeightKg:            req.WeightKg,
		PackageType:         req.PackageType,
		IsFragile:           req.IsFragile,
		SpecialInstructions: req.SpecialInstructions,
		ReceivingBranchID:   req.ReceivingBranchID,
		Status:              model.StatusInProgress,
		CurrentLocation:     currentLocation,
		CreatedAt:           now,
		UpdatedAt:           now,
		EstimatedDeliveryAt: now.AddDate(0, 0, 7),
	}
	created, err := s.repo.Create(shipment)
	if err != nil {
		return model.Shipment{}, err
	}
	event := model.ShipmentEvent{
		ID:         uuid.NewString(),
		TrackingID: created.TrackingID,
		ToStatus:   model.StatusInProgress,
		ChangedBy:  req.CreatedBy,
		Notes:      "Shipment created",
		Timestamp:  now,
	}
	_ = s.repo.AddEvent(event)
	s.upsertParties(created)
	return created, nil
}

func (s *ShipmentService) SaveDraft(req model.SaveDraftRequest) (model.Shipment, error) {
	now := time.Now().UTC()
	currentLocation := s.locationToBranchID(req.Origin.City)
	if b, ok := s.branchRepo.GetByID(req.ReceivingBranchID); ok {
		currentLocation = b.ID
	}
	shipment := model.Shipment{
		TrackingID:          generateDraftID(),
		SenderName:          req.SenderName,
		SenderPhone:         req.SenderPhone,
		SenderEmail:         req.SenderEmail,
		SenderDNI:           req.SenderDNI,
		Origin:              req.Origin,
		RecipientName:       req.RecipientName,
		RecipientPhone:      req.RecipientPhone,
		RecipientEmail:      req.RecipientEmail,
		RecipientDNI:        req.RecipientDNI,
		Destination:         req.Destination,
		WeightKg:            req.WeightKg,
		PackageType:         req.PackageType,
		IsFragile:           req.IsFragile,
		SpecialInstructions: req.SpecialInstructions,
		ReceivingBranchID:   req.ReceivingBranchID,
		Status:              model.StatusPending,
		CurrentLocation:     currentLocation,
		CreatedAt:           now,
		UpdatedAt:           now,
		EstimatedDeliveryAt: now.AddDate(0, 0, 7),
	}
	created, err := s.repo.Create(shipment)
	if err != nil {
		return model.Shipment{}, err
	}
	return created, nil
}

func (s *ShipmentService) UpdateDraft(draftID string, req model.SaveDraftRequest) (model.Shipment, error) {
	existing, err := s.repo.GetByTrackingID(draftID)
	if err != nil {
		return model.Shipment{}, err
	}
	if existing.Status != model.StatusPending {
		return model.Shipment{}, fmt.Errorf("only draft shipments can be updated")
	}
	existing.SenderName = req.SenderName
	existing.SenderPhone = req.SenderPhone
	existing.SenderEmail = req.SenderEmail
	existing.SenderDNI = req.SenderDNI
	existing.Origin = req.Origin
	existing.RecipientName = req.RecipientName
	existing.RecipientPhone = req.RecipientPhone
	existing.RecipientEmail = req.RecipientEmail
	existing.RecipientDNI = req.RecipientDNI
	existing.Destination = req.Destination
	existing.WeightKg = req.WeightKg
	existing.PackageType = req.PackageType
	existing.IsFragile = req.IsFragile
	existing.SpecialInstructions = req.SpecialInstructions
	existing.ReceivingBranchID = req.ReceivingBranchID
	existing.UpdatedAt = time.Now().UTC()
	// Prefer branch ID derived from receiving branch; fall back to origin city lookup.
	if req.ReceivingBranchID != "" {
		if b, ok := s.branchRepo.GetByID(req.ReceivingBranchID); ok {
			existing.CurrentLocation = b.ID
		}
	} else if req.Origin.City != "" {
		existing.CurrentLocation = s.locationToBranchID(req.Origin.City)
	}
	return s.repo.UpdateDraft(existing)
}

func (s *ShipmentService) ConfirmDraft(draftID string, changedBy string) (model.Shipment, error) {
	draft, err := s.repo.GetByTrackingID(draftID)
	if err != nil {
		return model.Shipment{}, err
	}
	if draft.Status != model.StatusPending {
		return model.Shipment{}, fmt.Errorf("only draft shipments can be confirmed")
	}
	// Validate required fields
	missing := []string{}
	if strings.TrimSpace(draft.SenderName) == "" {
		missing = append(missing, "sender name")
	}
	if strings.TrimSpace(draft.SenderPhone) == "" {
		missing = append(missing, "sender phone")
	}
	if strings.TrimSpace(draft.Origin.City) == "" || strings.TrimSpace(draft.Origin.Province) == "" {
		missing = append(missing, "origin city/province")
	}
	if strings.TrimSpace(draft.RecipientName) == "" {
		missing = append(missing, "recipient name")
	}
	if strings.TrimSpace(draft.RecipientPhone) == "" {
		missing = append(missing, "recipient phone")
	}
	if strings.TrimSpace(draft.Destination.City) == "" || strings.TrimSpace(draft.Destination.Province) == "" {
		missing = append(missing, "destination city/province")
	}
	if draft.WeightKg <= 0 {
		missing = append(missing, "weight")
	}
	if strings.TrimSpace(string(draft.PackageType)) == "" {
		missing = append(missing, "package type")
	}
	if strings.TrimSpace(draft.SenderDNI) == "" {
		missing = append(missing, "sender DNI")
	}
	if strings.TrimSpace(draft.RecipientDNI) == "" {
		missing = append(missing, "recipient DNI")
	}
	if len(missing) > 0 {
		return model.Shipment{}, fmt.Errorf("missing required fields: %s", strings.Join(missing, ", "))
	}
	trackingID := generateTrackingID()
	confirmed, err := s.repo.ConfirmShipment(draftID, trackingID, model.StatusInProgress)
	if err != nil {
		return model.Shipment{}, err
	}
	now := time.Now().UTC()
	from := model.StatusPending
	event := model.ShipmentEvent{
		ID:         uuid.NewString(),
		TrackingID: trackingID,
		FromStatus: &from,
		ToStatus:   model.StatusInProgress,
		ChangedBy:  changedBy,
		Notes:      "Shipment confirmed",
		Timestamp:  now,
	}
	_ = s.repo.AddEvent(event)
	s.upsertParties(confirmed)
	return confirmed, nil
}

func (s *ShipmentService) UpdateStatus(trackingID string, req model.UpdateStatusRequest) (model.Shipment, error) {
	if req.Status == model.StatusDeliveryFailed && strings.TrimSpace(req.Notes) == "" {
		return model.Shipment{}, fmt.Errorf("notes are required for delivery_failed")
	}
	if req.Status == model.StatusDelivering && strings.TrimSpace(req.DriverID) == "" {
		return model.Shipment{}, fmt.Errorf("driver_id is required when moving to delivering")
	}
	current, err := s.repo.GetByTrackingID(trackingID)
	if err != nil {
		return model.Shipment{}, err
	}
	if !isValidTransition(current.Status, req.Status) {
		return model.Shipment{}, fmt.Errorf("invalid transition: %s → %s", current.Status, req.Status)
	}
	// Validate returned: sender DNI must match (corrections take precedence)
	if req.Status == model.StatusReturned {
		if strings.TrimSpace(req.SenderDNI) == "" {
			return model.Shipment{}, fmt.Errorf("sender_dni is required for returned")
		}
		expectedSenderDNI := current.SenderDNI
		if current.Corrections != nil && current.Corrections.SenderDNI != nil {
			expectedSenderDNI = *current.Corrections.SenderDNI
		}
		if expectedSenderDNI != req.SenderDNI {
			return model.Shipment{}, fmt.Errorf("El DNI no coincide con el del remitente esperado")
		}
	}
	// Validate ready_for_return: only allowed when shipment is back at its origin branch.
	// Compares branch IDs directly (CurrentLocation stores branch ID).
	if req.Status == model.StatusReadyForReturn {
		if current.CurrentLocation != current.ReceivingBranchID {
			if b, ok := s.branchRepo.GetByID(current.ReceivingBranchID); ok {
				return model.Shipment{}, fmt.Errorf(
					"el envío no está en la sucursal de origen (%s) — retiro por remitente no aplica en esta sucursal", b.City,
				)
			}
			return model.Shipment{}, fmt.Errorf("el envío no está en la sucursal de origen")
		}
	}
	// Validate DNI before touching the repository (corrections take precedence)
	if req.Status == model.StatusDelivered {
		if strings.TrimSpace(req.RecipientDNI) == "" {
			return model.Shipment{}, fmt.Errorf("recipient_dni is required for delivery")
		}
		expectedRecipientDNI := current.RecipientDNI
		if current.Corrections != nil && current.Corrections.RecipientDNI != nil {
			expectedRecipientDNI = *current.Corrections.RecipientDNI
		}
		if expectedRecipientDNI != req.RecipientDNI {
			return model.Shipment{}, fmt.Errorf("El DNI no coincide con el del destinatario esperado")
		}
	}
	updated, err := s.repo.UpdateStatus(trackingID, req.Status)
	if err != nil {
		return model.Shipment{}, err
	}
	now := time.Now().UTC()
	// When arriving at_branch from in_transit, auto-derive the location from the last in_transit event
	if req.Status == model.StatusAtBranch && current.Status == model.StatusInTransit {
		events, _ := s.repo.GetEvents(trackingID)
		for i := len(events) - 1; i >= 0; i-- {
			if events[i].ToStatus == model.StatusInTransit {
				req.Location = events[i].Location
				break
			}
		}
	}
	// When returning at_branch from delivery_failed, auto-derive the location from the last at_branch event
	if req.Status == model.StatusAtBranch && current.Status == model.StatusDeliveryFailed {
		events, _ := s.repo.GetEvents(trackingID)
		for i := len(events) - 1; i >= 0; i-- {
			if events[i].ToStatus == model.StatusAtBranch {
				req.Location = events[i].Location
				break
			}
		}
	}
	if req.Status != model.StatusDelivered && req.Location != "" {
		locationID := s.locationToBranchID(req.Location)
		_ = s.repo.UpdateLocation(trackingID, locationID)
		updated.CurrentLocation = locationID
	}
	if req.Status == model.StatusDelivered {
		if err := s.repo.SetDeliveredAt(trackingID, now); err != nil {
			return model.Shipment{}, err
		}
		updated.DeliveredAt = &now
	}
	from := current.Status
	event := model.ShipmentEvent{
		ID:         uuid.NewString(),
		TrackingID: trackingID,
		FromStatus: &from,
		ToStatus:   req.Status,
		ChangedBy:  req.ChangedBy,
		Location:   req.Location,
		Notes:      req.Notes,
		Timestamp:  now,
	}
	_ = s.repo.AddEvent(event)
	return updated, nil
}

// CorrectShipment stores field corrections without modifying original data and
// auto-persists one comment per corrected field.
func (s *ShipmentService) CorrectShipment(trackingID, username string, req model.CorrectShipmentRequest) (model.Shipment, error) {
	if req.Corrections.IsEmpty() {
		return model.Shipment{}, fmt.Errorf("no corrections provided")
	}
	shipment, err := s.repo.GetByTrackingID(trackingID)
	if err != nil {
		return model.Shipment{}, err
	}
	if shipment.Status == model.StatusPending {
		return model.Shipment{}, fmt.Errorf("los borradores se editan directamente")
	}
	if shipment.Status == model.StatusDelivered || shipment.Status == model.StatusReturned || shipment.Status == model.StatusCancelled {
		return model.Shipment{}, fmt.Errorf("no se pueden corregir envíos finalizados")
	}
	updated, err := s.repo.ApplyCorrections(trackingID, req.Corrections)
	if err != nil {
		return model.Shipment{}, err
	}
	now := time.Now().UTC()
	from := shipment.Status
	event := model.ShipmentEvent{
		ID:         uuid.NewString(),
		TrackingID: trackingID,
		EventType:  "edited",
		FromStatus: &from,
		ToStatus:   shipment.Status,
		ChangedBy:  username,
		Notes:      fmt.Sprintf("Corrección de datos: %d campo(s) modificado(s)", len(req.Corrections.Fields())),
		Timestamp:  now,
	}
	_ = s.repo.AddEvent(event)
	for _, f := range req.Corrections.Fields() {
		body := fmt.Sprintf("[Corrección] %s. Nuevo valor: %s", f.Label, f.Value)
		_, _ = s.commentSvc.AddComment(trackingID, username, body)
	}
	return updated, nil
}

func (s *ShipmentService) CancelShipment(trackingID, username, reason string) (model.Shipment, error) {
	if strings.TrimSpace(reason) == "" {
		return model.Shipment{}, fmt.Errorf("el motivo de cancelación es obligatorio")
	}
	shipment, err := s.repo.GetByTrackingID(trackingID)
	if err != nil {
		return model.Shipment{}, err
	}
	nonCancellable := map[model.Status]bool{
		model.StatusPending:   true,
		model.StatusDelivered: true,
		model.StatusReturned:  true,
		model.StatusCancelled: true,
	}
	if nonCancellable[shipment.Status] {
		return model.Shipment{}, fmt.Errorf("no se puede cancelar un envío en estado '%s'", shipment.Status)
	}
	updated, err := s.repo.UpdateStatus(trackingID, model.StatusCancelled)
	if err != nil {
		return model.Shipment{}, err
	}
	now := time.Now().UTC()
	from := shipment.Status
	event := model.ShipmentEvent{
		ID:         uuid.NewString(),
		TrackingID: trackingID,
		FromStatus: &from,
		ToStatus:   model.StatusCancelled,
		ChangedBy:  username,
		Notes:      reason,
		Timestamp:  now,
	}
	_ = s.repo.AddEvent(event)
	body := fmt.Sprintf("[Cancelación] %s", reason)
	_, _ = s.commentSvc.AddComment(trackingID, username, body)
	return updated, nil
}

func (s *ShipmentService) GetByTrackingID(trackingID string) (model.Shipment, error) {
	return s.repo.GetByTrackingID(trackingID)
}

func (s *ShipmentService) List(filter model.ShipmentFilter) ([]model.Shipment, error) {
	return s.repo.List(filter)
}

func (s *ShipmentService) Search(query string) ([]model.Shipment, error) {
	if strings.TrimSpace(query) == "" {
		return s.repo.List(model.ShipmentFilter{})
	}
	return s.repo.Search(query)
}

func (s *ShipmentService) GetEvents(trackingID string) ([]model.ShipmentEvent, error) {
	return s.repo.GetEvents(trackingID)
}

func (s *ShipmentService) Stats() (model.Stats, error) {
	return s.repo.Stats()
}

func generateTrackingID() string {
	id := uuid.New().String()
	return fmt.Sprintf("LT-%s", strings.ToUpper(id[:8]))
}

func generateDraftID() string {
	id := uuid.New().String()
	return fmt.Sprintf("DRAFT-%s", strings.ToUpper(id[:8]))
}

func isValidTransition(from, to model.Status) bool {
	allowed := map[model.Status][]model.Status{
		model.StatusPending:        {},                      // drafts transition via ConfirmDraft, not UpdateStatus
		model.StatusInProgress:     {model.StatusInTransit}, // confirmed → start transit
		model.StatusInTransit:      {model.StatusAtBranch},
		model.StatusAtBranch:       {model.StatusInTransit, model.StatusDelivering, model.StatusReadyForPickup, model.StatusReadyForReturn},
		model.StatusDelivering:     {model.StatusDelivered, model.StatusDeliveryFailed},
		model.StatusDeliveryFailed: {model.StatusDelivering, model.StatusAtBranch},
		model.StatusDelivered:      {},
		model.StatusReadyForPickup: {model.StatusDelivered, model.StatusInTransit},
		model.StatusReadyForReturn: {model.StatusReturned},
		model.StatusReturned:       {},
		model.StatusCancelled:      {},
	}
	for _, s := range allowed[from] {
		if s == to {
			return true
		}
	}
	return false
}
