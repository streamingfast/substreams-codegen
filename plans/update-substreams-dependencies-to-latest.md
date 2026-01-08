# Implementation Plan for Codegen Agent

## Objective
Update all Substreams dependencies to the latest major versions across the `github.com/streamingfast/substreams-codegen` repository and submit a PR to the `develop` branch.

## How to Use This Plan
- Follow steps sequentially, completing each before moving to the next
- Verify your work at each checkpoint using the specified commands
- If you encounter ambiguity, refer to the linked source files for context
- Run all verification commands from the project root directory
- Ensure all tests pass before considering a step complete
- Create a pull request when implementation is complete

---

## Summary of Current State

### Rust Crates (Current -> Target)

| Crate | Current Version | Target Version | Notes |
|-------|----------------|----------------|-------|
| `substreams` | 0.6.0 | 0.7.0 | Core SDK |
| `substreams-ethereum` | 0.10.2 | 0.11.1 | EVM chains |
| `substreams-solana` | 0.14.1 | 0.14.2 | Minor bump |
| `substreams-solana-program-instructions` | 0.2.0 | 0.3.0 | Major bump |
| `substreams-near` | 0.10 | 0.10.2 | Minor bump |
| `substreams-database-change` | 2 | 3.0.0 | Major bump |
| `substreams-entity-change` | 2 | 2.0.0 | Already at latest |
| `prost` | 0.13.3 | 0.13.3 | Keep as-is (0.14.x may have breaking changes) |
| `prost-types` | 0.13.3 | 0.13.3 | Keep as-is |

### Foundational Module Packages (Current -> Target)

| Package | Current Version | Target Version | Used In |
|---------|----------------|----------------|---------|
| `ethereum-common` | v0.3.0 | v0.3.3 | evm-events-calls, evm-events-calls-raw, evm-hello-world |
| `solana-common` | v0.3.0 / v0.3.3 | v0.3.3 | sol-hello-world, sol-anchor, sol-transactions |
| `starknet-foundational` | v0.1.4 | v0.1.4 | starknet-hello-world, starknet-events (already latest) |
| `stellar-foundational` | v0.3.0 / v0.5.0 | v0.5.0 | stellar-minimal, stellar-transactions-operations |
| `injective-common` | v0.2.5 | v0.2.5 | injective-events (already latest) |
| `mantra-common` | v0.1.1 | v0.1.1 | mantra-events (already latest) |
| `tron_foundational` | v0.1.2 | v0.1.2 | tron-hello-world, tron-transactions (already latest) |
| `cosmos (substreams-cosmos)` | v0.1.5 | v0.1.5 | injective-hello-world, mantra-hello-world (already latest) |

### getrandom Dependency Analysis

The `getrandom` dependency with `features = ["custom"]` is required for:
- **EVM templates**: Uses `ethabi` crate which transitively depends on `ethereum-types` -> `getrandom`
- **Cosmos/Injective/Mantra templates**: Uses `cosmrs` crate which has similar WASM compatibility needs
- **Solana sol-anchor template**: Uses `anchor-lang` which requires it for WASM compatibility
- **Solana sol-hello-world template**: Incorrectly includes getrandom (no ethabi dependency)

Templates that should NOT have getrandom:
- `sol-transactions` (correct - no getrandom)
- `stellar-minimal` (correct - no getrandom)
- `stellar-transactions-operations` (correct - no getrandom)
- `tron-hello-world` (correct - no getrandom)
- `tron-transactions` (correct - no getrandom)
- `starknet-hello-world` (correct - no getrandom)
- `starknet-events` (correct - no getrandom)
- `near-hello-world` (correct - no getrandom)

Templates that incorrectly have getrandom:
- `sol-hello-world` - Has getrandom but does NOT use ethabi

**Note**: The sol-anchor template's getrandom dependency will be verified through integration tests.

---

## Step 1: Update Root Cargo.toml

This file is used for development/testing in the repo root.

**File**: `./Cargo.toml`

