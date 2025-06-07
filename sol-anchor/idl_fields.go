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
	ArrayRecursive  *FieldTypeArray
	VecRecursive    *FieldTypeVec
	OptionRecursive *FieldTypeOption

	Simple  *Simple
	Defined *Defined

	VecSimple        *VecSimple
	VecDefined       *VecDefined
	VecOptionSimple  *VecOptionSimple
	VecOptionDefined *VecOptionDefined

	OptionSimple            *OptionSimple
	OptionDefined           *OptionDefined
	OptionVecSimple         *OptionVecSimple
	OptionVecDefined        *OptionVecDefined
	OptionArraySimple       *OptionArraySimple
	OptionArrayDefined      *OptionArrayDefined
	OptionArrayArraySimple  *OptionArrayArraySimple
	OptionArrayArrayDefined *OptionArrayArrayDefined

	ArraySimple       *ArraySimple
	ArrayDefined      *ArrayDefined
	ArrayArraySimple  *ArrayArraySimple
	ArrayArrayDefined *ArrayArrayDefined
}

func (t *FieldType) GetResolvedFieldType() (ResolvedFieldType, error) {
	if t.IsSimple() {
		return t.Simple, nil
	}
	if t.IsDefined() {
		return t.Defined, nil
	}
	if t.IsVecSimple() {
		return t.VecSimple, nil
	}
	if t.IsVecDefined() {
		return t.VecDefined, nil
	}
	/*if t.IsVecOptionSimple() {
		return t.VecOptionSimple, nil
	}
	if t.IsVecOptionDefined() {
		return t.VecOptionDefined, nil
	}
	if t.IsOptionSimple() {
		return t.OptionSimple, nil
	}
	if t.IsOptionDefined() {
		return t.OptionDefined, nil
	}
	if t.IsOptionVecSimple() {
		return t.OptionVecSimple, nil
	}
	if t.IsOptionVecDefined() {
		return t.OptionVecDefined, nil
	}
	if t.IsOptionArraySimple() {
		return t.OptionArraySimple, nil
	}
	if t.IsOptionArrayDefined() {
		return t.OptionArrayDefined, nil
	}
	if t.IsOptionArrayArraySimple() {
		return t.OptionArrayArraySimple, nil
	}
	if t.IsOptionArrayArrayDefined() {
		return t.OptionArrayArrayDefined, nil
	}
	if t.IsArraySimple() {
		return t.ArraySimple, nil
	}
	if t.IsArrayDefined() {
		return t.ArrayDefined, nil
	}
	if t.IsArrayArraySimple() {
		return t.ArrayArraySimple, nil
	}
	if t.IsArrayArrayDefined() {
		return t.ArrayArrayDefined, nil
	}*/

	return nil, fmt.Errorf("Unsupported type")
}

func (t *FieldType) IsVecRecursive() bool {
	return t.VecRecursive != nil
}

func (t *FieldType) IsOptionRecursive() bool {
	return t.OptionRecursive != nil
}

func (t *FieldType) IsArrayRecursive() bool {
	return t.ArrayRecursive != nil
}

func (t *FieldType) IsSimple() bool {
	return t.Simple != nil
}

func (t *FieldType) IsDefined() bool {
	return t.Defined != nil
}

func (t *FieldType) IsVecSimple() bool {
	return t.VecSimple != nil
}

func (t *FieldType) IsVecDefined() bool {
	return t.VecDefined != nil
}

func (t *FieldType) IsVecOptionSimple() bool {
	return t.VecOptionSimple != nil
}

func (t *FieldType) IsVecOptionDefined() bool {
	return t.VecOptionDefined != nil
}

func (t *FieldType) IsOptionSimple() bool {
	return t.OptionSimple != nil
}

func (t *FieldType) IsOptionDefined() bool {
	return t.OptionDefined != nil
}

func (t *FieldType) IsOptionVecSimple() bool {
	return t.OptionVecSimple != nil
}

