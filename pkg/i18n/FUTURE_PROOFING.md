# Future-Proofing Status

This document tracks architectural decisions made to ensure the `pkg/i18n` package is resilient to future requirements.

**Last Updated:** 30 January 2026
**Status:** Core Complete

## 1. Extensibility: The "Magic Namespace" Problem

**Status: ✅ IMPLEMENTED**

*   **Solution:** The `KeyHandler` interface and middleware chain have been implemented.
*   **Details:** "Magic" keys like `i18n.label.*` are now handled by specific structs in `handler.go`. The core `T()` method iterates through these handlers.
*   **Benefit:** New patterns can be added via `AddHandler()` without modifying the core package.

## 2. API Design: Context & Disambiguation

**Status: ✅ IMPLEMENTED**

*   **Solution:** `TranslationContext` struct and `C()` builder have been created in `context.go`.
*   **Details:** `getEffectiveFormality()` now checks `*TranslationContext` for formality hints, in addition to `*Subject`.
*   **Benefit:** Translations can now be disambiguated via context, and formality can be set per-call without needing a Subject.

## 3. Storage: Interface-Driven Loading

**Status: ✅ IMPLEMENTED**

*   **Solution:** The `Loader` interface has been defined in `types.go`, and the default JSON logic moved to `FSLoader` in `loader.go`.
*   **Details:** `NewWithLoader` allows injecting any backend.
*   **Benefit:** Applications can now load translations from databases, remote APIs, or other file formats.

## 4. Standardization: Pluralization & CLDR

**Status: ⏳ PENDING**

*   **Current State:** The package still uses a custom `pluralRules` map in `types.go`.
*   **Recommendation:** When the need arises for more languages, replace the internal `pluralRules` map with a call to `golang.org/x/text/feature/plural` or a similar standard library wrapper. The current interface hides this implementation detail, so it can be swapped later without breaking changes.

## 5. Data Format: Vendor Compatibility

**Status: ⏳ ENABLED**

*   **Current State:** The default format is still the custom nested JSON.
*   **Future Path:** Thanks to the `Loader` interface, we can now implement a `PoLoader` or `ArbLoader` to support standard translation formats used by professional vendors, without changing the core service.

## Summary

The critical architectural risks (coupling, storage, and context) have been resolved. The remaining item (Pluralization standard) is an implementation detail that can be addressed incrementally without breaking the public API.