**Current content excerpt**:
```toml
[dependencies]
substreams = "0.6"
substreams-solana = "0.14"
substreams-solana-program-instructions = "0.2"
substreams-database-change = "2"
substreams-entity-change = "2"
```

**Updated content**:
```toml
[dependencies]
hex-literal = "0.3.4"
num-bigint = "0.4"
num-traits = "0.2.15"
prost = "0.13"
prost-types = "0.13"
substreams = "0.7"
substreams-solana = "0.14"
substreams-solana-program-instructions = "0.3"
substreams-database-change = "3"
substreams-entity-change = "2"
anchor-lang = ">=0.30.1"
sologger_log_context = "0.1.2"
base64 = "0.22.1"

# Required so that ethabi > ethereum-types build correctly under wasm32-unknown-unknown
[target.wasm32-unknown-unknown.dependencies]
getrandom = { version = "0.2", features = ["custom"] }
```

**Verification**:
```bash
cargo check
```

---

## Step 2: Update EVM Templates

### 2.1 Update evm-events-calls/templates/Cargo.toml.gotmpl

**File**: `./evm-events-calls/templates/Cargo.toml.gotmpl`

**Changes**:
- `substreams`: `"0.6.0"` -> `"0.7"`
- `substreams-ethereum`: `"0.10.2"` or `"0.10.0"` -> `"0.11"` (both dependencies and build-dependencies)
- `prost`: `"0.13.3"` -> `"0.13"`
- `prost-types`: `"0.13.3"` -> `"0.13"`

**Updated content**:
```toml
[package]
name = "{{ .Name }}"
version = "0.0.1"
edition = "2021"

[lib]
name = "substreams"
crate-type = ["cdylib"]

[dependencies]
ethabi = "17"
hex-literal = "0.3.4"
num-bigint = "0.4"
num-traits = "0.2.15"
prost = "0.13"
prost-types = "0.13"
substreams = "0.7"
substreams-ethereum = "0.11"

# Required so that ethabi > ethereum-types build correctly under wasm32-unknown-unknown
[target.wasm32-unknown-unknown.dependencies]
getrandom = { version = "0.2", features = ["custom"] }

[build-dependencies]
anyhow = "1"
substreams-ethereum = "0.11"
regex = "1.8"

[profile.release]
lto = true
opt-level = 's'
strip = "debuginfo"
```

### 2.2 Update evm-events-calls/templates/substreams.yaml.gotmpl

**File**: `./evm-events-calls/templates/substreams.yaml.gotmpl`

**Changes**:
- `ethereum-common`: `v0.3.0` -> `v0.3.3`

**Updated line 7**:
```yaml
    ethcommon: https://spkg.io/streamingfast/ethereum-common-v0.3.3.spkg
```

### 2.3 Update evm-events-calls/templates/sql/substreams.yaml.gotmpl

**File**: `./evm-events-calls/templates/sql/substreams.yaml.gotmpl`

**Changes**:
- `ethereum-common`: `v0.3.0` -> `v0.3.3`

**Updated line 9**:
```yaml
  ethcommon: https://spkg.io/streamingfast/ethereum-common-v0.3.3.spkg
```

### 2.4 Update evm-events-calls-raw/templates/Cargo.toml.gotmpl

**File**: `./evm-events-calls-raw/templates/Cargo.toml.gotmpl`

**Changes**:
- `substreams`: `"0.6.0"` -> `"0.7"`
- `substreams-ethereum`: `"0.10.2"` or `"0.10.0"` -> `"0.11"` (both dependencies and build-dependencies)
- `prost`: `"0.13.3"` -> `"0.13"`
- `prost-types`: `"0.13.3"` -> `"0.13"`

### 2.5 Update evm-events-calls-raw/templates/substreams.yaml.gotmpl

**File**: `./evm-events-calls-raw/templates/substreams.yaml.gotmpl`

**Changes**:
- `ethereum-common`: `v0.3.0` -> `v0.3.3`

