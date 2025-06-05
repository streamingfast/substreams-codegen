package solanchor

import "fmt"

/*const (
	Simple ResolvedFieldType = iota
	SimplePubKey

	Defined

	VecSimple
	VecDefined
	VecOptionSimple
	VecOptionDefined

	OptionSimple
	OptionDefined
	OptionVecSimple
	OptionVecDefined
	OptionArraySimple
	OptionArrayDefined
	OptionArrayArraySimple
	OptionArrayArrayDefined

	ArraySimple
	ArrayDefined
	ArrayArraySimple
	ArrayArrayDefined
)*/

// Parent struct
type ResolvedFieldType struct {
	Type string
}

// simple and defiend
type Simple struct {
	ResolvedFieldType
}
type Defined struct {
	ResolvedFieldType
}

// vec
type VecSimple struct {
	ResolvedFieldType
}
type VecDefined struct {
	ResolvedFieldType
}
type VecOptionSimple struct {
	ResolvedFieldType
}
type VecOptionDefined struct {
	ResolvedFieldType
}

// option
type OptionSimple struct {
	ResolvedFieldType
}
type OptionDefined struct {
	ResolvedFieldType
}
type OptionVecSimple struct {
	ResolvedFieldType
}
type OptionVecDefined struct {
	ResolvedFieldType
}
type OptionArraySimple struct {
	ResolvedFieldType
	Length int
}
type OptionArrayDefined struct {
	ResolvedFieldType
	Length int
}
type OptionArrayArraySimple struct {
	ResolvedFieldType
	LengthX int
	LengthY int
}
type OptionArrayArrayDefined struct {
	ResolvedFieldType
	LengthX int
	LengthY int
}

// array
type ArraySimple struct {
	ResolvedFieldType
	Length int
}
type ArrayDefined struct {
	ResolvedFieldType
	Length int
}
type ArrayArraySimple struct {
	ResolvedFieldType
	LengthX int
	LengthY int
}
type ArrayArrayDefined struct {
	ResolvedFieldType
	LengthX int
	LengthY int
}

// Add a method to the Status type
func (s ResolvedFieldType) String() string {
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
}

/* Rust Type generation */
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
}
