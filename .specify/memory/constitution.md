<!--
Sync Impact Report
- Version change: 0.0.0 → 1.0.0
- Modified principles: N/A (initial adoption)
- Added sections: Platform Constraints; Workflow Expectations
- Removed sections: None
- Templates requiring updates: .specify/templates/plan-template.md ✅; .specify/templates/spec-template.md ✅; .specify/templates/tasks-template.md ✅; .specify/templates/checklist-template.md ✅
- Follow-up TODOs: None
-->

# Online-PPT Constitution

## Core Principles

### P1 Slide Files Are Standalone HTML
- All presentation slides MUST reside under `presentations/<user-uuid>/<group>/slides` and follow the `slide-<n>.html` naming convention.
- Each slide MUST be a complete HTML document, declaring `<!DOCTYPE html>` and including all required styles or scripts inline or via project-approved assets.
- Slides MUST avoid runtime imports to untracked external URLs; shared assets MUST live under `public/` or presentation-local directories committed to version control.
Rationale: Enforces predictable slide discovery, keeps every slide self-contained for offline packaging, and prevents hidden runtime dependencies that break the static hosting model.

### P2 Config Drives Presentation State
- Every presentation MUST include `slides.config.json` as the single source of truth for order, visibility, metadata, and theme settings.
- Any change to slide availability, sequencing, or theme MUST be reflected in the corresponding config file within the same commit as the slide change.
- Config files MUST validate against the schema implied by shipped examples: `title`, `theme`, `settings`, and `slides` entries with `id`, `title`, `file`, and `visible` keys.
Rationale: Guarantees the UI and runtime stay synchronized with authored content and allows tooling to reason about presentation structure deterministically.

### P3 Vue 3 Frontend Discipline
- The runtime MUST remain a purely frontend Vue 3 + Vite application; new backend services or server-side rendering are prohibited.
- State management MUST use Pinia stores located under `src/stores`; shared logic belongs in `src/utils` or composables.
- Styling MUST leverage Tailwind CSS utilities layered with project styles (`src/tw.css`, `src/style.css`); global CSS additions outside these entry points require architectural review.
Rationale: Preserves the lightweight deployment model, keeps the codebase coherent around the selected stack, and avoids fragmentation that complicates maintenance.

### P4 Presentation UX Baseline
- Keyboard navigation (arrow keys, space, number keys, escape) MUST stay functional for every published presentation; regressions require immediate fixes.
- New slides MUST render responsively within a 16:9 viewport and maintain legible contrast against the active theme.
- Presenter-facing enhancements (notes, previews, progress) MUST degrade gracefully when unavailable and MUST NOT block baseline slide playback.
Rationale: Ensures presenters can rely on consistent controls and viewing quality regardless of content authorship, sustaining the framework's core user promise.

## Platform Constraints
- Supported runtime stack is `Vue 3`, `Vite`, `Pinia`, `Tailwind CSS`, and `@vueuse/core`; dependency additions require review for bundle size and browser compatibility.
- Build artifacts MUST be generated through `npm run build`; no custom bundlers or file watchers may replace the Vite pipeline.
- Public templates under `public/templates` MUST remain framework-agnostic HTML that adheres to Principle P1 when copied into slides.
- Asset naming MUST stay lowercase with hyphen separators to simplify static hosting and cross-platform tooling.

## Workflow Expectations
- Feature work MUST branch from `main` using the format `feature/<descriptor>` and link to a corresponding spec in `/specs` when generated.
- Pull requests MUST document which Core Principles were touched and show evidence of `npm run build` passing locally.
- Reviews MUST verify that slide changes include matching updates to `slides.config.json` and relevant store logic, or state why not applicable.
- Releases MUST record the supported slide groups and confirm that all slide HTML files load without console errors in the Vite preview.

## Governance
- This constitution supersedes conflicting guidance within the repository; deviations require a documented waiver recorded alongside the change.
- Amendments require consensus from active maintainers, an updated Sync Impact Report, and a semantic version bump recorded below.
- Compliance reviews occur at least once per quarter or prior to major feature releases, ensuring Core Principles remain enforced and tooling keeps pace.
- Violations discovered post-merge MUST be triaged within one business day and remediated or formally waived before the next release.

**Version**: 1.0.0 | **Ratified**: 2025-11-04 | **Last Amended**: 2025-11-04
