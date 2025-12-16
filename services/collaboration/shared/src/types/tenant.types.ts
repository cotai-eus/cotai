export interface TenantContext {
  tenantId: string;
  userId: string;
  userRoles: string[];
}

export enum TenantStatus {
  PROVISIONING = 'provisioning',
  ACTIVE = 'active',
  SUSPENDED = 'suspended',
  ARCHIVED = 'archived',
  DELETED = 'deleted',
}