### 2.6 Update evm-events-calls-raw/templates/sql/substreams.yaml.gotmpl

**File**: `./evm-events-calls-raw/templates/sql/substreams.yaml.gotmpl`

**Changes**:
- `ethereum-common`: `v0.3.0` -> `v0.3.3`

### 2.7 Update evm-hello-world/templates/Cargo.toml.gotmpl

**File**: `./evm-hello-world/templates/Cargo.toml.gotmpl`

**Changes**:
- `substreams`: `"0.6.0"` -> `"0.7"`
- `substreams-ethereum`: `"0.10.2"` or `"0.10.0"` -> `"0.11"` (both dependencies and build-dependencies)
- `prost`: `"0.13.3"` -> `"0.13"`
- `prost-types`: `"0.13.3"` -> `"0.13"`

### 2.8 Update evm-hello-world/templates/substreams.yaml.gotmpl

**File**: `./evm-hello-world/templates/substreams.yaml.gotmpl`

**Changes** (line 31 comment):
- Update comment reference from `v0.3.0` to `v0.3.3`

---

## Step 3: Update Solana Templates

### 3.1 Update sol-hello-world/templates/Cargo.toml.gotmpl

**File**: `./sol-hello-world/templates/Cargo.toml.gotmpl`

**Changes**:
- `substreams`: `"0.6.0"` -> `"0.7"`
- `substreams-solana`: `"0.14.1"` -> `"0.14"`
- `substreams-solana-program-instructions`: `"0.2.0"` -> `"0.3"`
- `prost`: `"0.13.3"` -> `"0.13"`
- `prost-types`: `"0.13.3"` -> `"0.13"`
- **Remove** the getrandom section (not needed - no ethabi dependency)

**Updated content**:
```toml
[package]
name = "{{ .Name }}"
version = "0.0.1"
edition = "2021"

[lib]
name = "substreams"
crate-type = ["cdylib"]

[dependencies]
hex-literal = "0.3.4"
num-bigint = "0.4"
num-traits = "0.2.15"
prost = "0.13"
prost-types = "0.13"
substreams = "0.7"
substreams-solana = "0.14"
substreams-solana-program-instructions = "0.3"

[build-dependencies]
anyhow = "1"
regex = "1.8"

[profile.release]
lto = true
opt-level = 's'
strip = "debuginfo"
```

### 3.2 Update sol-hello-world/templates/substreams.yaml.gotmpl

**File**: `./sol-hello-world/templates/substreams.yaml.gotmpl`

**Changes**:
- `solana-common`: `v0.3.0` -> `v0.3.3`

**Updated line 7**:
```yaml
  solana: https://spkg.io/streamingfast/solana-common-v0.3.3.spkg
```

### 3.3 Update sol-transactions/templates/Cargo.toml.gotmpl

**File**: `./sol-transactions/templates/Cargo.toml.gotmpl`

**Changes**:
- `substreams`: `"0.6.0"` -> `"0.7"`
- `substreams-solana`: `"0.14.1"` -> `"0.14"`
- `prost`: `"0.13.3"` -> `"0.13"`
- `prost-types`: `"0.13.3"` -> `"0.13"`

**Updated content**:
```toml
[package]
name = "{{ .Name }}"
version = "0.0.1"
edition = "2021"

[lib]
name = "substreams"
crate-type = ["cdylib"]

[dependencies]
prost = "0.13"
prost-types = "0.13"
substreams = "0.7"
substreams-solana = "0.14"

[profile.release]
lto = true
opt-level = 's'
strip = "debuginfo"
```

### 3.4 Update sol-anchor/templates/Cargo.toml.gotmpl

**File**: `./sol-anchor/templates/Cargo.toml.gotmpl`

**Changes**:
- `substreams`: `"0.6"` -> `"0.7"`
- `substreams-solana`: already `"0.14"` (correct format, keep as-is)
- `substreams-solana-program-instructions`: `"0.2"` -> `"0.3"`
- **Keep** getrandom (anchor-lang needs it for WASM compatibility)

