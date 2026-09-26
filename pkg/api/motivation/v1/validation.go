package motivationv1

import (
	"fmt"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

type allValidator interface {
	ValidateAll() error
}

// ValidateMessage composes generated PGV validation with semantic protobuf
// rules that PGV cannot express, including real calendar dates and the
// state-dependent terminal outcome of a Close operation.
func ValidateMessage(message proto.Message) error {
	if message == nil {
		return fmt.Errorf("validate protobuf message: message is required")
	}

	reflected := message.ProtoReflect()
	if !reflected.IsValid() {
		return fmt.Errorf("validate protobuf message: message is required")
	}

	validator, ok := message.(allValidator)
	if ok {
		if err := validator.ValidateAll(); err != nil {
			return fmt.Errorf("validate protobuf message: %w", err)
		}
	}

	if err := validateSemanticRules(reflected); err != nil {
		return fmt.Errorf("validate protobuf message: %w", err)
	}

	return nil
}

// ValidateUpdateBaseCriteriaRequest composes PGV validation with validation of
// the paths carried by the FieldMask well-known type. PGV does not inspect the
// contents of FieldMask, so the supported paths are read from the protobuf
// field option declared on update_mask.
func ValidateUpdateBaseCriteriaRequest(request *UpdateBaseCriteriaRequest) error {
	if err := ValidateMessage(request); err != nil {
		return fmt.Errorf("validate update base criteria request: %w", err)
	}

	return validateUpdateMaskPaths(
		request.ProtoReflect(),
		"validate update base criteria request",
	)
}

// ValidateUpdateCriterionRequest composes PGV validation with FieldMask path
// validation, mirroring the narrowed update contract of the base catalog.
func ValidateUpdateCriterionRequest(request *UpdateCriterionRequest) error {
	if err := ValidateMessage(request); err != nil {
		return fmt.Errorf("validate update criterion request: %w", err)
	}

	return validateUpdateMaskPaths(
		request.ProtoReflect(),
		"validate update criterion request",
	)
}

func validateUpdateMaskPaths(message protoreflect.Message, prefix string) error {
	maskField := message.Descriptor().Fields().ByName("update_mask")
	if maskField == nil || !message.Has(maskField) {
		return fmt.Errorf("%s: update_mask is required", prefix)
	}

	mask := message.Get(maskField).Message()
	pathsList := mask.Get(mask.Descriptor().Fields().ByName("paths")).List()

	paths := make([]string, 0, pathsList.Len())
	for index := 0; index < pathsList.Len(); index++ {
		paths = append(paths, pathsList.Get(index).String())
	}

	if len(paths) == 0 {
		return fmt.Errorf("%s: update_mask.paths must not be empty", prefix)
	}

	allowed, err := allowedFieldMaskPaths(message, prefix)
	if err != nil {
		return err
	}

	for _, path := range paths {
		if _, ok := allowed[path]; !ok {
			return fmt.Errorf("%s: update_mask.paths contains unsupported path %q", prefix, path)
		}

		if err := validateUpdateMaskPathValue(message, path, prefix); err != nil {
			return err
		}
	}

	return nil
}

func validateUpdateMaskPathValue(message protoreflect.Message, path, prefix string) error {
	if path == "valid_to" {
		// Absence deliberately clears the nullable validity end.
		return nil
	}

	field := message.Descriptor().Fields().ByName(protoreflect.Name(path))
	if field == nil || !message.Has(field) {
		return fmt.Errorf("%s: value for update_mask path %q is required", prefix, path)
	}

	return nil
}

func allowedFieldMaskPaths(
	message protoreflect.Message,
	prefix string,
) (map[string]struct{}, error) {
	descriptor := message.Descriptor().Fields().ByName("update_mask")

	options, ok := descriptor.Options().(*descriptorpb.FieldOptions)
	if !ok {
		return nil, fmt.Errorf("%s: read update_mask options", prefix)
	}

	extension := proto.GetExtension(options, E_AllowedFieldMaskPath)

	paths, ok := extension.([]string)
	if !ok {
		return nil, fmt.Errorf("%s: read allowed update_mask paths", prefix)
	}

	allowed := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		allowed[path] = struct{}{}
	}

	return allowed, nil
}

