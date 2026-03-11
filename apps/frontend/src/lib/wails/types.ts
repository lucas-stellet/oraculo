export interface KnownProject {
  path: string;
  name: string;
}

export interface ProjectWithStatus {
  name: string;
  path: string;
  online: boolean;
  port?: number;
}
