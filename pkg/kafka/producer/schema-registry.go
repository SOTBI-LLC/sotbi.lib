package producer

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/jhump/protoreflect/desc" //nolint:staticcheck
	"github.com/jhump/protoreflect/desc/protoprint"
	"github.com/riferrei/srclient"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/protoadapt"
)

const magicByte byte = 0x0

type SchemaRegistry struct {
	URL      string
	Username string
	Password string
  SchemaNames map[string]proto.Message
	schemas  map[string]int
}

// WithSchemaRegistry при инициализации продюсера публикует схемы в Kafka Schema Registry.
func WithSchemaRegistry(sr SchemaRegistry) func(opts *ProducerOptions) {
	return func(opts *ProducerOptions) {
		if sr.URL != "" {
			opts.schemaRegistry = &sr
		}
	}
}

func (r *SchemaRegistry) registerSchemas() error {
	sr := srclient.CreateSchemaRegistryClient(r.URL)

	if r.Username != "" && r.Password != "" {
		sr.SetCredentials(r.Username, r.Password)
	}

	sr.CachingEnabled(false)

	r.schemas = make(map[string]int)

	for topic, msg := range r.SchemaNames {
		valueSubject := fmt.Sprintf("%s-value", topic)

		valueSchema, err := toProtoFileString(msg)
		if err != nil {
			return err
		}

		if _, err = sr.ChangeSubjectCompatibilityLevel(
			valueSubject,
			srclient.BackwardTransitive,
		); err != nil {
			return err
		}

		schemaID, err := registerSubject(sr, valueSubject, valueSchema, srclient.Protobuf)
		if err != nil {
			return err
		}

		r.schemas[topic] = schemaID
	}

	return nil
}

func (r *SchemaRegistry) getSchemaID(topic string) *int {
	schemaID, ok := r.schemas[topic]
	if !ok {
		return nil
	}

	return &schemaID
}

func addSchemaIDPrefix(id int, msgBytes []byte) ([]byte, error) {
	var buf bytes.Buffer

	if err := buf.WriteByte(magicByte); err != nil {
		return nil, err
	}

	idBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(idBytes, uint32(id))

	if _, err := buf.Write(idBytes); err != nil {
		return nil, err
	}

	if _, err := buf.Write(msgBytes); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func toProtoFileString(msg proto.Message) (string, error) {
	messageDesc, err := desc.LoadMessageDescriptorForMessage(protoadapt.MessageV1Of(msg))
	if err != nil {
		return "", err
	}

	fileDesc := messageDesc.GetFile()

	printer := protoprint.Printer{Compact: true}

	var writer strings.Builder

	if err := printer.PrintProtoFile(fileDesc, &writer); err != nil {
		return "", err
	}

	return writer.String(), nil
}

func registerSubject(
	sr *srclient.SchemaRegistryClient,
	subject, schema string,
	schemaType srclient.SchemaType,
) (int, error) {
	compabilityLevel, err := sr.GetCompatibilityLevel(subject, true)
	if err != nil {
		var srcErr srclient.Error

		// API Managed Schema Registry от Яндекс возвращает ошибку 40401 Subject not found
		// вне зависимости от значения параметра defaultToGlobal
		if !errors.As(err, &srcErr) || srcErr.Code != 40401 {
			return 0, fmt.Errorf("не удается получить уровень совместимости для топика | %w", err)
		}
	}

	if compabilityLevel == nil || *compabilityLevel != srclient.BackwardTransitive {
		if _, err := sr.ChangeSubjectCompatibilityLevel(subject, srclient.BackwardTransitive); err != nil {
			return 0, fmt.Errorf("не удается изменить уровень совместимости для топика | %w", err)
		}
	}

	latestSchema, err := sr.GetLatestSchema(subject)
	if err == nil {
		isCompatible, err := sr.IsSchemaCompatible(
			subject,
			schema,
			strconv.Itoa(latestSchema.Version()),
			schemaType,
		)
		if err != nil {
			return 0, fmt.Errorf("не удается проверить совместимость схем для топика | %w", err)
		}

		if !isCompatible {
			return 0, fmt.Errorf("схема для %s несовместима с предыдущей версией", subject)
		}
	}

	s, err := sr.CreateSchema(subject, schema, schemaType)
	if err != nil {
		return 0, fmt.Errorf("не удается опубликовать схему в Schema Registry | %w", err)
	}

	return s.ID(), nil
}
