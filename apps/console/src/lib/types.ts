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

export interface Analysis {
  verdict: "ship" | "hold" | "iterate";
  summary: string;
  risks?: string[];
  model?: string;
  stats?: {
    control_rate: number;
    treatment_rate: number;
    relative_lift: number;
    p_value: number;
    significant: boolean;
  };
}

export interface FlagInput {
  key?: string;
  name: string;
  description?: string;
  enabled: boolean;
  rollout: number;
  tags?: string[];
}
