# Learnings - Output Flag Redesign

## Conventions & Patterns

(Agents will append findings here)

## Task 1: Remove FormatFields Constant

**Date:** 2026-02-09

**Changes Made:**
- Removed `FormatFields Format = "fields"` constant from `pkg/output/format.go`
- Removed `FormatFields` from `allFormats` slice in `pkg/output/format.go`
- Removed `case output.FormatFields.String():` and `printFields()` call from `cmd/get.go`
- Deleted `printFields()` function (lines 311-328) from `cmd/get.go`
- Updated error message in `cmd/get.go` to remove "fields" from valid formats list

**Files Modified:**
- `pkg/output/format.go`: Removed constant and allFormats entry
- `cmd/get.go`: Removed case statement and printFields function

**Verification:**
- `grep -r "FormatFields"` returns zero matches ✓
- `grep -r "printFields"` returns zero matches ✓
- `go build ./...` compiles successfully ✓

**Notes:**
- The `NormalizeFormat` function still has a fallback for "fields" → "table" mapping (line 84)
- This is intentional for backward compatibility and will be addressed in Task 3
- No issues encountered during removal - clean deletion


**Additional Changes (Test Files):**
- Removed `TestPrintFields` test function from `cmd/get_test.go`
- Updated `TestFormatConstants` in `pkg/output/output_test.go` to remove FormatFields assertion
- Updated `TestFormatIsValid` in `pkg/output/output_test.go` to remove FormatFields test case
- Updated `TestParseFormat` in `pkg/output/output_test.go` to remove "fields" test case
- Updated `TestAllFormats` in `pkg/output/output_test.go` to expect 5 formats instead of 6

**Final Verification Results:**
- `grep -r "FormatFields" pkg/ cmd/` → No matches ✓
- `grep -r "printFields" cmd/` → No matches ✓
- `go build ./...` → Success ✓
- `go test ./...` → All tests pass ✓

**Commit:**
- SHA: 5d932e0
- Message: "refactor(output): remove fields format (redundant with simple --show-metadata)"
- Files changed: 4 (pkg/output/format.go, cmd/get.go, pkg/output/output_test.go, cmd/get_test.go)
- Lines removed: 74 (constant, case statement, printFields function, tests)


## Task 2: Fix Watch Command to Use Global -o Flag

**Date:** 2026-02-09

**Changes Made:**
- Removed `json bool` field from `watchOpts` struct in `cmd/watch.go`
- Removed `BoolVarP` flag registration for `-o, --output` from `watchCmd.Flags()`
- Updated `runWatch()` to check global `outputFormat` variable instead of `watchOpts.json`
- Refactored `printWatchEvent()` to implement two output modes:
  - `outputFormat == "json"`: Print full JSON event structure
  - `outputFormat != "json"` (simple/default): Print raw value only (matches etcdctl)
- Updated watch command's Long description to mention `-o` flag support

**Files Modified:**
- `cmd/watch.go`: Removed boolean flag, updated to use global outputFormat

**Verification:**
- `grep -n "watchOpts.json" cmd/watch.go` → Zero matches ✓
- `grep -n 'BoolVarP.*output' cmd/watch.go` → Zero matches ✓
- `go build ./...` → Compiles successfully ✓

**Implementation Details:**
- Simple format now prints ONLY the raw value (one line per event)
- JSON format prints full event structure with type, key, value, revision, etc.
- Matches etcdctl watch behavior: simple = value-only, json = full event
- Removed verbose output format (was: `[PUT] rev=123 /key\n  value: foo`)
- New simple format: just `foo` (raw value only)

**Notes:**
- Watch command now uses global `-o` flag consistently with other commands
- No flag conflict - watch-specific `-o` flag removed
- Simple format is cleaner and matches etcdctl UX
- JSON format provides full event details for scripting/automation


## Task 4: Replace NormalizeFormat with ValidateFormat

### Changes Made
- Replaced `NormalizeFormat()` in `pkg/output/format.go` with `ValidateFormat(requested string, allowed []string) error`
- Removed silent fallback logic (tree→table, fields→table)
- Updated `normalizeOutputFormat()` in `cmd/helpers.go` to `validateOutputFormat()` that calls `ValidateFormat`
- Removed `formatsWithTree` and `formatsWithoutTree` helper variables from `cmd/helpers.go`
- Added per-command format validation with explicit allowed formats:
  - `get`: [simple, json, yaml, table, tree]
  - `parse`: [simple, json, yaml, table, tree]
  - `apply`: [simple, json, table]
  - `validate`: [simple, json, table]
  - `config get-contexts`: [simple, json, table]
  - `config view`: [json, yaml, table]
  - `version`: [simple, json]
  - `watch`: [simple, json]

### Validation Behavior
- Format validation now happens at the start of each command's `runXxx()` function
- Invalid formats return clear error messages: `invalid format: fields (valid: simple, json, yaml, table, tree)`
- No silent fallbacks - users get explicit errors instead of warnings
- Each command validates against its specific allowed formats

### Files Modified
- `pkg/output/format.go`: Replaced NormalizeFormat with ValidateFormat
- `cmd/helpers.go`: Updated helper function, removed format slice variables
- `cmd/get.go`: Added format validation
- `cmd/parse.go`: Added format validation
- `cmd/apply.go`: Added format validation, replaced normalizedFormat with outputFormat
- `cmd/validate.go`: Added format validation
- `cmd/watch.go`: Added format validation
- `cmd/config.go`: Added format validation for get-contexts and view commands
- `cmd/version.go`: Added format validation

### Test File Status
- `pkg/output/format_test.go` still references old `NormalizeFormat()` function
- Tests will be updated in Task 5 (as per plan)
- Manual testing confirms validation works correctly

### Verification Results
✅ `grep -n "func NormalizeFormat" pkg/output/` → No matches (function removed)
✅ `grep -n "NormalizeFormat(" pkg/ cmd/` → Only test file references remain
✅ `grep -n "ValidateFormat" pkg/output/format.go` → New function exists
✅ `grep -n "formatsWithTree\|formatsWithoutTree" cmd/helpers.go` → Zero matches
✅ `./etu get /test -o fields` → Returns error: "invalid format: fields (valid: simple, json, yaml, table, tree)"
✅ `./etu apply -f test.txt -o tree` → Returns error: "invalid format: tree (valid: simple, json, table)"
✅ `go build ./...` → Compiles successfully

### Key Insights
- Validation at command entry point provides better UX (fail fast)
- Explicit per-command allowed formats make behavior predictable
- Error messages now include the full list of valid formats for each command
- Removed indirection (formatsWithTree/formatsWithoutTree) makes code more maintainable
