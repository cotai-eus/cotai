import { Injectable, OnModuleDestroy, Logger } from '@nestjs/common';
import { DataSource, DataSourceOptions } from 'typeorm';

export interface TenantConnectionConfig {
  type: 'postgres';
  host: string;
  port: number;
  database: string;
  username: string;
  password: string;
  ssl?: boolean | { rejectUnauthorized: boolean };
  maxConnections?: number;
  idleTimeoutMillis?: number;
}

@Injectable()
export class TenantConnectionService implements OnModuleDestroy {
  private readonly logger = new Logger(TenantConnectionService.name);
  private connections: Map<string, DataSource> = new Map();
  private baseConfig: TenantConnectionConfig;

  constructor(config: TenantConnectionConfig) {
    this.baseConfig = config;
  }

  /**
   * Get or create a database connection for the specified tenant
   * @param tenantId Tenant identifier
   * @returns DataSource configured for the tenant schema
   */
  async getConnection(tenantId: string): Promise<DataSource> {
    const schemaName = this.getSchemaName(tenantId);

    // Return existing connection if available
    if (this.connections.has(schemaName)) {
      const connection = this.connections.get(schemaName)!;
      if (connection.isInitialized) {
        return connection;
      }
    }

    // Create new connection
    const dataSourceOptions: DataSourceOptions = {
      type: this.baseConfig.type,
      host: this.baseConfig.host,
      port: this.baseConfig.port,
      database: this.baseConfig.database,
      username: this.baseConfig.username,
      password: this.baseConfig.password,
      schema: schemaName,
      ssl: this.baseConfig.ssl,
      extra: {
        max: this.baseConfig.maxConnections || 10,
        idleTimeoutMillis: this.baseConfig.idleTimeoutMillis || 30000,
      },
      synchronize: false, // Always use migrations in production
      logging: process.env.NODE_ENV === 'development' ? ['error', 'warn'] : false,
    };

    const dataSource = new DataSource(dataSourceOptions);
    await dataSource.initialize();

    // Set search_path for the connection
    await dataSource.query(`SET search_path TO "${schemaName}", public`);

    this.connections.set(schemaName, dataSource);
    this.logger.log(`Created connection for tenant: ${tenantId} (schema: ${schemaName})`);

    return dataSource;
  }

  /**
   * Execute a callback within a tenant context
   * Ensures proper schema switching and RLS enforcement
   * @param tenantId Tenant identifier
   * @param callback Function to execute with tenant DataSource
   * @returns Result of the callback execution
   */
  async executeInTenantContext<T>(
    tenantId: string,
    callback: (dataSource: DataSource) => Promise<T>,
  ): Promise<T> {
    const connection = await this.getConnection(tenantId);

    // Set tenant context for RLS policies
    await connection.query(`SET app.current_tenant = '${tenantId}'`);

    try {
      return await callback(connection);
    } finally {
      // Clean up session variables
      await connection.query(`RESET app.current_tenant`);
    }
  }

  /**
   * Execute a transaction within tenant context
   * @param tenantId Tenant identifier
   * @param callback Function to execute within transaction
   * @returns Result of the callback execution
   */
  async executeInTransaction<T>(
    tenantId: string,
    callback: (queryRunner: any) => Promise<T>,
  ): Promise<T> {
    const connection = await this.getConnection(tenantId);
    const queryRunner = connection.createQueryRunner();

    await queryRunner.connect();
    await queryRunner.startTransaction();

    try {
      // Set tenant context for RLS
      await queryRunner.query(`SET app.current_tenant = '${tenantId}'`);

      const result = await callback(queryRunner);

      await queryRunner.commitTransaction();
      return result;
    } catch (error) {
      await queryRunner.rollbackTransaction();
      throw error;
    } finally {
      await queryRunner.release();
    }
  }

  /**
   * Get schema name for a tenant
   * @param tenantId Tenant identifier
   * @returns Schema name (tenant_{uuid_without_hyphens})
   */
  private getSchemaName(tenantId: string): string {
    // Remove hyphens from UUID for schema name
    const cleanId = tenantId.replace(/-/g, '');
    return `tenant_${cleanId}`;
  }

  /**
   * Close all tenant connections
   */
  async onModuleDestroy() {
    this.logger.log('Closing all tenant database connections...');

    for (const [schema, connection] of this.connections.entries()) {
      try {
        await connection.destroy();
        this.logger.log(`Closed connection for schema: ${schema}`);
      } catch (error) {
        this.logger.error(`Error closing connection for ${schema}:`, error);
      }
    }

    this.connections.clear();
  }

  /**
   * Get connection statistics
   */
  getStats() {
    return {
      activeConnections: this.connections.size,
      schemas: Array.from(this.connections.keys()),
    };
  }
}
