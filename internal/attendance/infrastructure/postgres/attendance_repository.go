package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"erp/internal/attendance/domain"
	"erp/internal/attendance/ports"
	"erp/internal/platform/database"
	"erp/internal/platform/database/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type attendanceRepository struct {
	db *database.DB
	q  *sqlc.Queries
}

func NewAttendanceRepository(db *database.DB) ports.AttendanceRepository {
	return &attendanceRepository{
		db: db,
		q:  sqlc.New(db.Pool),
	}
}

// ============================================================================
// SHIFTS
// ============================================================================

func (r *attendanceRepository) CreateShift(ctx context.Context, s *domain.Shift) (*domain.Shift, error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	workDaysJSON, err := json.Marshal(s.WorkDays)
	if err != nil {
		workDaysJSON = []byte(`["SUNDAY","MONDAY","TUESDAY","WEDNESDAY","THURSDAY"]`)
	}

	row, err := r.q.CreateShift(ctx, sqlc.CreateShiftParams{
		ID:                     toPgUUID(s.ID),
		LegalEntityID:          toPgUUID(s.LegalEntityID),
		Code:                   s.Code,
		NameEn:                 s.NameEn,
		NameAr:                 s.NameAr,
		StartTime:              toPgTime(s.StartTime),
		EndTime:                toPgTime(s.EndTime),
		GracePeriodMinutes:     s.GracePeriodMinutes,
		BreakDurationMinutes:   s.BreakDurationMinutes,
		HalfDayThresholdMinutes: s.HalfDayThresholdMinutes,
		IsNightShift:           s.IsNightShift,
		WorkDays:               workDaysJSON,
		IsActive:               s.IsActive,
	})
	if err != nil {
		return nil, err
	}
	return mapShift(row), nil
}

