# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v1.3.0] - 2025-12-05

### Added
- **Custom Error Types**: `ParseError` and `BindError` for better error handling
  - Both implement `Unwrap()` for use with `errors.Is()` and `errors.As()`
  - Helper functions: `NewParseError()`, `NewBindError()`, `IsParseError()`, `IsBindError()`
- **CI/CD Workflows**: GitHub Actions for automated testing and releases
  - `ci.yml`: Tests on Go 1.17-1.23, linting, security scanning, benchmarks
  - `release.yml`: Auto-generates releases when tags are pushed
- **Comprehensive Test Suite**:
  - `benchmark_test.go` with 26+ benchmarks covering all request types
  - `memory_test.go` for memory leak detection
  - Test coverage increased to 91.9%
- **Code Quality**: `.golangci.yml` linter configuration

### Changed
- **Optimized `replaceBracketKeyIntoDotKey`**: 
  - Moved `strings.Replacer` to package-level (reused across calls)
  - Added fast path for keys without brackets (0 allocations)
  - ~20ns for simple keys, ~75ns for bracket keys
- **Completely rewritten README.md**:
  - Added badges (Go version, license, Go Report Card)
  - Full API reference with tables
  - Documented `Cleanup()`, `FileHeaders`, `FormDataWithOptions()`, `Parse()`
  - Added type conversion reference
  - Added performance benchmarks section

### Fixed
- Fixed all linter warnings
- Fixed typos in documentation

## [v1.2.2] - 2025-03-28

### Fixed
- Bug fixes and improvements

## [v1.2.1] - 2025-03-28

### Fixed
- Minor bug fixes

## [v1.2.0] - 2025-03-28

### Added
- New features and improvements

## [v1.1.0] - 2025-03-28

### Added
- Additional functionality

## [v1.0.0] - 2025-03-28

### Added
- Initial stable release
- Form data parsing (multipart/form-data, application/x-www-form-urlencoded)
- JSON request parsing
- Query string parsing
- `ToBind()` for struct binding
- `ToMap()` for map conversion
- `ToJsonByte()` and `ToJsonString()` for JSON output
- Support for nested arrays and objects with bracket notation
- File upload support with `FileHeaders` type
- Automatic type conversion (string to int, float, bool)

[v1.3.0]: https://github.com/ezartsh/inrequest/compare/v1.2.2...v1.3.0
[v1.2.2]: https://github.com/ezartsh/inrequest/compare/v1.2.1...v1.2.2
[v1.2.1]: https://github.com/ezartsh/inrequest/compare/v1.2.0...v1.2.1
[v1.2.0]: https://github.com/ezartsh/inrequest/compare/v1.1.0...v1.2.0
[v1.1.0]: https://github.com/ezartsh/inrequest/compare/v1.0.0...v1.1.0
[v1.0.0]: https://github.com/ezartsh/inrequest/releases/tag/v1.0.0
