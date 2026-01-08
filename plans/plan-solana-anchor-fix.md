# Implementation Plan: Fix Solana Anchor Generator Project Name

## Overview

**Objective**: Fix the Solana Anchor generator to respect the user's input project name in the generated `substreams.yaml` file.

**Current Behavior**: The generator always writes `my_project` as the package name in `substreams.yaml`, regardless of user input.

**Expected Behavior**: The generator should use the user-provided project name (converted to underscore format) in `substreams.yaml`, matching the behavior in `Cargo.toml`.

## Problem Analysis

### Root Cause

The template file [sol-anchor/templates/substreams.yaml.gotmpl](sol-anchor/templates/substreams.yaml.gotmpl) has a hardcoded project name on line 3:

```yaml
name: my_project  # ❌ Hardcoded
```

Should be:

```yaml
name: {{ .GetModuleName }}  # ✅ Dynamic
```

### Why This Matters

- **Inconsistency**: `Cargo.toml` correctly uses `{{ .Name }}`, but `substreams.yaml` doesn't
- **User Confusion**: Users expect their chosen project name to appear in all generated files
- **Package Identity**: The package name in `substreams.yaml` identifies the Substreams module

### Context

The `GetModuleName()` method (from [base_state.go:46-48](base_state.go#L46-L48)) converts project names to the correct format:

```go
func (p *BaseConversationState) GetModuleName() string {
    return strings.ReplaceAll(p.Name, "-", "_")
}
```

This ensures `my-project` → `my_project` (YAML/identifier safe format).

## Implementation Steps

### Step 1: Update Template File

**File**: [sol-anchor/templates/substreams.yaml.gotmpl](sol-anchor/templates/substreams.yaml.gotmpl)

**Change**:
```diff
 specVersion: v0.1.0
 package:
-  name: my_project
+  name: {{ .GetModuleName }}
   version: v0.1.0
```

**Rationale**: Use the same pattern as other generators (e.g., [sol-hello-world/templates/substreams.yaml.gotmpl:3](sol-hello-world/templates/substreams.yaml.gotmpl#L3))

### Step 2: Verify No Other Hardcoded Values

**Action**: Review all template files in `sol-anchor/templates/` for other hardcoded values

**Known Status**:
- ✅ `Cargo.toml.gotmpl` - Correctly uses `{{ .Name }}`
- ✅ `substreams.yaml.gotmpl` line 42 - Correctly uses `{{ .ChainName }}`
- ❓ Other templates - Need verification

### Step 3: Run Tests

**Unit/Integration Tests**:
```bash
go test ./sol-anchor/... -v
```

**Manual Verification**:
1. Generate a project with custom name: `my-custom-project`
2. Check generated `substreams.yaml` has `name: my_custom_project`
3. Check generated `Cargo.toml` has `name = "my-custom-project"`

### Step 4: Validate Test Fixtures

**Action**: Review test files in `tests/sol-anchor/` directory

**Files to Check**:
- `tests/sol-anchor/jupiter-governance.json`
- `tests/sol-anchor/sanctum.json`
- `tests/sol-anchor/orca.json`
- `tests/sol-anchor/meteora.json`
- `tests/sol-anchor/oasis.json`
- `tests/sol-anchor/pump-fun.json`
- `tests/sol-anchor/jupiter-staking.json`
- `tests/sol-anchor/bonkswap.json`
- `tests/sol-anchor/raydium-cp-swap.json`
- `tests/sol-anchor/lifinity.json`

**Expected**: These fixtures should have `"name": "my_project"` in their state, which will now correctly generate `name: my_project` in the output.

## Files Modified

| File | Lines | Change Type | Description |
|------|-------|-------------|-------------|
| `sol-anchor/templates/substreams.yaml.gotmpl` | 3 | Template Fix | Replace hardcoded `my_project` with `{{ .GetModuleName }}` |

## Testing Plan

### Automated Tests

- [ ] Run `go test ./sol-anchor/... -v` - All tests must pass
- [ ] Run full integration test suite if available

### Manual Testing

- [ ] Generate project with name `test-project-one`
  - [ ] Verify `substreams.yaml` contains `name: test_project_one`
  - [ ] Verify `Cargo.toml` contains `name = "test-project-one"`
- [ ] Generate project with name `my_solana_anchor`
  - [ ] Verify `substreams.yaml` contains `name: my_solana_anchor`
  - [ ] Verify `Cargo.toml` contains `name = "my_solana_anchor"`

### Edge Cases

- [ ] Test with hyphens: `my-test-project` → `my_test_project`
- [ ] Test with underscores: `my_test_project` → `my_test_project`
- [ ] Test with numbers: `project123` → `project123`

## Risk Assessment

**Risk Level**: 🟢 LOW

**Justification**:
- Single line change in template file
- No Go code changes required
- Follows established pattern from other generators
- Template variable already exists and is proven to work
- Change is backwards compatible (templates still generate valid YAML)

**Mitigation**:
- Comprehensive test coverage before merge
- Review all test fixtures to ensure they still pass

## Success Criteria

- [x] Root cause identified and documented
- [ ] Template file updated with `{{ .GetModuleName }}`
- [ ] All existing tests pass
- [ ] Manual testing confirms project name is correctly rendered
- [ ] No other hardcoded values found in templates
- [ ] Generated files are valid and functional

## Rollback Plan

If issues are discovered:
1. Revert the single line change in `substreams.yaml.gotmpl`
2. Re-run tests to confirm stability
3. Investigate why the template variable didn't work as expected

## References

- Similar implementation: [sol-hello-world/templates/substreams.yaml.gotmpl:3](sol-hello-world/templates/substreams.yaml.gotmpl#L3)
- Base state code: [base_state.go:46-48](base_state.go#L46-L48)
- Template file: [sol-anchor/templates/substreams.yaml.gotmpl](sol-anchor/templates/substreams.yaml.gotmpl)
- Cargo template: [sol-anchor/templates/Cargo.toml.gotmpl](sol-anchor/templates/Cargo.toml.gotmpl)

## Notes

This is a straightforward bug fix that aligns the sol-anchor generator with the established patterns used in other generators within the codebase. The fix is minimal, low-risk, and has clear acceptance criteria.