**Updated content**:
```toml
[package]
name = "{{ .Name }}"
version = "0.0.1"
edition = "2021"

[lib]
name = "substreams"
crate-type = ["cdylib"]

[dependencies]
prost = "0.13"
prost-types = "0.13"
substreams = "0.7"
substreams-solana = "0.14"
substreams-solana-program-instructions = "0.3"
anchor-lang = ">=0.31.0"
sologger_log_context = "0.1.2"
base64 = "0.22.1"

# Required so that anchor-lang dependencies build correctly under wasm32-unknown-unknown
[target.wasm32-unknown-unknown.dependencies]
getrandom = { version = "0.2", features = ["custom"] }

[profile.release]
lto = true
opt-level = 's'
strip = "debuginfo"
```

### 3.5 Update sol-anchor/templates/substreams.yaml.gotmpl

**File**: `./sol-anchor/templates/substreams.yaml.gotmpl`

**Changes**:
- `solana-common`: `v0.3.0` -> `v0.3.3`

**Updated line 7**:
```yaml
  solana: https://spkg.io/streamingfast/solana-common-v0.3.3.spkg
```

---

## Step 4: Update Stellar Templates

### 4.1 Update stellar-minimal/templates/Cargo.toml.gotmpl

**File**: `./stellar-minimal/templates/Cargo.toml.gotmpl`

**Changes**:
- `substreams`: `"0.6.0"` -> `"0.7"`
- `prost`: `"0.13.3"` -> `"0.13"` (if present)
- `prost-types`: `"0.13.3"` -> `"0.13"` (if present)

### 4.2 Update stellar-minimal/templates/substreams.yaml.gotmpl

**File**: `./stellar-minimal/templates/substreams.yaml.gotmpl`

**Changes**:
- `stellar-foundational`: `v0.3.0` -> `v0.5.0`

**Updated line 7**:
```yaml
  stellar: https://github.com/streamingfast/substreams-foundational-modules/releases/download/stellar-foundational-v0.5.0/stellar-foundational-v0.5.0.spkg
```

### 4.3 Update stellar-transactions-operations/templates/Cargo.toml.gotmpl

**File**: `./stellar-transactions-operations/templates/Cargo.toml.gotmpl`

**Changes**:
- `substreams`: `"0.6.0"` -> `"0.7"`
- `prost`: `"0.13.3"` -> `"0.13"` (if present)
- `prost-types`: `"0.13.3"` -> `"0.13"` (if present)

### 4.4 stellar-transactions-operations/templates/substreams.yaml.gotmpl

**File**: `./stellar-transactions-operations/templates/substreams.yaml.gotmpl`

**No changes needed** - already uses `stellar-foundational@v0.5.0` (short format)

---

## Step 5: Update Starknet Templates

### 5.1 Update starknet-hello-world/templates/Cargo.toml.gotmpl

**File**: `./starknet-hello-world/templates/Cargo.toml.gotmpl`

**Changes**:
- `substreams`: `"0.6.0"` -> `"0.7"`
- `prost`: `"0.13.3"` -> `"0.13"` (if present)
- `prost-types`: `"0.13.3"` -> `"0.13"` (if present)

### 5.2 starknet-hello-world/templates/substreams.yaml.gotmpl

**No changes needed** - already at `starknet-foundational-v0.1.4`

### 5.3 Update starknet-events/templates/Cargo.toml.gotmpl

**File**: `./starknet-events/templates/Cargo.toml.gotmpl`

**Changes**:
- `substreams`: `"0.6.0"` -> `"0.7"`
- `prost`: `"0.13.3"` -> `"0.13"` (if present)
- `prost-types`: `"0.13.3"` -> `"0.13"` (if present)

### 5.4 starknet-events/templates/substreams.yaml.gotmpl

**No changes needed** - already at `starknet-foundational-v0.1.4`

---

## Step 6: Update TRON Templates

### 6.1 Update tron-hello-world/templates/Cargo.toml.gotmpl

