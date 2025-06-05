package solanchor

import "fmt"

// Parent struct
type ResolvedFieldTypeCommon struct {
	Type string
}
type ResolvedFieldType interface {
	ResolveRustType() string
	PrintNecessaryProtobufMessages() string
}

// simple and defiend
type Simple struct {
	ResolvedFieldTypeCommon
}

func (f *Simple) ResolveRustType() string {
	if f.Type == "pubkey" {
		return "PubKey"
	}

	return f.Type
}
func (f *Simple) PrintNecessaryProtobufMessages() string {
	return ""
}

type Defined struct {
	ResolvedFieldTypeCommon
}

func (f *Defined) ResolveRustType() string {
	return f.Type
}
func (f *Defined) PrintNecessaryProtobufMessages() string {
	return ""
}

// vec
type VecSimple struct {
	ResolvedFieldTypeCommon
}

func (f *VecSimple) ResolveRustType() string {
	return fmt.Sprintf("Vec<%s>", f.Type)
}
func (f *VecSimple) PrintNecessaryProtobufMessages() string {
	return ""
}

type VecDefined struct {
	ResolvedFieldTypeCommon
}

func (f *VecDefined) ResolveRustType() string {
	return fmt.Sprintf("Vec<%s>", f.Type)
}
func (f *VecDefined) PrintNecessaryProtobufMessages() string {
	return ""
}

type VecOptionSimple struct {
	ResolvedFieldTypeCommon
}

func (f *VecOptionSimple) ResolveRustType() string {
	return fmt.Sprintf("Vec<Option<%s>>", f.Type)
}
func (f *VecOptionSimple) PrintNecessaryProtobufMessages() string {
	return ""
}

type VecOptionDefined struct {
	ResolvedFieldTypeCommon
}

func (f *VecOptionDefined) ResolveRustType() string {
	return fmt.Sprintf("Vec<Option<%s>>", f.Type)
}
func (f *VecOptionDefined) PrintNecessaryProtobufMessages() string {
	return ""
}

// option
type OptionSimple struct {
	ResolvedFieldTypeCommon
}

func (f *OptionSimple) ResolveRustType() string {
	return fmt.Sprintf("Option<%s>", f.Type)
}
func (f *OptionSimple) PrintNecessaryProtobufMessages() string {
	return ""
}

type OptionDefined struct {
	ResolvedFieldTypeCommon
}

func (f *OptionDefined) ResolveRustType() string {
	return fmt.Sprintf("Option<%s>", f.Type)
}
func (f *OptionDefined) PrintNecessaryProtobufMessages() string {
	return ""
}

type OptionVecSimple struct {
	ResolvedFieldTypeCommon
}

func (f *OptionVecSimple) ResolveRustType() string {
	return fmt.Sprintf("Option<Vec<%s>>", f.Type)
}
func (f *OptionVecSimple) ResolveProtobufType() string {
	return fmt.Sprintf("OptionVecSimple%s", f.Type)
}
func (f *OptionVecSimple) PrintNecessaryProtobufMessages() string {
	innerMessageName := fmt.Sprintf("OptionVecSimple%sInner", f.Type)
	return fmt.Sprintf(`
		message %s {
			repeated %s value = 1;
		}

		message OptionVecSimple%s {
			optional %s inner = 1;
		}
	`, innerMessageName, f.Type, f.Type, innerMessageName)
}

type OptionVecDefined struct {
	ResolvedFieldTypeCommon
}

func (f *OptionVecDefined) ResolveRustType() string {
	return fmt.Sprintf("Option<Vec<%s>>", f.Type)
}
func (f *OptionVecDefined) ResolveProtobufType() string {
	return fmt.Sprintf("OptionVecSimple%s", f.Type)
}
func (f *OptionVecDefined) PrintNecessaryProtobufMessages() string {
	messageName := fmt.Sprintf("OptionVecDefined%s", f.Type)
	innerMessageName := fmt.Sprintf("OptionVecDefined%sInner", f.Type)
	return fmt.Sprintf(`
		message %s {
			repeated %s value = 1;
		}

		message %s {
			optional %s inner = 1;
		}
	`, innerMessageName, f.Type, messageName, innerMessageName)
}

type OptionArraySimple struct {
	ResolvedFieldTypeCommon
	Length int
}

func (f *OptionArraySimple) ResolveRustType() string {
	return fmt.Sprintf("Option<[%s;%d]>", f.Type, f.Length)
}
func (f *OptionArraySimple) ResolveProtobufType() string {
	return fmt.Sprintf("OptionArraySimple%s", f.Type)
}
func (f *OptionArraySimple) PrintNecessaryProtobufMessages() string {
	innerMessageName := fmt.Sprintf("OptionArraySimple%sInner", f.Type)
	return fmt.Sprintf(`
		message %s {
			repeated %s value = 1;
		}

		message OptionArraySimple%s {
			optional %s inner = 1;
		}
	`, innerMessageName, f.Type, f.Type, innerMessageName)
}

