package solanchor

import (
	"encoding/json"
	"fmt"
	//"strings"
)

// --- FIELDS

type Field struct {
	Name string    `json:"name"`
	Type FieldType `json:"type"`
	//Index bool      `json:"index"`
}

func (f *Field) SnakeCaseName() string {
	return toSnakeCase(f.Name, true)
}

func (f *Field) SnakeCaseNameWithoutInitialUnderscore() string {
	return toSnakeCase(f.Name, false)
}

// FIELD TYPE

type FieldTypeArray struct {
	Type   FieldType
	Length int
}

type FieldTypeVec struct {
	Type FieldType
}

type FieldTypeOption struct {
	Type FieldType
}

type FieldType struct {
	Simple  string
	Defined string
	Array   *FieldTypeArray
	Vec     *FieldTypeVec
	Option  *FieldTypeOption
}

func (t *FieldType) IsSimple() bool {
	return t.Simple != ""
}

func (t *FieldType) IsSimplePubKey() bool {
	return IsPublicKey(t.Simple)
}

func (t *FieldType) IsDefined() bool {
	return t.Defined != ""
}

func (t *FieldType) IsArray() bool {
	return t.Array != nil
}

func (t *FieldType) IsVec() bool {
	return t.Vec != nil
}

func (t *FieldType) IsOption() bool {
	return t.Option != nil
}

func (t *FieldType) IsTypeUsed(typeName string) bool {
	return isTypeUsedRecursive(t, typeName)
}

func isTypeUsedRecursive(t *FieldType, typeName string) bool {
	if t == nil {
		return false
	}
	if t.Defined == typeName {
		return true
	}
	if t.Array != nil {
		return isTypeUsedRecursive(&t.Array.Type, typeName)
	}
	if t.Vec != nil {
		return isTypeUsedRecursive(&t.Vec.Type, typeName)
	}
	if t.Option != nil {
		return isTypeUsedRecursive(&t.Option.Type, typeName)
	}
	return false
}

func unmarshalDefined(data []byte) string {
	var definedType struct {
		Defined string `json:"defined"`
	}
	if err := json.Unmarshal(data, &definedType); err == nil && definedType.Defined != "" {
		return definedType.Defined
	}

	var definedAlternateType struct {
		Defined struct {
			Name string `json:"name"`
		} `json:"defined"`
	}
	if err := json.Unmarshal(data, &definedAlternateType); err == nil && definedAlternateType.Defined.Name != "" {
		return definedAlternateType.Defined.Name
	}

	return ""
}

func unmarshalSimple(data []byte) string {
	var simpleType string
	if err := json.Unmarshal(data, &simpleType); err == nil && simpleType != "" {
		return simpleType
	}

	return ""
}

func (t *FieldType) UnmarshalJSON(data []byte) error {
	fieldType, err := unmarshallFieldType(data)
	if err != nil {
		return fmt.Errorf("failed to unmarshal Type: %s", string(data))
	}

	*t = *fieldType

	return nil
}

func unmarshallFieldType(data []byte) (*FieldType, error) {
	fieldType := FieldType{}

	if result := unmarshalSimple(data); result != "" {
		fieldType.Simple = result
		return &fieldType, nil
	}

	if result := unmarshalDefined(data); result != "" {
		fieldType.Defined = result
		return &fieldType, nil
	}

	// Try "vec"
	var vecWrapper struct {
		Vec json.RawMessage `json:"vec"`
	}
	if err := json.Unmarshal(data, &vecWrapper); err == nil && len(vecWrapper.Vec) > 0 {
		innerFieldType, err := unmarshallFieldType(vecWrapper.Vec)
		if err != nil {
			return nil, nil
		}
		fieldType.Vec = &FieldTypeVec{*innerFieldType}
		return &fieldType, nil
	}

	// Try "option"
	var optionWrapper struct {
		Option json.RawMessage `json:"option"`
	}
	if err := json.Unmarshal(data, &optionWrapper); err == nil && len(optionWrapper.Option) > 0 {
		innerFieldType, err := unmarshallFieldType(optionWrapper.Option)
		if err != nil {
			return nil, nil
		}
		fieldType.Option = &FieldTypeOption{*innerFieldType}
		return &fieldType, nil
	}

	// Try "array"
	var arrayWrapper struct {
		Array []json.RawMessage `json:"array"`
	}
	if err := json.Unmarshal(data, &arrayWrapper); err == nil && len(arrayWrapper.Array) == 2 {
		innerFieldType, err := unmarshallFieldType(arrayWrapper.Array[0])
		if err != nil {
			return nil, nil
		}
		var length int
		if err := json.Unmarshal(arrayWrapper.Array[1], &length); err != nil {
			return nil, nil
		}
		fieldType.Array = &FieldTypeArray{
			Type:   *innerFieldType,
			Length: length,
		}
		return &fieldType, nil
	}

	return nil, fmt.Errorf("failed to unmarshal Type: %s", string(data))
}
