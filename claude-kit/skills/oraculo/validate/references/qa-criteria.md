# Validate QA Criteria

## Validation Scope

QA reviews:
- the diff
- the relevant specs
- the test results

QA does not review the implementer's private reasoning.

## Functional Correctness

Check whether the implementation satisfies the acceptance criteria and behaves correctly for the specified inputs and outputs.

## Standards Compliance

Check whether the implementation follows the project's conventions, structure, and established patterns.

## Edge Cases

Check empty states, boundaries, concurrency, and any scenario that could break the intended behavior.

## Test Quality

Check whether tests prove the required behavior rather than merely exercising the code superficially.

## Scope Adherence

Check whether the implementation stayed within the intended task and did not add unrelated features or sprawl.

## Severity Language

Use explicit severity:
- blocker: must be fixed before approval
- important: significant issue that should be fixed before approval
- minor: useful improvement but not necessarily a release blocker
