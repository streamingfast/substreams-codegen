package solanchor

func uniqueStrings(input []string) []string {
	seen := make(map[string]struct{})
	var result []string

	for _, val := range input {
		if _, exists := seen[val]; !exists {
			seen[val] = struct{}{}
			result = append(result, val)
		}
	}

	return result
}

func IDLTypeToRustType(idlType string) string {
	switch idlType {
	case "string":
		return "String"
	case "bytes":
		return "Vec<u8>"
	case "publicKey", "pubKey", "Pubkey", "PubKey":
		return "[u8;32]"
	default:
		return idlType
	}
}

func IDLTypeToProtobufType(idlType string) string {
	switch idlType {
	case "string":
		return "string"
	case "bool":
		return "bool"
	case "u8", "i8":
		return "int32"
	case "u16", "i16":
		return "int32"
	case "u32", "i32":
		return "int32"
	case "u64", "i64":
		return "int64"
	case "u128", "i128":
		return "string" // Protobuf doesn't support 128-bit ints directly
	case "bytes":
		return "bytes"
	case "publicKey", "pubKey", "Pubkey", "PubKey":
		return "string"
	default:
		// Assume it's a user-defined message
		return idlType
	}
}
