---
name: OpenSpec: Review
description: Review and evaluate OpenSpec changes for technical quality, completeness, and alignment with project standards.
category: OpenSpec
tags: [openspec, review]
---
<!-- OPENSPEC:START -->
**Guardrails**
- Favor straightforward, minimal implementations first and add complexity only when it is requested or clearly required.
- Keep changes tightly scoped to the requested outcome.
- Refer to `openspec/AGENTS.md` (located inside the `openspec/` directory—run `ls openspec` or `openspec update` if you don't see it) if you need additional OpenSpec conventions or clarifications.
- Provide constructive, actionable feedback that helps improve the proposal while maintaining project standards.

**Steps**
Track these steps as TODOs and complete them one by one.
1. Identify the change ID to review (via the prompt or `openspec list`).
2. Read `changes/<id>/proposal.md`, `design.md` (if present), `tasks.md`, and all spec deltas under `changes/<id>/specs/`.
3. Review relevant design documents in `docs/design/` to understand current architecture and ensure alignment with existing patterns.
4. Validate the proposal structure with `openspec validate <id> --strict` and ensure all validation issues are addressed.
5. Evaluate technical alignment with project architecture, conventions, and domain context.
6. Review the completeness and clarity of requirements, scenarios, and acceptance criteria.
7. Assess task breakdown for feasibility, proper sequencing, and measurable deliverables.
8. Check for consistency with existing specs and identify potential conflicts or overlaps.
9. Provide comprehensive review feedback with specific recommendations for improvement or approval.

**Review Criteria**
- **Technical Quality**: Aligns with Go conventions, project architecture, and tech stack choices
- **Completeness**: All necessary components (proposal, tasks, specs) are present and well-defined
- **Clarity**: Requirements and scenarios are unambiguous and testable
- **Feasibility**: Tasks are properly sized, sequenced, and technically achievable
- **Consistency**: Maintains alignment with existing codebase and project conventions
- **Risk Assessment**: Identifies potential technical risks and provides mitigation strategies

**Reference**
- Use `openspec show <id> --json --deltas-only` to inspect detailed change information
- Search existing requirements with `rg -n "Requirement:|Scenario:" openspec/specs` to identify overlaps
- Review project conventions in `openspec/project.md` and `openspec/AGENTS.md`
- Use `openspec list --specs` to understand current specification landscape
<!-- OPENSPEC:END -->