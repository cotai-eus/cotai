import { Repository, DataSource, FindOptionsWhere, FindManyOptions } from 'typeorm';
import { Logger } from '@nestjs/common';

/**
 * Base repository with tenant isolation
 * All queries automatically include tenant_id filter
 */
export abstract class BaseTenantRepository<T extends { tenant_id?: string }> {
  protected readonly logger: Logger;
  protected abstract entityClass: new () => T;

  constructor(protected readonly dataSource: DataSource) {
    this.logger = new Logger(this.constructor.name);
  }

  /**
   * Get TypeORM repository instance
   */
  protected getRepository(): Repository<T> {
    return this.dataSource.getRepository(this.entityClass);
  }

  /**
   * Find all entities for a tenant
   * @param tenantId Tenant identifier
   * @param options Additional find options
   */
  async findByTenant(
    tenantId: string,
    options?: Omit<FindManyOptions<T>, 'where'>,
  ): Promise<T[]> {
    const repo = this.getRepository();
    return repo.find({
      ...options,
      where: { tenant_id: tenantId } as FindOptionsWhere<T>,
    });
  }

  /**
   * Find one entity by ID for a tenant
   * @param tenantId Tenant identifier
   * @param id Entity ID
   */
  async findOneByTenant(tenantId: string, where: FindOptionsWhere<T>): Promise<T | null> {
    const repo = this.getRepository();
    return repo.findOne({
      where: { tenant_id: tenantId, ...where } as FindOptionsWhere<T>,
    });
  }

  /**
   * Create a new entity for a tenant
   * @param tenantId Tenant identifier
   * @param data Entity data
   */
  async createForTenant(tenantId: string, data: Partial<T>): Promise<T> {
    const repo = this.getRepository();
    const entity = repo.create({ tenant_id: tenantId, ...data } as any);
    return repo.save(entity) as unknown as Promise<T>;
  }

  /**
   * Update an entity for a tenant
   * @param tenantId Tenant identifier
   * @param id Entity ID
   * @param data Update data
   */
  async updateForTenant(
    tenantId: string,
    where: FindOptionsWhere<T>,
    data: Partial<T>,
  ): Promise<T | null> {
    const repo = this.getRepository();
    const entity = await this.findOneByTenant(tenantId, where);

    if (!entity) {
      return null;
    }

    const updated = repo.merge(entity, data as any);
    return repo.save(updated);
  }

  /**
   * Delete an entity for a tenant
   * @param tenantId Tenant identifier
   * @param where Find conditions
   */
  async deleteForTenant(tenantId: string, where: FindOptionsWhere<T>): Promise<boolean> {
    const repo = this.getRepository();
    const result = await repo.delete({
      tenant_id: tenantId,
      ...where,
    } as any);

    return (result.affected || 0) > 0;
  }

  /**
   * Count entities for a tenant
   * @param tenantId Tenant identifier
   * @param where Additional conditions
   */
  async countByTenant(tenantId: string, where?: FindOptionsWhere<T>): Promise<number> {
    const repo = this.getRepository();
    return repo.count({
      where: { tenant_id: tenantId, ...where } as FindOptionsWhere<T>,
    });
  }

  /**
   * Check if entity exists for a tenant
   * @param tenantId Tenant identifier
   * @param where Find conditions
   */
  async existsByTenant(tenantId: string, where: FindOptionsWhere<T>): Promise<boolean> {
    const count = await this.countByTenant(tenantId, where);
    return count > 0;
  }
}
