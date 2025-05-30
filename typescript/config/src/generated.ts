// Code generated by tygo. DO NOT EDIT.
import type { 
  LogLevel,
  Duration,
  ByteSize,
  ConnectorDriverType as TsConnectorDriverType,
  ConnectorConfig as TsConnectorConfig,
  UpstreamType as TsUpstreamType,
  NetworkArchitecture as TsNetworkArchitecture,
  AuthType as TsAuthType,
  AuthStrategyConfig as TsAuthStrategyConfig,
  SelectionPolicyEvalFunction
} from "./types"

//////////
// source: architecture_evm.go

export const UpstreamTypeEvm: UpstreamType = "evm";
export type EvmUpstream = 
    Upstream;
export type EvmNodeType = string;
export const EvmNodeTypeUnknown: EvmNodeType = "unknown";
export const EvmNodeTypeFull: EvmNodeType = "full";
export const EvmNodeTypeArchive: EvmNodeType = "archive";
export type EvmSyncingState = number /* int */;
export const EvmSyncingStateUnknown: EvmSyncingState = 0;
export const EvmSyncingStateSyncing: EvmSyncingState = 1;
export const EvmSyncingStateNotSyncing: EvmSyncingState = 2;
export type EvmStatePoller = any;

//////////
// source: cache_dal.go

export type CacheDAL = any;

//////////
// source: cache_mock.go

export interface MockCacheDal {
  mock: any /* mock.Mock */;
}

//////////
// source: config.go

/**
 * Config represents the configuration of the application.
 */
