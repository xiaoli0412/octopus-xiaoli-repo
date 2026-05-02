import type { AITaskArtifacts, AIProfile } from '@/api/endpoints/ai-automation';

type JSONObject = Record<string, unknown>;

export type ResultDiffSummary = {
  added: string[];
  removed: string[];
  shared: string[];
};

export type ResultConsumptionView = {
  rawOutput: string;
  resultObject: JSONObject | null;
  domainPayload: JSONObject | null;
  summary: string;
  toolExecution: JSONObject | null;
  toolExecutionSummary: JSONObject | null;
  resultProfileID?: number;
  protectedActions: Array<JSONObject>;
};

function asObject(value: unknown): JSONObject | null {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return null;
  return value as JSONObject;
}

function asObjectArray(value: unknown): Array<JSONObject> {
  if (!Array.isArray(value)) return [];
  return value
    .map((item) => asObject(item))
    .filter((item): item is JSONObject => item !== null);
}

function parseResultJSON(resultJSON?: string): JSONObject | null {
  if (!resultJSON) return null;

  try {
    return asObject(JSON.parse(resultJSON));
  } catch {
    return null;
  }
}

export function consumeResultView(artifacts?: AITaskArtifacts, fallbackSummary?: string): ResultConsumptionView {
  const parsedResultJSON = parseResultJSON(artifacts?.result_json);
  const resultObject = asObject(artifacts?.result_payload)
    ?? asObject(parsedResultJSON?.result_payload)
    ?? parsedResultJSON;
  const toolExecution = asObject(resultObject?.tool_execution);
  const toolExecutionSummary = asObject(resultObject?.tool_execution_summary);
  const writes = asObjectArray(toolExecution?.writes);
  const profileWrite = writes.find((item) => item.key === 'profile_write');
  const protectedActions = asObjectArray(toolExecution?.protected_actions);
  const domainPayload = asObject(resultObject?.domain_payload);
  const summary = typeof resultObject?.summary === 'string' && resultObject.summary.trim()
    ? resultObject.summary.trim()
    : (fallbackSummary?.trim() || '');

  return {
    rawOutput: artifacts?.result_json || '',
    resultObject,
    domainPayload,
    summary,
    toolExecution,
    toolExecutionSummary,
    resultProfileID: typeof profileWrite?.profile_id === 'number' ? profileWrite.profile_id : undefined,
    protectedActions,
  };
}

export function buildResultDiffSummary(currentProfile?: AIProfile, resultDomainPayload?: JSONObject | null): ResultDiffSummary {
  const currentPayload = asObject(currentProfile?.domain_payload);
  const currentKeys = new Set(Object.keys(currentPayload || {}));
  const resultKeys = new Set(Object.keys(resultDomainPayload || {}));

  const added = [...resultKeys].filter((key) => !currentKeys.has(key)).sort();
  const removed = [...currentKeys].filter((key) => !resultKeys.has(key)).sort();
  const shared = [...resultKeys].filter((key) => currentKeys.has(key)).sort();

  return { added, removed, shared };
}

export function compactObjectKeys(value: unknown) {
  const objectValue = asObject(value);
  if (!objectValue) return '-';
  const keys = Object.keys(objectValue);
  return keys.length > 0 ? keys.slice(0, 8).join(', ') : '-';
}

export function stringifyPanelJSON(value: unknown, fallback = '-') {
  if (value === null || value === undefined) return fallback;
  try {
    return JSON.stringify(value, null, 2);
  } catch {
    return fallback;
  }
}

export function summarizeProtectedAction(action: JSONObject) {
  const key = typeof action.key === 'string' ? action.key : 'unknown';
  const executed = action.executed === true;
  const reason = typeof action.reason === 'string' ? action.reason : '';
  return {
    key,
    executed,
    reason,
  };
}