// ValidateSavePerformanceSheetRequest composes PGV validation with the batch
// rules PGV cannot express: the command must change at least one cell and
// score entries must reference distinct criteria. Score range, comment rules
// and the final-state invariant remain domain authority.
func ValidateSavePerformanceSheetRequest(request *SavePerformanceSheetRequest) error {
	if err := ValidateMessage(request); err != nil {
		return fmt.Errorf("validate save performance sheet request: %w", err)
	}

	if len(request.GetScores()) == 0 && request.GetAdjustment() == nil {
		return fmt.Errorf(
			"validate save performance sheet request: at least one score change or adjustment change is required",
		)
	}

	seen := make(map[string]struct{}, len(request.GetScores()))
	for _, change := range request.GetScores() {
		id := change.GetSheetCriterionId()
		if _, ok := seen[id]; ok {
			return fmt.Errorf(
				"validate save performance sheet request: duplicate sheet_criterion_id %q",
				id,
			)
		}

		seen[id] = struct{}{}
	}

	return nil
}

func validateUpdateBaseCriteriaPathValue(request *UpdateBaseCriteriaRequest, path string) error {//nolint:unused
	if path == "valid_to" {
		// Absence deliberately clears the nullable validity end.
		return nil
	}

	message := request.ProtoReflect()

	field := message.Descriptor().Fields().ByName(protoreflect.Name(path))
	if field == nil || !message.Has(field) {
		return fmt.Errorf(
			"validate update base criteria request: value for update_mask path %q is required",
			path,
		)
	}

	return nil
}

func updateBaseCriteriaAllowedPaths() (map[string]struct{}, error) {//nolint:unused
	descriptor := (&UpdateBaseCriteriaRequest{}).ProtoReflect().
		Descriptor().
		Fields().
		ByName("update_mask")

	options, ok := descriptor.Options().(*descriptorpb.FieldOptions)
	if !ok {
		return nil, fmt.Errorf("validate update base criteria request: read update_mask options")
	}

	extension := proto.GetExtension(options, E_AllowedFieldMaskPath)

	paths, ok := extension.([]string)
	if !ok {
		return nil, fmt.Errorf(
			"validate update base criteria request: read allowed update_mask paths",
		)
	}

	allowed := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		allowed[path] = struct{}{}
	}

	return allowed, nil
}

func validateSemanticRules(message protoreflect.Message) error {
	switch value := message.Interface().(type) {
	case *Date:
		if err := validateCalendarDate(value); err != nil {
			return err
		}
	case *ClosePeriodOperation:
		if err := validateClosePeriodOperationOutcome(value); err != nil {
			return err
		}
	case *PeriodCriterion:
		if err := validateCriterionScope(
			value.GetType(),
			value.GetBaseCriteriaId(),
			value.GetCriterionId(),
			value.GetPositionId(),
		); err != nil {
			return err
		}
	case *SheetCriterion:
		if err := validateCriterionScope(
			value.GetType(),
			value.GetBaseCriteriaId(),
			value.GetCriterionId(),
			value.GetPositionId(),
		); err != nil {
			return err
		}
	case *CreateBaseCriteriaRequest:
		if err := validateCatalogInterval(value.GetValidFrom(), value.GetValidTo()); err != nil {
			return err
		}
	case *CreateCriterionRequest:
		if err := validateCatalogInterval(value.GetValidFrom(), value.GetValidTo()); err != nil {
			return err
		}
	}

	fields := message.Descriptor().Fields()
	for index := 0; index < fields.Len(); index++ {
		field := fields.Get(index)
		if field.Message() == nil || !message.Has(field) {
			continue
		}

		if field.IsMap() {
			if field.MapValue().Message() == nil {
				continue
			}

			var validationErr error

			message.Get(field).
				Map().
				Range(func(_ protoreflect.MapKey, value protoreflect.Value) bool {
					validationErr = validateSemanticRules(value.Message())

					return validationErr == nil
				})

			if validationErr != nil {
				return validationErr
			}

			continue
		}

		if field.IsList() {
			list := message.Get(field).List()
			for listIndex := 0; listIndex < list.Len(); listIndex++ {
				if err := validateSemanticRules(list.Get(listIndex).Message()); err != nil {
					return err
				}
			}

			continue
		}

		if err := validateSemanticRules(message.Get(field).Message()); err != nil {
			return err
		}
	}

	return nil
}