type OptionArrayDefined struct {
	ResolvedFieldTypeCommon
	Length int
}

func (f *OptionArrayDefined) ResolveRustType() string {
	return fmt.Sprintf("Option<[%s;%d]>", f.Type, f.Length)
}
func (f *OptionArrayDefined) ResolveProtobufType() string {
	return fmt.Sprintf("OptionArrayDefined%s", f.Type)
}
func (f *OptionArrayDefined) PrintNecessaryProtobufMessages() string {
	innerMessageName := fmt.Sprintf("OptionArrayDefined%sInner", f.Type)
	return fmt.Sprintf(`
		message %s {
			repeated %s value = 1;
		}

		message OptionArrayDefined%s {
			optional %s inner = 1;
		}
	`, innerMessageName, f.Type, f.Type, innerMessageName)
}

type OptionArrayArraySimple struct {
	ResolvedFieldTypeCommon
	Length      int
	OuterLength int
}

func (f *OptionArrayArraySimple) ResolveRustType() string {
	return fmt.Sprintf("Option<[[%s;%d];%d]>", f.Type, f.Length, f.OuterLength)
}
func (f *OptionArrayArraySimple) ResolveProtobufType() string {
	return fmt.Sprintf("OptionArrayArraySimple%s", f.Type)
}
func (f *OptionArrayArraySimple) PrintNecessaryProtobufMessages() string {
	innerMessageName := fmt.Sprintf("OptionArrayArraySimple%sInner", f.Type)
	innerArrayMessageName := fmt.Sprintf("OptionArrayArraySimple%sInnerArray", f.Type)

	return fmt.Sprintf(`
		message %s {
			repeated %s inner = 1;
		}

		message %s {
			repeated %s arrayInner = 1;
		}

		message OptionArrayArraySimple%s {
			optional %s inner = 1;
		}
	`, innerArrayMessageName, f.Type, innerMessageName, innerArrayMessageName, f.Type, innerMessageName)
}

type OptionArrayArrayDefined struct {
	ResolvedFieldTypeCommon
	Length      int
	OuterLength int
}

func (f *OptionArrayArrayDefined) ResolveRustType() string {
	return fmt.Sprintf("Option<[[%s;%d];%d]>", f.Type, f.Length, f.OuterLength)
}
func (f *OptionArrayArrayDefined) ResolveProtobufType() string {
	return fmt.Sprintf("OptionArrayArraySimple%s", f.Type)
}
func (f *OptionArrayArrayDefined) PrintNecessaryProtobufMessages() string {
	innerMessageName := fmt.Sprintf("OptionArrayArrayDefined%sInner", f.Type)
	innerArrayMessageName := fmt.Sprintf("OptionArrayArrayDefined%sInnerArray", f.Type)

	return fmt.Sprintf(`
		message %s {
			repeated %s inner = 1;
		}

		message %s {
			repeated %s arrayInner = 1;
		}

		message OptionArrayArrayDefined%s {
			optional %s inner = 1;
		}
	`, innerArrayMessageName, f.Type, innerMessageName, innerArrayMessageName, f.Type, innerMessageName)
}

// array
type ArraySimple struct {
	ResolvedFieldTypeCommon
	Length int
}

func (f *ArraySimple) ResolveRustType() string {
	return fmt.Sprintf("[%s;%d]", f.Type, f.Length)
}

func (f *ArraySimple) PrintNecessaryProtobufMessages() string {
	return ""
}

type ArrayDefined struct {
	ResolvedFieldTypeCommon
	Length int
}

func (f *ArrayDefined) ResolveRustType() string {
	return fmt.Sprintf("[%s;%d]", f.Type, f.Length)
}

func (f *ArrayDefined) PrintNecessaryProtobufMessages() string {
	return ""
}

type ArrayArraySimple struct {
	ResolvedFieldTypeCommon
	Length      int
	OuterLength int
}

func (f *ArrayArraySimple) ResolveRustType() string {
	return fmt.Sprintf("[[%s;%d];%d]", f.Type, f.Length, f.OuterLength)
}

func (f *ArrayArraySimple) PrintNecessaryProtobufMessages() string {
	return ""
}

type ArrayArrayDefined struct {
	ResolvedFieldTypeCommon
	Length      int
	OuterLength int
}

func (f *ArrayArrayDefined) ResolveRustType() string {
	return fmt.Sprintf("[[%s;%d];%d]", f.Type, f.Length, f.OuterLength)
}

func (f *ArrayArrayDefined) PrintNecessaryProtobufMessages() string {
	return ""
}

