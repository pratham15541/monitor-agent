export type DeviceStatus = "ONLINE" | "OFFLINE";

export type Device = {
  id: string;
  hostname: string;
  ipAddress: string;
  os: string;
  status: DeviceStatus;
  lastSeenAt: string | null;
};

export type Metric = {
  id: number;
  cpuUsage: number;
  memoryUsage: number;
  diskUsage: number;
  networkIn: number;
  networkOut: number;
  createdAt: string | null;
  device?: Device;
};

export type CompanyProfile = {
  id: string;
  name: string;
  email: string;
  apiToken: string;
};

export type ProcessInfo = {
  pid: number;
  name?: string;
  exe?: string;
  cmdline?: string;
  username?: string;
  status?: string[];
  ppid?: number;
  createTime?: number;
  isRunning?: boolean;
  threads?: number;
  cpuPercent?: number;
  memoryRssBytes?: number;
  memoryVmsBytes?: number;
  memoryPercent?: number;
  ioReadBytes?: number;
  ioWriteBytes?: number;
};

export type ConnectionInfo = {
  pid: number;
  family: number;
  type: number;
  status: string;
  local: {
    ip: string;
    port: number;
  };
  remote: {
    ip: string;
    port: number;
  };
};

export type MemoryDetails = {
  total: number;
  available: number;
  used: number;
  usedPercent: number;
  free: number;
  cached: number;
  buffers: number;
  active: number;
  inactive: number;
  shared: number;
  slab: number;
  pageTables: number;
  swapCached: number;
  sreclaimable: number;
  sunreclaim: number;
  swapTotal?: number;
  swapUsed?: number;
  swapFree?: number;
  swapUsedPercent?: number;
};

export type ServicesSnapshot = {
  source: string;
  output: string;
};

export type LogsSnapshot = {
  agent: string;
  system: string;
};

export type DetailedMetricsPayload = {
  collectedAt: string;
  processes: ProcessInfo[];
  connections: ConnectionInfo[];
  memory: MemoryDetails;
  services: ServicesSnapshot;
  logs: LogsSnapshot;
  os: string;
};

export type MetricDetail = {
  id: number;
  detailsJson: string;
  createdAt: string;
};
