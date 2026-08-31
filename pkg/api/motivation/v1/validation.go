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

	paths := request.GetUpdateMask().GetPaths()
	if len(paths) == 0 {
		return fmt.Errorf(
			"validate update base criteria request: update_mask.paths must not be empty",
		)
	}

	allowed, err := updateBaseCriteriaAllowedPaths()
	if err != nil {
		return err
	}

	for _, path := range paths {
		if _, ok := allowed[path]; !ok {
			return fmt.Errorf(
				"validate update base criteria request: update_mask.paths contains unsupported path %q",
				path,
			)
		}

		if err := validateUpdateBaseCriteriaPathValue(request, path); err != nil {
			return err
		}
	}

	return nil
}

func validateUpdateBaseCriteriaPathValue(request *UpdateBaseCriteriaRequest, path string) error {
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

func updateBaseCriteriaAllowedPaths() (map[string]struct{}, error) {
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
