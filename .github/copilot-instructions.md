# GitHub Copilot Instructions

## Development Rules

### Version Management
- **IMPORTANT**: Whenever you are going to merge/push to the `main` branch (main branch), you MUST:
  1. Update the `QpVersion` in the `models/qp_defaults.go` file
  2. Increment the version following the current semantic pattern
  3. If it ends with `.0` it means stable version
  4. Development versions can use other suffixes

### Version Location
```go
// File: models/qp_defaults.go
const QpVersion = "3.25.2207.0127" // <-- ALWAYS UPDATE BEFORE MERGE TO MAIN
```

### Mandatory Process before Push/Merge to Main:
1. ✅ Verify that all changes are working properly
2. ✅ Run tests if they exist
3. ✅ **UPDATE QpVersion** in the `models/qp_defaults.go` file
4. ✅ Make commit with the new version
5. ✅ Then merge/push to main

### Version Increment Example:
- Current version: `3.25.2207.0127`
- Next version: `3.25.2207.0128` (simple increment)
- Or new version: `3.25.MMDD.HHMM` (based on current date/time)

## CRITICAL REMINDER
🚨 **NEVER merge to main without updating QpVersion** 🚨

This is a mandatory project rule for version control.