**File**: `./tron-hello-world/templates/Cargo.toml.gotmpl`

**Changes**:
- `substreams`: `"0.6.0"` -> `"0.7"`
- `prost`: `"0.13.3"` -> `"0.13"` (if present)
- `prost-types`: `"0.13.3"` -> `"0.13"` (if present)

### 6.2 tron-hello-world/templates/substreams.yaml.gotmpl

**No changes needed** - already at `tron_foundational/v0.1.2`

### 6.3 Update tron-transactions/templates/Cargo.toml.gotmpl

**File**: `./tron-transactions/templates/Cargo.toml.gotmpl`

**Changes**:
- `substreams`: `"0.6.0"` -> `"0.7"`
- `prost`: `"0.13.3"` -> `"0.13"` (if present)
- `prost-types`: `"0.13.3"` -> `"0.13"` (if present)

### 6.4 tron-transactions/templates/substreams.yaml.gotmpl

**No changes needed** - already at `tron_foundational/v0.1.2`

---

## Step 7: Update NEAR Templates

### 7.1 Update near-hello-world/templates/Cargo.toml.gotmpl

**File**: `./near-hello-world/templates/Cargo.toml.gotmpl`

**Changes**:
- `substreams`: `"0.6"` -> `"0.7"`
- `substreams-near`: already `"0.10"` (correct format, keep as-is)
- `prost`: already `"0.13"` (correct format, keep as-is)
- `prost-types`: already `"0.13"` (correct format, keep as-is)

**Updated content**:
```toml
[package]
name = "{{ .Name }}"
version = "0.1.0"
edition = "2021"

[lib]
name = "substreams"
crate-type = ["cdylib"]

[dependencies]
substreams = "0.7"
substreams-near = "0.10"
prost = "0.13"
prost-types = "0.13"

```

---

## Step 8: Update Cosmos-Based Templates (Injective, Mantra)

### 8.1 Update injective-events/templates/Cargo.toml.gotmpl

**File**: `./injective-events/templates/Cargo.toml.gotmpl`

**Changes**:
- `substreams`: `"0.6.0"` -> `"0.7"`
- `prost`: `"0.13.3"` -> `"0.13"`
- `prost-types`: `"0.13.3"` -> `"0.13"`
- `cosmrs`: `"0.16.0"` -> `"0.16"`
- Update getrandom comment to be accurate

**Updated content**:
```toml
[package]
name = "{{ .Name }}"
version = "0.0.1"
edition = "2021"

[lib]
name = "substreams"
crate-type = ["cdylib"]

[dependencies]
prost = "0.13"
prost-types = "0.13"
substreams = "0.7"
cosmrs = { version = "0.16", features = ["cosmwasm"] }

# Required so that cosmrs dependencies build correctly under wasm32-unknown-unknown
[target.wasm32-unknown-unknown.dependencies]
getrandom = { version = "0.2", features = ["custom"] }

[profile.release]
lto = true
opt-level = 's'
strip = "debuginfo"
```

### 8.2 injective-events/templates/substreams.yaml.gotmpl

**No changes needed** - already at `injective-common-v0.2.5`

### 8.3 Update injective-hello-world/templates/Cargo.toml.gotmpl

**File**: `./injective-hello-world/templates/Cargo.toml.gotmpl`

**Changes**:
- `substreams`: `"0.6.0"` -> `"0.7"`
- `prost`: `"0.13.3"` -> `"0.13"` (if present)
- `prost-types`: `"0.13.3"` -> `"0.13"` (if present)
- `cosmrs`: `"0.16.0"` -> `"0.16"` (if present)
- Update getrandom comment (if present)

### 8.4 injective-hello-world/templates/substreams.yaml.gotmpl

**No changes needed** - already at `cosmos-v0.1.5`

### 8.5 Update mantra-events/templates/Cargo.toml.gotmpl

**File**: `./mantra-events/templates/Cargo.toml.gotmpl`