// Add a method to the Status type
/*func (s ResolvedFieldType) String() string {
	switch s {
	case Simple:
		return "Active"
	case SimplePubKey:
		return "Inactive"
	case Defined:
		return "Inactive"
	case VecSimple:
		return "Inactive"
	case VecDefined:
		return "Inactive"
	case VecOptionSimple:
		return "Inactive"
	case VecOptionDefined:
		return "Inactive"

	case OptionSimple:
		return "Inactive"
	case OptionDefined:
		return "Inactive"
	case OptionVecSimple:
		return "Inactive"
	case OptionVecDefined:
		return "Inactive"
	case OptionArraySimple:
		return "Inactive"
	case OptionArrayDefined:
		return "Inactive"
	case OptionArrayArraySimple:
		return "Inactive"
	case OptionArrayArrayDefined:
		return "Inactive"

	case SimplePubKey:
		return "Inactive"
	case SimplePubKey:
		return "Inactive"
	default:
		return "Unknown"
	}
}*/

/*
func (f *FieldType) ResolveRustType() string {
	if f.IsSimple() {
		return f.Simple
	}

	if f.IsSimplePubKey() {
		return "PubKey"
	}

	if f.IsDefined() {
		return f.Defined
	}

	// Vec
	if f.IsVec() {
		if f.Vec.Type.IsSimple() {
			return fmt.Sprintf("Vec<%s>", CastInRustIfNeeded(f.Vec.Type.Simple))
		}
		if f.Vec.Type.IsDefined() {
			return fmt.Sprintf("Vec<%s>", f.Vec.Type.Defined)
		}
		if f.Vec.Type.IsOption() {
			if f.Vec.Type.Option.Type.IsSimple() {
				return fmt.Sprintf("Vec<Option<%s>>", CastInRustIfNeeded(f.Vec.Type.Option.Type.Simple))
			}
			if f.Vec.Type.Option.Type.IsDefined() {
				return fmt.Sprintf("Vec<Option<%s>>", f.Vec.Type.Option.Type.Defined)
			}
		}
	}

	// Option
	if f.IsOption() {
		if f.Option.Type.IsSimple() {
			return fmt.Sprintf("Option<%s>", CastInRustIfNeeded(f.Option.Type.Simple))
		}
		if f.Option.Type.IsDefined() {
			return fmt.Sprintf("Option<%s>", f.Option.Type.Defined)
		}
		if f.Option.Type.IsVec() {
			if f.Option.Type.Vec.Type.IsSimple() {
				return fmt.Sprintf("Option<Vec<%s>>", CastInRustIfNeeded(f.Option.Type.Vec.Type.Simple))
			}
			if f.Option.Type.Vec.Type.IsDefined() {
				return fmt.Sprintf("Option<Vec<%s>>", f.Option.Type.Vec.Type.Defined)
			}
		}
		if f.Option.Type.IsArray() {
			if f.Option.Type.Array.Type.IsSimple() {
				return fmt.Sprintf("Option<[%s;%d]>", CastInRustIfNeeded(f.Option.Type.Array.Type.Simple), f.Option.Type.Array.Length)
			}
			if f.Option.Type.Array.Type.IsDefined() {
				return fmt.Sprintf("Option<[%s;%d]>", f.Option.Type.Array.Type.Defined, f.Option.Type.Array.Length)
			}
			if f.Option.Type.Array.Type.IsArray() {
				if f.Option.Type.Array.Type.Array.Type.IsSimple() {
					return fmt.Sprintf("Option<[[%s;%d];%d]>", CastInRustIfNeeded(f.Option.Type.Array.Type.Array.Type.Simple), f.Option.Type.Array.Type.Array.Length, f.Option.Type.Array.Length)
				}
				if f.Option.Type.Array.Type.Array.Type.IsDefined() {
					return fmt.Sprintf("Option<[[%s;%d];%d]>", f.Option.Type.Array.Type.Array.Type.Defined, f.Option.Type.Array.Type.Array.Length, f.Option.Type.Array.Length)
				}
			}
		}
	}

	// Array
	if f.IsArray() {
		if f.Array.Type.IsSimple() {
			return fmt.Sprintf("[%s;%d]", CastInRustIfNeeded(f.Array.Type.Simple), f.Array.Length)
		}
		if f.Array.Type.IsDefined() {
			return fmt.Sprintf("[%s;%d]", f.Array.Type.Defined, f.Array.Length)
		}
		if f.Array.Type.IsArray() {
			if f.Array.Type.Array.Type.IsSimple() {
				return fmt.Sprintf("[[%s;%d];%d]", CastInRustIfNeeded(f.Array.Type.Array.Type.Simple), f.Array.Type.Array.Length, f.Array.Length)
			}
			if f.Array.Type.Array.Type.IsDefined() {
				return fmt.Sprintf("[[%s;%d];%d]", f.Array.Type.Array.Type.Defined, f.Array.Type.Array.Length, f.Array.Length)
			}
		}
	}

	return ""
}*/
