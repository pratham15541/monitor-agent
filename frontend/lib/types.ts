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