func (t *FieldType) IsOptionVecDefined() bool {
	return t.OptionVecDefined != nil
}

func (t *FieldType) IsOptionArraySimple() bool {
	return t.OptionArraySimple != nil
}

func (t *FieldType) IsOptionArrayDefined() bool {
	return t.OptionArrayDefined != nil
}

func (t *FieldType) IsOptionArrayArraySimple() bool {
	return t.OptionArrayArraySimple != nil
}

func (t *FieldType) IsOptionArrayArrayDefined() bool {
	return t.OptionArrayArrayDefined != nil
}

func (t *FieldType) IsArraySimple() bool {
	return t.ArraySimple != nil
}

func (t *FieldType) IsArrayDefined() bool {
	return t.ArrayDefined != nil
}

func (t *FieldType) IsArrayArraySimple() bool {
	return t.ArrayArraySimple != nil
}

func (t *FieldType) IsArrayArrayDefined() bool {
	return t.ArrayArrayDefined != nil
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
		fieldType.Simple = &Simple{}
		fieldType.Simple.Type = result
		return &fieldType, nil
	}

	if result := unmarshalDefined(data); result != "" {
		fieldType.Defined = &Defined{}
		fieldType.Defined.Type = result
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
		fieldType.VecRecursive = &FieldTypeVec{*innerFieldType}

		// build type
		if fieldType.VecRecursive.Type.IsSimple() {
			fieldType.VecSimple = &VecSimple{}
			fieldType.VecSimple.Type = CastInRustIfNeeded(fieldType.VecRecursive.Type.Simple.Type)
		}
		if fieldType.VecRecursive.Type.IsDefined() {
			fieldType.VecDefined = &VecDefined{}
			fieldType.VecDefined.Type = fieldType.VecRecursive.Type.Defined.Type
		}
		if fieldType.VecRecursive.Type.IsOptionRecursive() {
			if fieldType.VecRecursive.Type.OptionRecursive.Type.IsSimple() {
				fieldType.VecOptionSimple = &VecOptionSimple{}
				fieldType.VecOptionSimple.Type = CastInRustIfNeeded(fieldType.VecRecursive.Type.OptionRecursive.Type.Simple.Type)
			}
			if fieldType.VecRecursive.Type.OptionRecursive.Type.IsDefined() {
				fieldType.VecOptionDefined = &VecOptionDefined{}
				fieldType.VecOptionDefined.Type = fieldType.VecRecursive.Type.OptionRecursive.Type.Defined.Type
			}
		}

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
		fieldType.OptionRecursive = &FieldTypeOption{*innerFieldType}

		// build type
		if fieldType.OptionRecursive.Type.IsSimple() {
			fieldType.OptionSimple = &OptionSimple{}
			fieldType.OptionSimple.Type = CastInRustIfNeeded(fieldType.OptionRecursive.Type.Simple.Type)
		}
		if fieldType.OptionRecursive.Type.IsDefined() {
			fieldType.OptionDefined = &OptionDefined{}
			fieldType.OptionDefined.Type = fieldType.OptionRecursive.Type.Defined.Type
		}
		if fieldType.OptionRecursive.Type.IsVecRecursive() {
			if fieldType.OptionRecursive.Type.VecRecursive.Type.IsSimple() {
				fieldType.OptionVecSimple = &OptionVecSimple{}
				fieldType.OptionVecSimple.Type = CastInRustIfNeeded(fieldType.OptionRecursive.Type.VecRecursive.Type.Simple.Type)
			}
			if fieldType.OptionRecursive.Type.VecRecursive.Type.IsDefined() {
				fieldType.OptionVecDefined = &OptionVecDefined{}
				fieldType.OptionVecDefined.Type = fieldType.OptionRecursive.Type.VecRecursive.Type.Defined.Type
			}
		}
		if fieldType.OptionRecursive.Type.IsArrayRecursive() {
			if fieldType.OptionRecursive.Type.ArrayRecursive.Type.IsSimple() {
				fieldType.OptionArraySimple = &OptionArraySimple{}
				fieldType.OptionArraySimple.Type = CastInRustIfNeeded(fieldType.OptionRecursive.Type.ArrayRecursive.Type.Simple.Type)
				fieldType.OptionArraySimple.Length = fieldType.OptionRecursive.Type.ArrayRecursive.Length
			}
			if fieldType.OptionRecursive.Type.ArrayRecursive.Type.IsDefined() {
				fieldType.OptionArrayDefined = &OptionArrayDefined{}
				fieldType.OptionArrayDefined.Type = fieldType.OptionRecursive.Type.ArrayRecursive.Type.Defined.Type
				fieldType.OptionArrayDefined.Length = fieldType.OptionRecursive.Type.ArrayRecursive.Length
			}
			// nested array
			if fieldType.OptionRecursive.Type.ArrayRecursive.Type.IsArrayRecursive() {
				if fieldType.OptionRecursive.Type.ArrayRecursive.Type.ArrayRecursive.Type.IsSimple() {
					fieldType.OptionArrayArraySimple = &OptionArrayArraySimple{}
					fieldType.OptionArrayArraySimple.Type = CastInRustIfNeeded(fieldType.OptionRecursive.Type.ArrayRecursive.Type.ArrayRecursive.Type.Simple.Type)
					fieldType.OptionArrayArraySimple.Length = fieldType.OptionRecursive.Type.ArrayRecursive.Type.ArrayRecursive.Length
					fieldType.OptionArrayArraySimple.OuterLength = fieldType.OptionRecursive.Type.ArrayRecursive.Length
				}
				if fieldType.OptionRecursive.Type.ArrayRecursive.Type.ArrayRecursive.Type.IsDefined() {
					fieldType.OptionArrayArrayDefined = &OptionArrayArrayDefined{}
					fieldType.OptionArrayArrayDefined.Type = fieldType.OptionRecursive.Type.ArrayRecursive.Type.ArrayRecursive.Type.Defined.Type
					fieldType.OptionArrayArrayDefined.Length = fieldType.OptionRecursive.Type.ArrayRecursive.Type.ArrayRecursive.Length
					fieldType.OptionArrayArrayDefined.OuterLength = fieldType.OptionRecursive.Type.ArrayRecursive.Length
				}
			}
		}

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
		fieldType.ArrayRecursive = &FieldTypeArray{
			Type:   *innerFieldType,
			Length: length,
		}

		if innerFieldType.IsSimple() {
			fieldType.ArraySimple = &ArraySimple{}
			fieldType.ArraySimple.Type = CastInRustIfNeeded(innerFieldType.Simple.Type)
			fieldType.ArraySimple.Length = length
		}
		if innerFieldType.IsDefined() {
			fieldType.ArrayDefined = &ArrayDefined{}
			fieldType.ArrayDefined.Type = innerFieldType.Defined.Type
			fieldType.ArrayDefined.Length = length
		}
		if innerFieldType.IsArrayRecursive() {
			innerArray := innerFieldType.ArrayRecursive
			if innerArray.Type.IsSimple() {
				fieldType.ArrayArraySimple = &ArrayArraySimple{}
				fieldType.ArrayArraySimple.Type = CastInRustIfNeeded(innerArray.Type.Simple.Type)
				fieldType.ArrayArraySimple.Length = innerArray.Length
				fieldType.ArrayArraySimple.OuterLength = length
			}
			if innerArray.Type.IsDefined() {
				fieldType.ArrayArrayDefined = &ArrayArrayDefined{}
				fieldType.ArrayArrayDefined.Type = innerArray.Type.Defined.Type
				fieldType.ArrayArrayDefined.Length = innerArray.Length
				fieldType.ArrayArrayDefined.OuterLength = length
			}
		}

		return &fieldType, nil
	}

	return nil, fmt.Errorf("failed to unmarshal Type: %s", string(data))
}