**Changes**:
- `substreams`: `"0.6.0"` -> `"0.7"`
- `prost`: `"0.13.3"` -> `"0.13"` (if present)
- `prost-types`: `"0.13.3"` -> `"0.13"` (if present)
- `cosmrs`: `"0.16.0"` -> `"0.16"` (if present)
- Update getrandom comment (if present)

### 8.6 mantra-events/templates/substreams.yaml.gotmpl

**No changes needed** - already at `mantra-common-v0.1.1`

### 8.7 Update mantra-hello-world/templates/Cargo.toml.gotmpl

**File**: `./mantra-hello-world/templates/Cargo.toml.gotmpl`

**Changes**:
- `substreams`: `"0.6.0"` -> `"0.7"`
- `prost`: `"0.13.3"` -> `"0.13"` (if present)
- `prost-types`: `"0.13.3"` -> `"0.13"` (if present)
- `cosmrs`: `"0.16.0"` -> `"0.16"` (if present)
- Update getrandom comment (if present)

### 8.8 mantra-hello-world/templates/substreams.yaml.gotmpl

**No changes needed** - already at `cosmos-v0.1.5`

---

## Verification Steps

After implementing all changes:

### 1. Go Unit Tests

```bash
go test ./... -v
```

Expected: All tests pass

### 2. Build Verification (Root Cargo.toml)

```bash
cargo check
```

Expected: Clean build with no errors

### 3. Template Syntax Verification

For each template directory, verify YAML syntax is valid:

```bash
# Check all YAML templates parse correctly
for f in $(find . -name "*.yaml.gotmpl"); do
  echo "Checking $f"
  # Basic syntax check - ensure no obvious issues
  cat "$f" | head -20
done
```

### 4. Integration Tests

**CRITICAL**: Run integration tests to verify all templates generate and build correctly:

```bash
RUN_INTEGRATION_TESTS=true INTEGRATION_TESTS_IN_DOCKER=true go test ./tests -v -parallel 4
```

**Expected**: All integration tests pass

**Notes for Remote Agent**:
- If you have access to run Docker internally, you should be able to run these tests
- The integration tests will verify that all generated code compiles and works correctly
- This is especially important for verifying the sol-anchor getrandom dependency
- If you cannot run Docker locally, you can monitor CI feedback after pushing changes and address any failures iteratively

**Fallback Strategy**:
If you cannot run the integration tests locally:
1. Push your changes to a branch
2. Monitor the CI pipeline for test results
3. If failures occur, analyze the CI logs and fix the issues
4. Iterate until all tests pass

---

## Project Standards Checklist

- [ ] All Cargo.toml.gotmpl files updated with new substreams version (0.7.0)
- [ ] All chain-specific crates updated (substreams-ethereum 0.11.1, substreams-solana 0.14.2, etc.)
- [ ] substreams-solana-program-instructions bumped to 0.3.0
- [ ] substreams-database-change bumped to 3.0.0
- [ ] Foundational module SPKGs updated (ethereum-common v0.3.3, solana-common v0.3.3, stellar-foundational v0.5.0)
- [ ] getrandom removed from sol-hello-world (not needed)
- [ ] getrandom comments updated to be accurate (ethabi -> cosmrs where applicable)
- [ ] All Go tests pass
- [ ] Root Cargo.toml compiles

---

## Commit Strategy

Create a single commit with all changes:

```
Update Substreams dependencies to latest versions

- Bump substreams crate from 0.6.0 to 0.7.0 across all templates
- Bump substreams-ethereum from 0.10.x to 0.11.1
- Bump substreams-solana from 0.14.1 to 0.14.2
- Bump substreams-solana-program-instructions from 0.2.x to 0.3.0
- Bump substreams-database-change from 2.x to 3.0.0
- Update ethereum-common SPKG from v0.3.0 to v0.3.3
- Update solana-common SPKG from v0.3.0 to v0.3.3
- Update stellar-foundational SPKG from v0.3.0 to v0.5.0
- Remove unnecessary getrandom from sol-hello-world template
- Fix getrandom comments for Cosmos-based templates
```