export interface Config {
  logLevel?: LogLevel;
  clusterKey?: string;
  server?: ServerConfig;
  admin?: AdminConfig;
  database?: DatabaseConfig;
  projects?: (ProjectConfig | undefined)[];
  rateLimiters?: RateLimiterConfig;
  metrics?: MetricsConfig;
  proxyPools?: (ProxyPoolConfig | undefined)[];
}
export interface ServerConfig {
  listenV4?: boolean;
  httpHostV4?: string;
  listenV6?: boolean;
  httpHostV6?: string;
  httpPort?: number /* int */;
  maxTimeout?: Duration;
  readTimeout?: Duration;
  writeTimeout?: Duration;
  enableGzip?: boolean;
  tls?: TLSConfig;
  aliasing?: AliasingConfig;
}
export interface AdminConfig {
  auth?: AuthConfig;
  cors?: CORSConfig;
}
export interface AliasingConfig {
  rules: (AliasingRuleConfig | undefined)[];
}
export interface AliasingRuleConfig {
  matchDomain: string;
  serveProject: string;
  serveArchitecture: string;
  serveChain: string;
}
export interface DatabaseConfig {
  evmJsonRpcCache?: CacheConfig;
  sharedState?: SharedStateConfig;
}
export interface SharedStateConfig {
  clusterKey?: string;
  connector?: ConnectorConfig;
  fallbackTimeout?: Duration;
  lockTtl?: Duration;
}
export interface CacheConfig {
  connectors?: TsConnectorConfig[];
  policies?: (CachePolicyConfig | undefined)[];
  methods?: { [key: string]: CacheMethodConfig | undefined};
}
export interface CacheMethodConfig {
  reqRefs: any[][];
  respRefs: any[][];
  finalized: boolean;
  realtime: boolean;
}
export interface CachePolicyConfig {
  connector: string;
  network?: string;
  method?: string;
  params?: any[];
  finality?: DataFinalityState;
  empty?: CacheEmptyBehavior;
  minItemSize?: ByteSize;
  maxItemSize?: ByteSize;
  ttl?: Duration;
}
export type ConnectorDriverType = string;
export const DriverMemory: ConnectorDriverType = "memory";
export const DriverRedis: ConnectorDriverType = "redis";
export const DriverPostgreSQL: ConnectorDriverType = "postgresql";
export const DriverDynamoDB: ConnectorDriverType = "dynamodb";
export interface ConnectorConfig {
  id?: string;
  driver: TsConnectorDriverType;
  memory?: MemoryConnectorConfig;
  redis?: RedisConnectorConfig;
  dynamodb?: DynamoDBConnectorConfig;
  postgresql?: PostgreSQLConnectorConfig;
}
export interface MemoryConnectorConfig {
  maxItems: number /* int */;
}
export interface MockConnectorConfig {
  memoryconnectorconfig: MemoryConnectorConfig;
  getdelay: number /* time in nanoseconds (time.Duration) */;
  setdelay: number /* time in nanoseconds (time.Duration) */;
  geterrorrate: number /* float64 */;
  seterrorrate: number /* float64 */;
}
export interface TLSConfig {
  enabled: boolean;
  certFile: string;
  keyFile: string;
  caFile?: string;
  insecureSkipVerify?: boolean;
}
export interface RedisConnectorConfig {
  addr: string;
  db: number /* int */;
  tls?: TLSConfig;
  connPoolSize: number /* int */;
  initTimeout?: Duration;
  getTimeout?: Duration;
  setTimeout?: Duration;
}
export interface DynamoDBConnectorConfig {
  table?: string;
  region?: string;
  endpoint?: string;
  auth?: AwsAuthConfig;
  partitionKeyName?: string;
  rangeKeyName?: string;
  reverseIndexName?: string;
  ttlAttributeName?: string;
  initTimeout?: Duration;
  getTimeout?: Duration;
  setTimeout?: Duration;
  statePollInterval?: Duration;
}
export interface PostgreSQLConnectorConfig {
  connectionUri: string;
  table: string;
  minConns?: number /* int32 */;
  maxConns?: number /* int32 */;
  initTimeout?: Duration;
  getTimeout?: Duration;
  setTimeout?: Duration;
}
export interface AwsAuthConfig {
  mode: 'file' | 'env' | 'secret'; // "file", "env", "secret"
  credentialsFile: string;
  profile: string;
  accessKeyID: string;
  secretAccessKey: string;
}
export interface ProjectConfig {
  id: string;
  auth?: AuthConfig;
  cors?: CORSConfig;
  providers?: (ProviderConfig | undefined)[];
  upstreamDefaults?: UpstreamConfig;
  upstreams?: (UpstreamConfig | undefined)[];
  networkDefaults?: NetworkDefaults;
  networks?: (NetworkConfig | undefined)[];
  rateLimitBudget?: string;
  healthCheck?: HealthCheckConfig;
}
export interface NetworkDefaults {
  rateLimitBudget?: string;
  failsafe?: FailsafeConfig;
  selectionPolicy?: SelectionPolicyConfig;
  directiveDefaults?: DirectiveDefaultsConfig;
  evm?: EvmNetworkConfig;
}
export interface CORSConfig {
  allowedOrigins: string[];
  allowedMethods: string[];
  allowedHeaders: string[];
  exposedHeaders: string[];
  allowCredentials?: boolean;
  maxAge: number /* int */;
}
export type VendorSettings = { [key: string]: any};
export interface ProviderConfig {
  id?: string;
  vendor: string;
  settings?: VendorSettings;
  onlyNetworks?: string[];
  upstreamIdTemplate?: string;
  overrides?: { [key: string]: UpstreamConfig | undefined};
}
export interface UpstreamConfig {
  id?: string;
  type?: TsUpstreamType;
  group?: string;
  vendorName?: string;
  endpoint?: string;
  evm?: EvmUpstreamConfig;
  jsonRpc?: JsonRpcUpstreamConfig;
  ignoreMethods?: string[];
  allowMethods?: string[];
  autoIgnoreUnsupportedMethods?: boolean;
  failsafe?: FailsafeConfig;
  rateLimitBudget?: string;
  rateLimitAutoTune?: RateLimitAutoTuneConfig;
  routing?: RoutingConfig;
}
export interface RoutingConfig {
  scoreMultipliers: (ScoreMultiplierConfig | undefined)[];
}
export interface ScoreMultiplierConfig {
  network: string;
  method: string;
  overall: number /* float64 */;
  errorRate: number /* float64 */;
  p90latency: number /* float64 */;
  totalRequests: number /* float64 */;
  throttledRate: number /* float64 */;
  blockHeadLag: number /* float64 */;
  finalizationLag: number /* float64 */;
}
export type Alias = UpstreamConfig;
export interface RateLimitAutoTuneConfig {
  enabled?: boolean;
  adjustmentPeriod: Duration;
  errorRateThreshold: number /* float64 */;
  increaseFactor: number /* float64 */;
  decreaseFactor: number /* float64 */;
  minBudget: number /* int */;
  maxBudget: number /* int */;
}
export interface JsonRpcUpstreamConfig {
  supportsBatch?: boolean;
  batchMaxSize?: number /* int */;
  batchMaxWait?: Duration;
  enableGzip?: boolean;
  headers?: { [key: string]: string};
  proxyPool?: string;
}
export interface EvmUpstreamConfig {
  chainId: number /* int64 */;
  nodeType?: EvmNodeType;
  statePollerInterval?: Duration;
  statePollerDebounce?: Duration;
  maxAvailableRecentBlocks?: number /* int64 */;
  getLogsMaxBlockRange?: number /* int64 */;
}
export interface FailsafeConfig {
  retry?: RetryPolicyConfig;
  circuitBreaker?: CircuitBreakerPolicyConfig;
  timeout?: TimeoutPolicyConfig;
  hedge?: HedgePolicyConfig;
  consensus?: ConsensusPolicyConfig;
}
export interface RetryPolicyConfig {
  maxAttempts: number /* int */;
  delay?: Duration;
  backoffMaxDelay?: Duration;
  backoffFactor?: number /* float32 */;
  jitter?: Duration;
}
export interface CircuitBreakerPolicyConfig {
  failureThresholdCount: number /* uint */;
  failureThresholdCapacity: number /* uint */;
  halfOpenAfter?: Duration;
  successThresholdCount: number /* uint */;
  successThresholdCapacity: number /* uint */;
}
export interface TimeoutPolicyConfig {
  duration?: Duration;
}
export interface HedgePolicyConfig {
  delay?: Duration;
  maxCount: number /* int */;
  quantile?: number /* float64 */;
  minDelay?: Duration;
  maxDelay?: Duration;
}
export type ConsensusFailureBehavior = string;
export const ConsensusFailureBehaviorReturnError: ConsensusFailureBehavior = "returnError";
export const ConsensusFailureBehaviorAcceptAnyValidResult: ConsensusFailureBehavior = "acceptAnyValidResult";
export const ConsensusFailureBehaviorPreferBlockHeadLeader: ConsensusFailureBehavior = "preferBlockHeadLeader";
export const ConsensusFailureBehaviorOnlyBlockHeadLeader: ConsensusFailureBehavior = "onlyBlockHeadLeader";
export type ConsensusLowParticipantsBehavior = string;
export const ConsensusLowParticipantsBehaviorReturnError: ConsensusLowParticipantsBehavior = "returnError";
export const ConsensusLowParticipantsBehaviorAcceptAnyValidResult: ConsensusLowParticipantsBehavior = "acceptAnyValidResult";
export const ConsensusLowParticipantsBehaviorPreferBlockHeadLeader: ConsensusLowParticipantsBehavior = "preferBlockHeadLeader";
export const ConsensusLowParticipantsBehaviorOnlyBlockHeadLeader: ConsensusLowParticipantsBehavior = "onlyBlockHeadLeader";
export type ConsensusDisputeBehavior = string;
export const ConsensusDisputeBehaviorReturnError: ConsensusDisputeBehavior = "returnError";
export const ConsensusDisputeBehaviorAcceptAnyValidResult: ConsensusDisputeBehavior = "acceptAnyValidResult";
export const ConsensusDisputeBehaviorPreferBlockHeadLeader: ConsensusDisputeBehavior = "preferBlockHeadLeader";
export const ConsensusDisputeBehaviorOnlyBlockHeadLeader: ConsensusDisputeBehavior = "onlyBlockHeadLeader";
export interface ConsensusPolicyConfig {
  requiredParticipants: number /* int */;
  agreementThreshold?: number /* int */;
  failureBehavior?: ConsensusFailureBehavior;
  disputeBehavior?: ConsensusDisputeBehavior;
  lowParticipantsBehavior?: ConsensusLowParticipantsBehavior;
  punishMisbehavior?: PunishMisbehaviorConfig;
}
export interface PunishMisbehaviorConfig {
  disputeThreshold: number /* int */;
  sitOutPenalty?: string;
}
export interface RateLimiterConfig {
  budgets: RateLimitBudgetConfig[];
}
export interface RateLimitBudgetConfig {
  id: string;
  rules: RateLimitRuleConfig[];
}
export interface RateLimitRuleConfig {
  method: string;
  maxCount: number /* uint */;
  period: Duration;
  waitTime: Duration;
}
export interface ProxyPoolConfig {
  id: string;
  urls: string[];
}
export interface HealthCheckConfig {
  scoreMetricsWindowSize: Duration;
}
export interface NetworkConfig {
  architecture: TsNetworkArchitecture;
  rateLimitBudget?: string;
  failsafe?: FailsafeConfig;
  evm?: EvmNetworkConfig;
  selectionPolicy?: SelectionPolicyConfig;
  directiveDefaults?: DirectiveDefaultsConfig;
}
export interface DirectiveDefaultsConfig {
  retryEmpty?: boolean;
  retryPending?: boolean;
  skipCacheRead?: boolean;
  useUpstream?: string;
}
export interface EvmNetworkConfig {
  chainId: number /* int64 */;
  fallbackFinalityDepth?: number /* int64 */;
  fallbackStatePollerDebounce?: Duration;
  integrity?: EvmIntegrityConfig;
}
export interface EvmIntegrityConfig {
  enforceHighestBlock?: boolean;
  enforceGetLogsBlockRange?: boolean;
}
export interface SelectionPolicyConfig {
  evalInterval?: Duration;
  evalFunction?: SelectionPolicyEvalFunction | undefined;
  evalPerMethod?: boolean;
  resampleExcluded?: boolean;
  resampleInterval?: Duration;
  resampleCount?: number /* int */;
}
export type AuthType = string;
export const AuthTypeSecret: AuthType = "secret";
export const AuthTypeJwt: AuthType = "jwt";
export const AuthTypeSiwe: AuthType = "siwe";
export const AuthTypeNetwork: AuthType = "network";
export interface AuthConfig {
  strategies: TsAuthStrategyConfig[];
}
export interface AuthStrategyConfig {
  ignoreMethods?: string[];
  allowMethods?: string[];
  rateLimitBudget?: string;
  type: TsAuthType;
  network?: NetworkStrategyConfig;
  secret?: SecretStrategyConfig;
  jwt?: JwtStrategyConfig;
  siwe?: SiweStrategyConfig;
}
export interface SecretStrategyConfig {
  value: string;
}
export interface JwtStrategyConfig {
  allowedIssuers: string[];
  allowedAudiences: string[];
  allowedAlgorithms: string[];
  requiredClaims: string[];
  verificationKeys: { [key: string]: string};
}
export interface SiweStrategyConfig {
  allowedDomains: string[];
}
export interface NetworkStrategyConfig {
  allowedIPs: string[];
  allowedCIDRs: string[];
  allowLocalhost: boolean;
  trustedProxies: string[];
}
export interface MetricsConfig {
  enabled?: boolean;
  listenV4?: boolean;
  hostV4?: string;
  listenV6?: boolean;
  hostV6?: string;
  port?: number /* int */;
}

