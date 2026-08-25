export interface Flag {
  key: string;
  name: string;
  description?: string;
  enabled: boolean;
  rollout: number;
  tags?: string[];
  created_at: string;
  updated_at: string;
}

export interface AuditEntry {
  flag_key: string;
  action: "created" | "updated" | "deleted";
  actor: string;
  timestamp: string;
}

export interface FlagInput {
  key?: string;
  name: string;
  description?: string;
  enabled: boolean;
  rollout: number;
  tags?: string[];
}