---

## Summary of Files to Modify

### Cargo.toml.gotmpl files (17 files):
1. `./Cargo.toml` (root)
2. `./evm-events-calls/templates/Cargo.toml.gotmpl`
3. `./evm-events-calls-raw/templates/Cargo.toml.gotmpl`
4. `./evm-hello-world/templates/Cargo.toml.gotmpl`
5. `./sol-hello-world/templates/Cargo.toml.gotmpl`
6. `./sol-transactions/templates/Cargo.toml.gotmpl`
7. `./sol-anchor/templates/Cargo.toml.gotmpl`
8. `./stellar-minimal/templates/Cargo.toml.gotmpl`
9. `./stellar-transactions-operations/templates/Cargo.toml.gotmpl`
10. `./starknet-hello-world/templates/Cargo.toml.gotmpl`
11. `./starknet-events/templates/Cargo.toml.gotmpl`
12. `./tron-hello-world/templates/Cargo.toml.gotmpl`
13. `./tron-transactions/templates/Cargo.toml.gotmpl`
14. `./near-hello-world/templates/Cargo.toml.gotmpl`
15. `./injective-events/templates/Cargo.toml.gotmpl`
16. `./injective-hello-world/templates/Cargo.toml.gotmpl`
17. `./mantra-events/templates/Cargo.toml.gotmpl`
18. `./mantra-hello-world/templates/Cargo.toml.gotmpl`

### substreams.yaml.gotmpl files (8 files need updates):
1. `./evm-events-calls/templates/substreams.yaml.gotmpl`
2. `./evm-events-calls/templates/sql/substreams.yaml.gotmpl`
3. `./evm-events-calls-raw/templates/substreams.yaml.gotmpl`
4. `./evm-events-calls-raw/templates/sql/substreams.yaml.gotmpl`
5. `./evm-hello-world/templates/substreams.yaml.gotmpl` (comment only)
6. `./sol-hello-world/templates/substreams.yaml.gotmpl`
7. `./sol-anchor/templates/substreams.yaml.gotmpl`
8. `./stellar-minimal/templates/substreams.yaml.gotmpl`

---

## Key Implementation Notes

### Note 1: prost Version Strategy

The codebase currently uses `prost = "0.13.3"`. While `prost 0.14.x` exists, upgrading requires careful consideration:
- The substreams crates may not yet support prost 0.14
- Keep at 0.13.x unless substreams 0.7.0 explicitly requires 0.14

### Note 2: substreams-ethereum Build Dependencies

In EVM templates, `substreams-ethereum` appears in both `[dependencies]` and `[build-dependencies]`. Both must be updated to the same version (0.11.1).

### Note 3: Version Format for Rust Crate Dependencies

**IMPORTANT**: All Rust crate dependencies should use `<major>.<minor>` format (NOT `<major>.<minor>.<patch>`):
- `substreams 0.7.0` → `"0.7"` (not `"0.7.0"`)
- `substreams-solana 0.14.2` → `"0.14"` (not `"0.14.2"`)
- `substreams-ethereum 0.11.1` → `"0.11"` (not `"0.11.1"`)
- `substreams-database-change 3.0.0` → `"3.0"` (not `"3.0.0"`)
- `prost 0.13.3` → `"0.13"` (not `"0.13.3"`)
- `prost-types 0.13.3` → `"0.13"` (not `"0.13.3"`)

This ensures proper semver compatibility while allowing automatic patch updates for dependencies.

### Note 4: getrandom Dependency

The getrandom dependency comment should accurately reflect why it's needed:
- For EVM templates: "Required so that ethabi > ethereum-types build correctly under wasm32-unknown-unknown"
- For Cosmos templates: "Required so that cosmrs dependencies build correctly under wasm32-unknown-unknown"
- For sol-anchor: "Required so that anchor-lang dependencies build correctly under wasm32-unknown-unknown"