//////////
// source: data.go

export type DataFinalityState = number /* int */;
/**
 * Finalized gets 0 intentionally so that when user has not specified finality,
 * it defaults to finalized, which is safest sane default for caching.
 * This attribute will be calculated based on extracted block number (from request and/or response)
 * and comparing to the upstream (one that returned the response) 'finalized' block (fetch via evm state poller).
 */
export const DataFinalityStateFinalized: DataFinalityState = 0;
/**
 * When we CAN determine the block number, and it's after the upstream 'finalized' block, we consider the data unfinalized.
 */
export const DataFinalityStateUnfinalized: DataFinalityState = 1;
/**
 * Certain methods points are meant to be realtime and updated with every new block (e.g. eth_gasPrice).
 * These can be cached with short TTLs to improve performance.
 */
export const DataFinalityStateRealtime: DataFinalityState = 2;
/**
 * When we CANNOT determine the block number (e.g some trace by hash calls), we consider the data unknown.
 * Most often it is safe to cache this data for longer as they're access when block hash is provided directly.
 */
export const DataFinalityStateUnknown: DataFinalityState = 3;
export type CacheEmptyBehavior = number /* int */;
export const CacheEmptyBehaviorIgnore: CacheEmptyBehavior = 0;
export const CacheEmptyBehaviorAllow: CacheEmptyBehavior = 1;
export const CacheEmptyBehaviorOnly: CacheEmptyBehavior = 2;

//////////
// source: network.go

export type NetworkArchitecture = string;
export const ArchitectureEvm: NetworkArchitecture = "evm";
export type Network = any;
export type QuantileTracker = any;
export type TrackedMetrics = any;

//////////
// source: upstream.go

export type Scope = string;
/**
 * Policies must be created with a "network" in mind,
 * assuming there will be many upstreams e.g. Retry might endup using a different upstream
 */
export const ScopeNetwork: Scope = "network";
/**
 * Policies must be created with one only "upstream" in mind
 * e.g. Retry with be towards the same upstream
 */
export const ScopeUpstream: Scope = "upstream";
export type UpstreamType = string;
export type Upstream = any;