func (r *attendanceRepository) GetShiftByID(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Shift, error) {
	row, err := r.q.GetShiftByID(ctx, sqlc.GetShiftByIDParams{
		ID:            toPgUUID(id),
		LegalEntityID: toPgUUID(legalEntityID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrShiftNotFound
		}
		return nil, err
	}
	return mapShift(row), nil
}

func (r *attendanceRepository) GetShiftByCode(ctx context.Context, legalEntityID uuid.UUID, code string) (*domain.Shift, error) {
	row, err := r.q.GetShiftByCode(ctx, sqlc.GetShiftByCodeParams{
		LegalEntityID: toPgUUID(legalEntityID),
		Code:          code,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrShiftNotFound
		}
		return nil, err
	}
	return mapShift(row), nil
}

func (r *attendanceRepository) ListShifts(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.Shift, error) {
	rows, err := r.q.ListShifts(ctx, toPgUUID(legalEntityID))
	if err != nil {
		return nil, err
	}
	res := make([]*domain.Shift, len(rows))
	for i, row := range rows {
		res[i] = mapShift(row)
	}
	return res, nil
}

// ============================================================================
// SHIFT ASSIGNMENTS
// ============================================================================

func (r *attendanceRepository) CreateShiftAssignment(ctx context.Context, a *domain.ShiftAssignment) (*domain.ShiftAssignment, error) {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	row, err := r.q.CreateShiftAssignment(ctx, sqlc.CreateShiftAssignmentParams{
		ID:         toPgUUID(a.ID),
		EmployeeID: toPgUUID(a.EmployeeID),
		ShiftID:    toPgUUID(a.ShiftID),
		ValidFrom:  toPgDate(a.ValidFrom),
		ValidTo:    toPgDatePtr(a.ValidTo),
		IsActive:   a.IsActive,
	})
	if err != nil {
		return nil, err
	}
	return mapShiftAssignment(row), nil
}

func (r *attendanceRepository) GetActiveShiftAssignment(ctx context.Context, employeeID uuid.UUID) (*domain.ShiftAssignment, error) {
	row, err := r.q.GetActiveShiftAssignment(ctx, toPgUUID(employeeID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	shiftNameEn := row.ShiftNameEn
	shiftNameAr := row.ShiftNameAr
	return &domain.ShiftAssignment{
		ID:          toUUID(row.ID),
		EmployeeID:  toUUID(row.EmployeeID),
		ShiftID:     toUUID(row.ShiftID),
		ShiftNameEn: &shiftNameEn,
		ShiftNameAr: &shiftNameAr,
		ValidFrom:   fromPgDate(row.ValidFrom),
		ValidTo:     fromPgDatePtr(row.ValidTo),
		IsActive:    row.IsActive,
		CreatedAt:   fromPgTimestamptz(row.CreatedAt),
	}, nil
}

func (r *attendanceRepository) ListShiftAssignments(ctx context.Context, employeeID uuid.UUID) ([]*domain.ShiftAssignment, error) {
	rows, err := r.q.ListShiftAssignments(ctx, toPgUUID(employeeID))
	if err != nil {
		return nil, err
	}
	res := make([]*domain.ShiftAssignment, len(rows))
	for i, row := range rows {
		code := row.ShiftCode
		nameEn := row.ShiftNameEn
		nameAr := row.ShiftNameAr
		res[i] = &domain.ShiftAssignment{
			ID:          toUUID(row.ID),
			EmployeeID:  toUUID(row.EmployeeID),
			ShiftID:     toUUID(row.ShiftID),
			ShiftCode:   &code,
			ShiftNameEn: &nameEn,
			ShiftNameAr: &nameAr,
			ValidFrom:   fromPgDate(row.ValidFrom),
			ValidTo:     fromPgDatePtr(row.ValidTo),
			IsActive:    row.IsActive,
			CreatedAt:   fromPgTimestamptz(row.CreatedAt),
		}
	}
	return res, nil
}

// ============================================================================
// ATTENDANCE RECORDS
// ============================================================================

func (r *attendanceRepository) CreateAttendanceRecord(ctx context.Context, rec *domain.AttendanceRecord) (*domain.AttendanceRecord, error) {
	if rec.ID == uuid.Nil {
		rec.ID = uuid.New()
	}
	var checkOutMethod *string
	if rec.CheckOutMethod != nil {
		s := string(*rec.CheckOutMethod)
		checkOutMethod = &s
	}

	row, err := r.q.CreateAttendanceRecord(ctx, sqlc.CreateAttendanceRecordParams{
		ID:                    toPgUUID(rec.ID),
		LegalEntityID:         toPgUUID(rec.LegalEntityID),
		EmployeeID:            toPgUUID(rec.EmployeeID),
		ShiftID:               toPgUUIDPtr(rec.ShiftID),
		WorkDate:              toPgDate(rec.WorkDate),
		ClockIn:               toPgTimestamptzPtr(rec.ClockIn),
		ClockOut:              toPgTimestamptzPtr(rec.ClockOut),
		CheckInMethod:         string(rec.CheckInMethod),
		CheckOutMethod:        toPgTextPtr(checkOutMethod),
		ExpectedWorkHours:     toPgNumeric(rec.ExpectedWorkHours),
		ActualWorkHours:       toPgNumeric(rec.ActualWorkHours),
		LateMinutes:           rec.LateMinutes,
		EarlyDepartureMinutes: rec.EarlyDepartureMinutes,
		OvertimeMinutes:       rec.OvertimeMinutes,
		Status:                string(rec.Status),
		Notes:                 toPgTextPtr(rec.Notes),
	})
	if err != nil {
		return nil, err
	}
	return mapAttendanceRecord(row), nil
}

func (r *attendanceRepository) GetAttendanceRecordByID(ctx context.Context, id uuid.UUID) (*domain.AttendanceRecord, error) {
	row, err := r.q.GetAttendanceRecordByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrRecordNotFound
		}
		return nil, err
	}
	return mapAttendanceRecord(row), nil
}

func (r *attendanceRepository) GetAttendanceRecordByEmployeeAndDate(ctx context.Context, employeeID uuid.UUID, date time.Time) (*domain.AttendanceRecord, error) {
	row, err := r.q.GetAttendanceRecordByEmployeeAndDate(ctx, sqlc.GetAttendanceRecordByEmployeeAndDateParams{
		EmployeeID: toPgUUID(employeeID),
		WorkDate:   toPgDate(date),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return mapAttendanceRecord(row), nil
}

func (r *attendanceRepository) UpdateAttendanceRecordClockOut(ctx context.Context, rec *domain.AttendanceRecord) (*domain.AttendanceRecord, error) {
	var checkOutMethod *string
	if rec.CheckOutMethod != nil {
		s := string(*rec.CheckOutMethod)
		checkOutMethod = &s
	}

	row, err := r.q.UpdateAttendanceRecordClockOut(ctx, sqlc.UpdateAttendanceRecordClockOutParams{
		ID:                    toPgUUID(rec.ID),
		LegalEntityID:         toPgUUID(rec.LegalEntityID),
		ClockOut:              toPgTimestamptzPtr(rec.ClockOut),
		CheckOutMethod:        toPgTextPtr(checkOutMethod),
		ActualWorkHours:       toPgNumeric(rec.ActualWorkHours),
		EarlyDepartureMinutes: rec.EarlyDepartureMinutes,
		OvertimeMinutes:       rec.OvertimeMinutes,
		Status:                string(rec.Status),
		Notes:                 toPgTextPtr(rec.Notes),
	})
	if err != nil {
		return nil, err
	}
	return mapAttendanceRecord(row), nil
}

func (r *attendanceRepository) ListAttendanceRecordsByDate(ctx context.Context, legalEntityID uuid.UUID, date time.Time) ([]*domain.AttendanceRecord, error) {
	rows, err := r.q.ListAttendanceRecordsByDate(ctx, sqlc.ListAttendanceRecordsByDateParams{
		LegalEntityID: toPgUUID(legalEntityID),
		WorkDate:      toPgDate(date),
	})
	if err != nil {
		return nil, err
	}
	res := make([]*domain.AttendanceRecord, len(rows))
	for i, row := range rows {
		empNum := row.EmployeeNumber
		nameEn := row.FirstNameEn + " " + row.LastNameEn
		nameAr := row.FirstNameAr + " " + row.LastNameAr
		rec := mapAttendanceRecord(sqlc.AttendanceRecords{
			ID:                    row.ID,
			LegalEntityID:         row.LegalEntityID,
			EmployeeID:            row.EmployeeID,
			ShiftID:               row.ShiftID,
			WorkDate:              row.WorkDate,
			ClockIn:               row.ClockIn,
			ClockOut:              row.ClockOut,
			CheckInMethod:         row.CheckInMethod,
			CheckOutMethod:        row.CheckOutMethod,
			ExpectedWorkHours:     row.ExpectedWorkHours,
			ActualWorkHours:       row.ActualWorkHours,
			LateMinutes:           row.LateMinutes,
			EarlyDepartureMinutes: row.EarlyDepartureMinutes,
			OvertimeMinutes:       row.OvertimeMinutes,
			Status:                row.Status,
			Notes:                 row.Notes,
			CreatedAt:             row.CreatedAt,
			UpdatedAt:             row.UpdatedAt,
		})
		rec.EmployeeNumber = &empNum
		rec.EmployeeNameEn = &nameEn
		rec.EmployeeNameAr = &nameAr
		res[i] = rec
	}
	return res, nil
}

func (r *attendanceRepository) ListEmployeeAttendanceRecords(ctx context.Context, employeeID uuid.UUID, from, to time.Time) ([]*domain.AttendanceRecord, error) {
	rows, err := r.q.ListEmployeeAttendanceRecords(ctx, sqlc.ListEmployeeAttendanceRecordsParams{
		EmployeeID: toPgUUID(employeeID),
		WorkDate:   toPgDate(from),
		WorkDate_2: toPgDate(to),
	})
	if err != nil {
		return nil, err
	}
	res := make([]*domain.AttendanceRecord, len(rows))
	for i, row := range rows {
		res[i] = mapAttendanceRecord(row)
	}
	return res, nil
}

// ============================================================================
// OVERTIME REQUESTS
// ============================================================================

func (r *attendanceRepository) CreateOvertimeRequest(ctx context.Context, o *domain.OvertimeRequest) (*domain.OvertimeRequest, error) {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	row, err := r.q.CreateOvertimeRequest(ctx, sqlc.CreateOvertimeRequestParams{
		ID:              toPgUUID(o.ID),
		LegalEntityID:   toPgUUID(o.LegalEntityID),
		EmployeeID:      toPgUUID(o.EmployeeID),
		RequestDate:     toPgDate(o.RequestDate),
		OvertimeDate:    toPgDate(o.OvertimeDate),
		OvertimeHours:   toPgNumeric(o.OvertimeHours),
		RateMultiplier:  toPgNumeric(o.RateMultiplier),
		Reason:          o.Reason,
		Status:          string(o.Status),
	})
	if err != nil {
		return nil, err
	}
	return mapOvertimeRequest(row), nil
}

func (r *attendanceRepository) GetOvertimeRequestByID(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.OvertimeRequest, error) {
	row, err := r.q.GetOvertimeRequestByID(ctx, sqlc.GetOvertimeRequestByIDParams{
		ID:            toPgUUID(id),
		LegalEntityID: toPgUUID(legalEntityID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrOvertimeNotFound
		}
		return nil, err
	}
	return mapOvertimeRequest(row), nil
}

func (r *attendanceRepository) ListOvertimeRequests(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.OvertimeRequest, error) {
	rows, err := r.q.ListOvertimeRequests(ctx, toPgUUID(legalEntityID))
	if err != nil {
		return nil, err
	}
	res := make([]*domain.OvertimeRequest, len(rows))
	for i, row := range rows {
		empNum := row.EmployeeNumber
		nameEn := row.FirstNameEn + " " + row.LastNameEn
		nameAr := row.FirstNameAr + " " + row.LastNameAr
		ot := mapOvertimeRequest(sqlc.AttendanceOvertimeRequests{
			ID:               row.ID,
			LegalEntityID:    row.LegalEntityID,
			EmployeeID:       row.EmployeeID,
			RequestDate:      row.RequestDate,
			OvertimeDate:     row.OvertimeDate,
			OvertimeHours:    row.OvertimeHours,
			RateMultiplier:   row.RateMultiplier,
			Reason:           row.Reason,
			Status:           row.Status,
			ApprovedByUserID: row.ApprovedByUserID,
			ApprovedAt:       row.ApprovedAt,
			CreatedAt:        row.CreatedAt,
			UpdatedAt:        row.UpdatedAt,
		})
		ot.EmployeeNumber = &empNum
		ot.EmployeeNameEn = &nameEn
		ot.EmployeeNameAr = &nameAr
		res[i] = ot
	}
	return res, nil
}

func (r *attendanceRepository) UpdateOvertimeStatus(ctx context.Context, legalEntityID, id uuid.UUID, status domain.RequestStatus, approvedBy uuid.UUID) (*domain.OvertimeRequest, error) {
	row, err := r.q.UpdateOvertimeStatus(ctx, sqlc.UpdateOvertimeStatusParams{
		ID:               toPgUUID(id),
		LegalEntityID:    toPgUUID(legalEntityID),
		Status:           string(status),
		ApprovedByUserID: toPgUUIDPtr(&approvedBy),
	})
	if err != nil {
		return nil, err
	}
	return mapOvertimeRequest(row), nil
}

// ============================================================================
// ADJUSTMENTS
// ============================================================================

func (r *attendanceRepository) CreateAttendanceAdjustment(ctx context.Context, a *domain.AttendanceAdjustment) (*domain.AttendanceAdjustment, error) {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	row, err := r.q.CreateAttendanceAdjustment(ctx, sqlc.CreateAttendanceAdjustmentParams{
		ID:                    toPgUUID(a.ID),
		AttendanceRecordID:    toPgUUID(a.AttendanceRecordID),
		RequestedByEmployeeID: toPgUUID(a.RequestedByEmployeeID),
		OriginalClockIn:       toPgTimestamptzPtr(a.OriginalClockIn),
		OriginalClockOut:      toPgTimestamptzPtr(a.OriginalClockOut),
		AdjustedClockIn:       toPgTimestamptz(a.AdjustedClockIn),
		AdjustedClockOut:      toPgTimestamptz(a.AdjustedClockOut),
		Reason:                a.Reason,
		Status:                string(a.Status),
	})
	if err != nil {
		return nil, err
	}
	return mapAdjustment(row), nil
}

func (r *attendanceRepository) GetAttendanceAdjustmentByID(ctx context.Context, id uuid.UUID) (*domain.AttendanceAdjustment, error) {
	row, err := r.q.GetAttendanceAdjustmentByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrAdjustmentNotFound
		}
		return nil, err
	}
	return mapAdjustment(row), nil
}

func (r *attendanceRepository) UpdateAdjustmentStatus(ctx context.Context, id uuid.UUID, status domain.RequestStatus, reviewedBy uuid.UUID, reviewNotes string) (*domain.AttendanceAdjustment, error) {
	row, err := r.q.UpdateAdjustmentStatus(ctx, sqlc.UpdateAdjustmentStatusParams{
		ID:                 toPgUUID(id),
		Status:             string(status),
		ReviewedByUserID:   toPgUUIDPtr(&reviewedBy),
		ReviewNotes:        toPgTextPtr(&reviewNotes),
	})
	if err != nil {
		return nil, err
	}
	return mapAdjustment(row), nil
}

// ============================================================================
// MAPPERS
// ============================================================================

func mapShift(r sqlc.AttendanceShifts) *domain.Shift {
	var workDays []string
	_ = json.Unmarshal(r.WorkDays, &workDays)
	return &domain.Shift{
		ID:                     toUUID(r.ID),
		LegalEntityID:          toUUID(r.LegalEntityID),
		Code:                   r.Code,
		NameEn:                 r.NameEn,
		NameAr:                 r.NameAr,
		StartTime:              fromPgTime(r.StartTime),
		EndTime:                fromPgTime(r.EndTime),
		GracePeriodMinutes:     r.GracePeriodMinutes,
		BreakDurationMinutes:   r.BreakDurationMinutes,
		HalfDayThresholdMinutes: r.HalfDayThresholdMinutes,
		IsNightShift:           r.IsNightShift,
		WorkDays:               workDays,
		IsActive:               r.IsActive,
		CreatedAt:              fromPgTimestamptz(r.CreatedAt),
		UpdatedAt:              fromPgTimestamptz(r.UpdatedAt),
	}
}

func mapShiftAssignment(r sqlc.AttendanceShiftAssignments) *domain.ShiftAssignment {
	return &domain.ShiftAssignment{
		ID:         toUUID(r.ID),
		EmployeeID: toUUID(r.EmployeeID),
		ShiftID:    toUUID(r.ShiftID),
		ValidFrom:  fromPgDate(r.ValidFrom),
		ValidTo:    fromPgDatePtr(r.ValidTo),
		IsActive:   r.IsActive,
		CreatedAt:  fromPgTimestamptz(r.CreatedAt),
	}
}

func mapAttendanceRecord(r sqlc.AttendanceRecords) *domain.AttendanceRecord {
	var checkOutMethod *domain.CheckMethod
	if r.CheckOutMethod.Valid {
		m := domain.CheckMethod(r.CheckOutMethod.String)
		checkOutMethod = &m
	}

	return &domain.AttendanceRecord{
		ID:                    toUUID(r.ID),
		LegalEntityID:         toUUID(r.LegalEntityID),
		EmployeeID:            toUUID(r.EmployeeID),
		ShiftID:               toUUIDPtr(r.ShiftID),
		WorkDate:              fromPgDate(r.WorkDate),
		ClockIn:               fromPgTimestamptzPtr(r.ClockIn),
		ClockOut:              fromPgTimestamptzPtr(r.ClockOut),
		CheckInMethod:         domain.CheckMethod(r.CheckInMethod),
		CheckOutMethod:        checkOutMethod,
		ExpectedWorkHours:     fromPgNumeric(r.ExpectedWorkHours),
		ActualWorkHours:       fromPgNumeric(r.ActualWorkHours),
		LateMinutes:           r.LateMinutes,
		EarlyDepartureMinutes: r.EarlyDepartureMinutes,
		OvertimeMinutes:       r.OvertimeMinutes,
		Status:                domain.AttendanceStatus(r.Status),
		Notes:                 fromPgTextPtr(r.Notes),
		CreatedAt:             fromPgTimestamptz(r.CreatedAt),
		UpdatedAt:             fromPgTimestamptz(r.UpdatedAt),
	}
}

func mapOvertimeRequest(r sqlc.AttendanceOvertimeRequests) *domain.OvertimeRequest {
	return &domain.OvertimeRequest{
		ID:               toUUID(r.ID),
		LegalEntityID:    toUUID(r.LegalEntityID),
		EmployeeID:       toUUID(r.EmployeeID),
		RequestDate:      fromPgDate(r.RequestDate),
		OvertimeDate:     fromPgDate(r.OvertimeDate),
		OvertimeHours:    fromPgNumeric(r.OvertimeHours),
		RateMultiplier:   fromPgNumeric(r.RateMultiplier),
		Reason:           r.Reason,
		Status:           domain.RequestStatus(r.Status),
		ApprovedByUserID: toUUIDPtr(r.ApprovedByUserID),
		ApprovedAt:       fromPgTimestamptzPtr(r.ApprovedAt),
		CreatedAt:        fromPgTimestamptz(r.CreatedAt),
		UpdatedAt:        fromPgTimestamptz(r.UpdatedAt),
	}
}

func mapAdjustment(r sqlc.AttendanceAdjustments) *domain.AttendanceAdjustment {
	return &domain.AttendanceAdjustment{
		ID:                    toUUID(r.ID),
		AttendanceRecordID:    toUUID(r.AttendanceRecordID),
		RequestedByEmployeeID: toUUID(r.RequestedByEmployeeID),
		OriginalClockIn:       fromPgTimestamptzPtr(r.OriginalClockIn),
		OriginalClockOut:      fromPgTimestamptzPtr(r.OriginalClockOut),
		AdjustedClockIn:       fromPgTimestamptz(r.AdjustedClockIn),
		AdjustedClockOut:      fromPgTimestamptz(r.AdjustedClockOut),
		Reason:                r.Reason,
		Status:                domain.RequestStatus(r.Status),
		ReviewedByUserID:      toUUIDPtr(r.ReviewedByUserID),
		ReviewedAt:            fromPgTimestamptzPtr(r.ReviewedAt),
		ReviewNotes:           fromPgTextPtr(r.ReviewNotes),
		CreatedAt:             fromPgTimestamptz(r.CreatedAt),
	}
}
