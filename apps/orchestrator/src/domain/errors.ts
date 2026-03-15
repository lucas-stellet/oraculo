export type DomainErrorCode =
  | "not_found"
  | "already_exists"
  | "invalid_transition"
  | "cyclic_dependency"
  | "missing_required"
  | "invalid_phase"
  | "approval_decided";

export class DomainError extends Error {
  constructor(
    public readonly code: DomainErrorCode,
    message: string,
    public readonly context?: Record<string, unknown>,
  ) {
    super(message);
    this.name = "DomainError";
  }
}

export const ErrNotFound = (entity: string, id: string | number) =>
  new DomainError("not_found", `${entity} '${id}' not found`);

export const ErrAlreadyExists = (entity: string, id: string | number) =>
  new DomainError("already_exists", `${entity} '${id}' already exists`);

export const ErrInvalidTransition = (from: string, to: string) =>
  new DomainError("invalid_transition", `cannot transition from '${from}' to '${to}'`);

export const ErrCyclicDependency = (from: string | number, to: string | number) =>
  new DomainError("cyclic_dependency", `adding ${from} → ${to} creates a cycle`);

export const ErrMissingRequired = (field: string) =>
  new DomainError("missing_required", `missing required field: ${field}`);

export const ErrInvalidPhase = (phase: string) =>
  new DomainError("invalid_phase", `invalid phase: ${phase}`);

export const ErrApprovalDecided = (id: string) =>
  new DomainError("approval_decided", `approval '${id}' already decided`);