func validateCalendarDate(date *Date) error {
	calendarDate := time.Date(
		int(date.GetYear()),
		time.Month(date.GetMonth()),
		int(date.GetDay()),
		0,
		0,
		0,
		0,
		time.UTC,
	)
	if calendarDate.Year() != int(date.GetYear()) ||
		calendarDate.Month() != time.Month(date.GetMonth()) ||
		calendarDate.Day() != int(date.GetDay()) {
		return fmt.Errorf(
			"date %04d-%02d-%02d is not a valid calendar date",
			date.GetYear(),
			date.GetMonth(),
			date.GetDay(),
		)
	}

	return nil
}

// validateCatalogInterval rejects a validity end before the validity start on
// catalog creation requests; PGV cannot compare two message fields.
func validateCatalogInterval(validFrom, validTo *Date) error {
	if validTo == nil || validFrom == nil {
		return nil
	}

	if dateDayNumber(validTo) < dateDayNumber(validFrom) {
		return fmt.Errorf(
			"valid_to %04d-%02d-%02d is before valid_from %04d-%02d-%02d",
			validTo.GetYear(), validTo.GetMonth(), validTo.GetDay(),
			validFrom.GetYear(), validFrom.GetMonth(), validFrom.GetDay(),
		)
	}

	return nil
}

func dateDayNumber(date *Date) int {
	return int(date.GetYear())*10000 + int(date.GetMonth())*100 + int(date.GetDay())
}

// validateCriterionScope enforces the frozen type/source/position contract
// shared by PeriodCriterion and SheetCriterion: BASE refers to the base
// catalog without a position, SPECIAL refers to the special catalog with a
// positive position.
func validateCriterionScope(
	criterionType CriterionType,
	baseCriteriaID string,
	criterionID string,
	positionID int64,
) error {
	switch criterionType {
	case CriterionType_CRITERION_TYPE_BASE:
		if baseCriteriaID == "" || criterionID != "" {
			return fmt.Errorf("base criterion must reference base_criteria_id and no other source")
		}

		if positionID != 0 {
			return fmt.Errorf("base criterion must not carry a position_id")
		}
	case CriterionType_CRITERION_TYPE_SPECIAL:
		if criterionID == "" || baseCriteriaID != "" {
			return fmt.Errorf("special criterion must reference criterion_id and no other source")
		}

		if positionID <= 0 {
			return fmt.Errorf("special criterion must carry a positive position_id")
		}
	default:
		return fmt.Errorf("criterion type must not be unspecified")
	}

	return nil
}

func validateClosePeriodOperationOutcome(operation *ClosePeriodOperation) error {
	switch operation.GetState() {
	case ClosePeriodOperationState_CLOSE_PERIOD_OPERATION_STATE_QUEUED,
		ClosePeriodOperationState_CLOSE_PERIOD_OPERATION_STATE_RUNNING:
		if operation.GetTerminalOutcome() != nil {
			return fmt.Errorf(
				"close period operation terminal_outcome must be absent before completion",
			)
		}
	case ClosePeriodOperationState_CLOSE_PERIOD_OPERATION_STATE_SUCCEEDED:
		outcome, ok := operation.GetTerminalOutcome().(*ClosePeriodOperation_Result)
		if !ok || outcome.Result == nil {
			return fmt.Errorf("succeeded close period operation must contain a period result")
		}
	case ClosePeriodOperationState_CLOSE_PERIOD_OPERATION_STATE_FAILED:
		outcome, ok := operation.GetTerminalOutcome().(*ClosePeriodOperation_Error)
		if !ok || outcome.Error == nil {
			return fmt.Errorf("failed close period operation must contain an RPC status error")
		}
	default:
		return fmt.Errorf("close period operation state must not be unspecified")
	}

	return nil
}
