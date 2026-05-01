#!/usr/bin/env node

import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, '..');

function readText(relativePath) {
  const fullPath = path.join(repoRoot, relativePath);
  return fs.readFileSync(fullPath, 'utf8');
}

function countLines(text) {
  return text.split(/\r?\n/).length;
}

function assertIncludes(relativePath, text, snippets) {
  for (const snippet of snippets) {
    if (!text.includes(snippet)) {
      throw new Error(`${relativePath} is missing expected snippet: ${snippet}`);
    }
  }
}

function assertExcludes(relativePath, text, snippets) {
  for (const snippet of snippets) {
    if (text.includes(snippet)) {
      throw new Error(`${relativePath} still contains copied-wrapper snippet: ${snippet}`);
    }
  }
}

function assertSameMembers(label, actualMembers, expectedMembers) {
  const actual = [...actualMembers].sort();
  const expected = [...expectedMembers].sort();

  const actualOnly = actual.filter((entry) => !expected.includes(entry));
  const expectedOnly = expected.filter((entry) => !actual.includes(entry));

  if (actualOnly.length === 0 && expectedOnly.length === 0) {
    return;
  }

  const details = [];
  if (actualOnly.length > 0) {
    details.push(`unexpected: ${actualOnly.join(', ')}`);
  }
  if (expectedOnly.length > 0) {
    details.push(`missing: ${expectedOnly.join(', ')}`);
  }

  throw new Error(`${label} drifted (${details.join('; ')})`);
}

function listBrowserSmokePowerShellScripts() {
  return fs
    .readdirSync(path.join(repoRoot, 'scripts'), { withFileTypes: true })
    .filter((entry) => entry.isFile() && /^verify-.*browser-smoke.*\.ps1$/i.test(entry.name))
    .map((entry) => path.posix.join('scripts', entry.name))
    .sort();
}

const copiedWrapperMarkers = [
  'function Assert-NodeSmokeSucceeded',
  'function Wait-HttpOk',
  'function Resolve-NodeSmokeScriptPath',
  'function Start-LocalOctopusBackend',
  'function Read-LogTail',
];

const thinForwarderChecks = [
  {
    file: 'scripts/verify-backup-browser-smoke.ps1',
    maxLines: 90,
    include: [
      "verify-channel-create-browser-smoke.ps1",
      "[Alias('BrowserPath')]",
      '[string]$Browser',
      "$env:OCTOPUS_UI_SMOKE_SCRIPT = 'scripts/verify-backup-browser-smoke.mjs'",
      "$env:OCTOPUS_UI_SMOKE_SUCCESS_MARKER = 'backup-browser-smoke passed'",
      "$env:OCTOPUS_UI_SMOKE_LABEL = 'backup'",
      "if ($PSBoundParameters.ContainsKey('Browser')) { $forwardParams.Browser = $Browser }",
    ],
  },
  {
    file: 'scripts/verify-ccswitch-browser-smoke-cli.ps1',
    maxLines: 90,
    include: [
      "verify-channel-create-browser-smoke.ps1",
      "[Alias('BrowserPath')]",
      '[string]$Browser',
      "$env:OCTOPUS_UI_SMOKE_SCRIPT = 'scripts/verify-ccswitch-browser-smoke.mjs'",
      "$env:OCTOPUS_UI_SMOKE_SUCCESS_MARKER = 'ccswitch-browser-smoke passed'",
      "$env:OCTOPUS_UI_SMOKE_LABEL = 'ccswitch'",
      "if ($PSBoundParameters.ContainsKey('Browser')) { $forwardParams.Browser = $Browser }",
    ],
  },
  {
    file: 'scripts/verify-group-create-browser-smoke.ps1',
    maxLines: 150,
    include: [
      "verify-group-create-browser-smoke-cdp.ps1",
      "verify-channel-create-browser-smoke.ps1",
      "[Alias('BrowserPath')]",
      '[string]$Browser',
      "$env:OCTOPUS_UI_SMOKE_SCRIPT = 'scripts/verify-group-create-browser-smoke.mjs'",
      "$env:OCTOPUS_UI_SMOKE_SUCCESS_MARKER = 'group-create-browser-smoke passed'",
      "$env:OCTOPUS_UI_SMOKE_LABEL = 'group create'",
      "if ($PSBoundParameters.ContainsKey('Browser')) { $forwardParams.BrowserPath = $Browser }",
      "if ($PSBoundParameters.ContainsKey('Browser')) { $forwardParams.Browser = $Browser }",
    ],
  },
  {
    file: 'scripts/verify-setting-help-browser-smoke.ps1',
    maxLines: 150,
    include: [
      "verify-channel-create-browser-smoke-cdp.ps1",
      "verify-channel-create-browser-smoke.ps1",
      "[Alias('BrowserPath')]",
      '[string]$Browser',
      "NodeSmokeScript = 'scripts/verify-setting-help-browser-smoke-cdp.mjs'",
      "NodeSmokeSuccessMarker = 'setting-help-browser-smoke-cdp passed'",
      "SmokeLabel = 'settings help'",
      "[string]$CdpPageBootstrapStrategy = 'auto'",
      "[string]$CdpBootstrapCommandOrder = 'page-lifecycle-runtime'",
      "if ($PSBoundParameters.ContainsKey('Browser')) { $forwardParams.BrowserPath = $Browser }",
      "$env:OCTOPUS_UI_SMOKE_SCRIPT = 'scripts/verify-setting-help-browser-smoke.mjs'",
      "$env:OCTOPUS_UI_SMOKE_SUCCESS_MARKER = 'setting-help-browser-smoke passed'",
      "$env:OCTOPUS_UI_SMOKE_LABEL = 'settings help'",
      "if ($PSBoundParameters.ContainsKey('Browser')) { $forwardParams.Browser = $Browser }",
      'if ($RequireExternalCdpPreflight) { $forwardParams.RequireExternalCdpPreflight = $true }',
    ],
  },
  {
    file: 'scripts/verify-group-create-browser-smoke-cdp.ps1',
    maxLines: 110,
    include: [
      "verify-channel-create-browser-smoke-cdp.ps1",
      "[Alias('BrowserPath')]",
      '[string]$Browser',
      "NodeSmokeScript = 'scripts/verify-group-create-browser-smoke-cdp.mjs'",
      "NodeSmokeSuccessMarker = 'group-create-browser-smoke-cdp passed'",
      "SmokeLabel = 'group create'",
      "[string]$CdpPageBootstrapStrategy = 'attached-session'",
      "[string]$CdpBootstrapCommandOrder = 'page-lifecycle-runtime'",
      "if ($PSBoundParameters.ContainsKey('Browser')) { $forwardParams.BrowserPath = $Browser }",
      'if ($RequireExternalCdpPreflight) { $forwardParams.RequireExternalCdpPreflight = $true }',
    ],
  },
  {
    file: 'scripts/verify-ccswitch-browser-smoke.ps1',
    maxLines: 110,
    include: [
      "verify-channel-create-browser-smoke-cdp.ps1",
      "[Alias('BrowserPath')]",
      '[string]$Browser',
      "NodeSmokeSuccessMarker = 'ccswitch-browser-smoke-cdp passed'",
      "SmokeLabel = 'ccswitch'",
      "$env:OCTOPUS_UI_SMOKE_SCENARIO = 'ccswitch'",
      "[string]$CdpPageBootstrapStrategy = 'attached-session'",
      "[string]$CdpBootstrapCommandOrder = 'runtime-page-lifecycle'",
      "if ($PSBoundParameters.ContainsKey('Browser')) { $forwardParams.BrowserPath = $Browser }",
      'if ($RequireExternalCdpPreflight) { $forwardParams.RequireExternalCdpPreflight = $true }',
    ],
  },
  {
    file: 'scripts/verify-channel-page-browser-smoke.ps1',
    maxLines: 110,
    include: [
      "verify-channel-create-browser-smoke-cdp.ps1",
      "[Alias('BrowserPath')]",
      '[string]$Browser',
      "NodeSmokeSuccessMarker = 'channel-page-browser-smoke-cdp passed'",
      "SmokeLabel = 'channel page'",
      "$env:OCTOPUS_UI_SMOKE_SCENARIO = 'channel-page'",
      "[string]$CdpPageBootstrapStrategy = 'attached-session'",
      "[string]$CdpBootstrapCommandOrder = 'page-lifecycle-runtime'",
      "if ($PSBoundParameters.ContainsKey('Browser')) { $forwardParams.BrowserPath = $Browser }",
      'if ($RequireExternalCdpPreflight) { $forwardParams.RequireExternalCdpPreflight = $true }',
    ],
  },
  {
    file: 'scripts/verify-home-layout-browser-smoke.ps1',
    maxLines: 110,
    include: [
      "verify-channel-create-browser-smoke-cdp.ps1",
      "[Alias('BrowserPath')]",
      '[string]$Browser',
      "NodeSmokeSuccessMarker = 'home-layout-browser-smoke-cdp passed'",
      "SmokeLabel = 'home layout'",
      "$env:OCTOPUS_UI_SMOKE_SCENARIO = 'home-layout'",
      "[string]$CdpPageBootstrapStrategy = 'attached-session'",
      "[string]$CdpBootstrapCommandOrder = 'page-lifecycle-runtime'",
      "if ($PSBoundParameters.ContainsKey('Browser')) { $forwardParams.BrowserPath = $Browser }",
      'if ($RequireExternalCdpPreflight) { $forwardParams.RequireExternalCdpPreflight = $true }',
    ],
  },
  {
    file: 'scripts/verify-model-layout-browser-smoke.ps1',
    maxLines: 110,
    include: [
      "verify-channel-create-browser-smoke-cdp.ps1",
      "[Alias('BrowserPath')]",
      '[string]$Browser',
      "NodeSmokeSuccessMarker = 'model-layout-browser-smoke-cdp passed'",
      "SmokeLabel = 'model layout'",
      "$env:OCTOPUS_UI_SMOKE_SCENARIO = 'model-layout'",
      "[string]$CdpPageBootstrapStrategy = 'attached-session'",
      "[string]$CdpBootstrapCommandOrder = 'page-lifecycle-runtime'",
      "if ($PSBoundParameters.ContainsKey('Browser')) { $forwardParams.BrowserPath = $Browser }",
      'if ($RequireExternalCdpPreflight) { $forwardParams.RequireExternalCdpPreflight = $true }',
    ],
  },
];

const rootWrapperChecks = [
  {
    file: 'scripts/verify-channel-create-browser-smoke.ps1',
    include: [
      "[ValidateSet('cdp', 'cli')]",
      "[Alias('BrowserPath')]",
      '[string]$Browser',
      '[int]$CdpPort = 9222',
      '[string]$CdpUrl',
      '[int]$CdpCommandTimeoutMs = 15000',
      "[string]$EdgeLaunchPreset = 'default'",
      "[string]$EdgeProfileStrategy = 'temp-random'",
      "[string]$CdpPageBootstrapStrategy = 'attached-session'",
      "[string]$CdpBootstrapCommandOrder = 'page-lifecycle-runtime'",
      '[switch]$BootstrapExternalCdpSession',
      '[switch]$RequireExternalCdpPreflight',
      '[switch]$SelfStartServices',
      "if ($PSBoundParameters.ContainsKey('Browser')) {",
      '$forwardParams.BrowserPath = $Browser',
      "throw (\"Unable to resolve Node.js executable for {0} smoke.\" -f $smokeLabel)",
      "throw (\"Node {0} is blocked by a host-level child-process 'spawn EPERM' failure while launching Playwright CLI.",
      "$tempRoot = Join-Path $env:TEMP ('octopus-' + $smokeLabelSlug + '-smoke-' + [guid]::NewGuid().ToString('N'))",
      'CdpPort = $CdpPort',
      'CdpCommandTimeoutMs = $CdpCommandTimeoutMs',
      'EdgeLaunchPreset = $EdgeLaunchPreset',
      'EdgeProfileStrategy = $EdgeProfileStrategy',
      'CdpPageBootstrapStrategy = $CdpPageBootstrapStrategy',
      'CdpBootstrapCommandOrder = $CdpBootstrapCommandOrder',
      "if ($PSBoundParameters.ContainsKey('CdpUrl')) {",
      '$forwardParams.CdpUrl = $CdpUrl',
      'if ($BootstrapExternalCdpSession) {',
      '$forwardParams.BootstrapExternalCdpSession = $true',
      'if ($RequireExternalCdpPreflight) {',
      '$forwardParams.RequireExternalCdpPreflight = $true',
      'if ($SelfStartServices) {',
      '$forwardParams.SelfStartServices = $true',
      'function Test-IsCodexBundledNode',
      'function Resolve-NodeCommandPath',
      "$nodeSmokeScript = if ([string]::IsNullOrWhiteSpace($env:OCTOPUS_UI_SMOKE_SCRIPT)) { 'scripts/verify-channel-create-browser-smoke.mjs' } else { $env:OCTOPUS_UI_SMOKE_SCRIPT }",
      "$nodeSmokeSuccessMarker = if ([string]::IsNullOrWhiteSpace($env:OCTOPUS_UI_SMOKE_SUCCESS_MARKER)) { 'channel-create-browser-smoke passed' } else { $env:OCTOPUS_UI_SMOKE_SUCCESS_MARKER }",
      "$smokeLabel = if ([string]::IsNullOrWhiteSpace($env:OCTOPUS_UI_SMOKE_LABEL)) { 'channel create' } else { $env:OCTOPUS_UI_SMOKE_LABEL }",
    ],
  },
  {
    file: 'scripts/verify-channel-create-browser-smoke-cdp.ps1',
    include: [
      "[Alias('BrowserPath')]",
      '[string]$Browser',
      "[string]$NodeSmokeScript = 'scripts/verify-channel-create-browser-smoke-cdp.mjs'",
      "[string]$NodeSmokeSuccessMarker = 'channel-create-browser-smoke-cdp passed'",
      "[string]$SmokeLabel = 'channel create'",
      '[switch]$BootstrapExternalCdpSession',
      '[switch]$RequireExternalCdpPreflight',
      'function Test-IsCodexBundledNode',
      'function Resolve-NodeCommandPath',
      '$bootstrapArgs.BrowserPath = $Browser',
      '$resolvedBrowserPath = Resolve-CommandPath -Candidates @(',
      '    $Browser,',
    ],
  },
  {
    file: 'scripts/verify-ai-automation-learning-browser-smoke.ps1',
    include: [
      "Join-Path $scriptDir 'verify-channel-create-browser-smoke-cdp.ps1'",
      "[Alias('BrowserPath')]",
      '[string]$Browser',
      '[int]$StableDiagnosticFreshnessThresholdHours = 24',
      "[switch]$UseHostFriendlyExternalDefaults",
      '[switch]$SelfStartLocalServices',
      "NodeSmokeSuccessMarker = 'ai-automation-learning-browser-smoke-cdp passed'",
      "$env:OCTOPUS_UI_SMOKE_SCENARIO = 'ai-learning'",
      "if ($PSBoundParameters.ContainsKey('Browser')) { $forwardParams.BrowserPath = $Browser }",
      'if ($SelfStartServices -or $SelfStartLocalServices) { $forwardParams.SelfStartServices = $true }',
      '$forwardParams.BootstrapExternalCdpSession = $true',
      'Get-ExternalPreflightRecommendedRefreshCommand',
      'Get-ExternalPreflightTimeoutComparisonCommand',
      'Write-StableExternalPreflightDiagnosticPreview',
      "Write-Host ('Stable external preflight freshness threshold (hours): {0}' -f $FreshnessThresholdHours)",
      '-FreshnessThresholdHours $StableDiagnosticFreshnessThresholdHours',
    ],
  },
];

const runtimeSupportChecks = [
  {
    file: 'scripts/runtime-win.ps1',
    include: [
      "[ValidateSet('status', 'stop', 'healthcheck', 'check-only')]",
      "Write-Info ('Port scan mode: {0}' -f $PortScanDetails)",
      'Port owner hints are informational only in low-privilege mode; stop still targets workspace-attributed octopus_repo processes only.',
      'Loopback localhost readiness: blocked on {0} by Windows service-provider initialization; local self-start/external smoke will fail until the host networking stack is repaired.',
      'Loopback localhost readiness: usable on {0} for local smoke preflight.',
      "Write-Info ('Command: powershell -ExecutionPolicy Bypass -File .\\scripts\\runtime-win.ps1 -Action status')",
      "Write-Info ('Command: powershell -ExecutionPolicy Bypass -File .\\scripts\\runtime-win.ps1 -Action stop')",
      "Write-Info ('Command: go run main.go healthcheck')",
      "Write-Success 'Runtime management entrypoints are available.'",
      "Write-Host 'Automation entrypoints:'",
      "Write-Host '  root: D:\\GPT-codex\\octopus_repo'",
      "Write-Host '  frontend: D:\\GPT-codex\\octopus_repo\\web'",
      "Write-Host '  scripts: D:\\GPT-codex\\octopus_repo\\scripts'",
      "Write-Host '  docs: D:\\GPT-codex\\octopus_repo\\docs'",
      "ScanMode = 'netstat'",
      "ScanDetails = 'netstat -ano -p tcp'",
    ],
  },
];

const activeStatusDocChecks = [
  {
    file: 'docs/archive/status/CURRENT_STATUS_AND_PLAN.zh-CN.md',
    include: [
      '/D:/GPT-codex/octopus_repo/docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md',
    ],
    exclude: [
      '/D:/GPT-codex/octopus_repo/docs/LLM-Gateway-Refactor-Plan.zh-CN.md',
    ],
  },
  {
    file: 'docs/archive/status/ENV_READY_AND_NEXT_PLAN.zh-CN.md',
    include: [
      '/D:/GPT-codex/octopus_repo/docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md',
    ],
    exclude: [
      '/D:/GPT-codex/octopus_repo/docs/LLM-Gateway-Refactor-Plan.zh-CN.md',
    ],
  },
  {
    file: 'docs/archive/status/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md',
    include: [
      '/D:/GPT-codex/octopus_repo/docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md',
    ],
    exclude: [
      '/D:/GPT-codex/octopus_repo/docs/LLM-Gateway-Refactor-Plan.zh-CN.md',
    ],
  },
  {
    file: 'docs/archive/status/DETAILED_EXECUTION_WORKFLOW.zh-CN.md',
    include: [
      '/D:/GPT-codex/octopus_repo/docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md',
      '/D:/GPT-codex/octopus_repo/docs/archive/status/DETAILED_EXECUTION_WORKFLOW.zh-CN.md',
      '/D:/GPT-codex/octopus_repo/docs/archive/status/ENV_READY_AND_NEXT_PLAN.zh-CN.md',
      '/D:/GPT-codex/octopus_repo/docs/archive/status/CURRENT_STATUS_AND_PLAN.zh-CN.md',
    ],
    exclude: [
      '/D:/GPT-codex/octopus_repo/docs/LLM-Gateway-Refactor-Plan.zh-CN.md',
      '/D:/GPT-codex/octopus_repo/docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md',
      '/D:/GPT-codex/octopus_repo/docs/ENV_READY_AND_NEXT_PLAN.zh-CN.md',
      '/D:/GPT-codex/octopus_repo/docs/CURRENT_STATUS_AND_PLAN.zh-CN.md',
    ],
  },
];

const activeArchiveInputDocChecks = [
  {
    file: 'docs/archive/requirements/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md',
    include: [
      '/D:/GPT-codex/octopus_repo/docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md',
      '/D:/GPT-codex/octopus_repo/docs/archive/status/CURRENT_STATUS_AND_PLAN.zh-CN.md',
      '/D:/GPT-codex/octopus_repo/docs/archive/status/DETAILED_EXECUTION_WORKFLOW.zh-CN.md',
      '/D:/GPT-codex/octopus_repo/docs/archive/status/ENV_READY_AND_NEXT_PLAN.zh-CN.md',
      '/D:/GPT-codex/octopus_repo/docs/archive/status/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md',
      '/D:/GPT-codex/octopus_repo/docs/archive/requirements/DYNAMIC_ROUTING_REQUIREMENTS.zh-CN.md',
      '/D:/GPT-codex/octopus_repo/docs/archive/planning/DYNAMIC_ROUTING_IMPLEMENTATION_PLAN.zh-CN.md',
    ],
    exclude: [
      '/D:/GPT-codex/octopus_repo/docs/LLM-Gateway-Refactor-Plan.zh-CN.md',
      '/D:/GPT-codex/octopus_repo/docs/CURRENT_STATUS_AND_PLAN.zh-CN.md',
      '/D:/GPT-codex/octopus_repo/docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md',
      '/D:/GPT-codex/octopus_repo/docs/ENV_READY_AND_NEXT_PLAN.zh-CN.md',
      '/D:/GPT-codex/octopus_repo/docs/FRONTEND_UI_MAINLINE_STATUS.zh-CN.md',
      '/D:/GPT-codex/octopus_repo/docs/DYNAMIC_ROUTING_REQUIREMENTS.zh-CN.md',
      '/D:/GPT-codex/octopus_repo/docs/DYNAMIC_ROUTING_IMPLEMENTATION_PLAN.zh-CN.md',
    ],
  },
  {
    file: 'docs/archive/planning/DYNAMIC_ROUTING_IMPLEMENTATION_PLAN.zh-CN.md',
    include: [
      '/D:/GPT-codex/octopus_repo/docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md',
      '/D:/GPT-codex/octopus_repo/docs/archive/status/DETAILED_EXECUTION_WORKFLOW.zh-CN.md',
      '/D:/GPT-codex/octopus_repo/docs/archive/requirements/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md',
    ],
    exclude: [
      '/D:/GPT-codex/octopus_repo/docs/LLM-Gateway-Refactor-Plan.zh-CN.md',
      '/D:/GPT-codex/octopus_repo/docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md',
      '/D:/GPT-codex/octopus_repo/docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md',
    ],
  },
  {
    file: 'docs/archive/requirements/DYNAMIC_ROUTING_REQUIREMENTS.zh-CN.md',
    include: [
      '/D:/GPT-codex/octopus_repo/docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md',
      '/D:/GPT-codex/octopus_repo/docs/archive/requirements/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md',
    ],
    exclude: [
      '/D:/GPT-codex/octopus_repo/docs/LLM-Gateway-Refactor-Plan.zh-CN.md',
      '/D:/GPT-codex/octopus_repo/docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md',
    ],
  },
  {
    file: 'docs/archive/requirements/AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md',
    include: [
      '/D:/GPT-codex/octopus_repo/docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md',
      '/D:/GPT-codex/octopus_repo/docs/archive/requirements/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md',
      '/D:/GPT-codex/octopus_repo/docs/archive/status/DETAILED_EXECUTION_WORKFLOW.zh-CN.md',
    ],
    exclude: [
      '/D:/GPT-codex/octopus_repo/docs/LLM-Gateway-Refactor-Plan.zh-CN.md',
      '/D:/GPT-codex/octopus_repo/docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md',
      '/D:/GPT-codex/octopus_repo/docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md',
    ],
  },
  {
    file: 'docs/archive/requirements/AI_AUTOMATION_CENTER_REQUIREMENTS.md',
    include: [
      '/D:/GPT-codex/octopus_repo/docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md',
      '/D:/GPT-codex/octopus_repo/docs/archive/requirements/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md',
      '/D:/GPT-codex/octopus_repo/docs/archive/status/DETAILED_EXECUTION_WORKFLOW.zh-CN.md',
    ],
    exclude: [
      '/D:/GPT-codex/octopus_repo/docs/LLM-Gateway-Refactor-Plan.zh-CN.md',
      '/D:/GPT-codex/octopus_repo/docs/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md',
      '/D:/GPT-codex/octopus_repo/docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md',
    ],
  },
  {
    file: 'docs/archive/worklog/worklog/README.zh-CN.md',
    include: [
      '/D:/GPT-codex/octopus_repo/docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md',
      '/D:/GPT-codex/octopus_repo/docs/archive/status/DETAILED_EXECUTION_WORKFLOW.zh-CN.md',
      'docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md',
    ],
    exclude: [
      '/D:/GPT-codex/octopus_repo/docs/LLM-Gateway-Refactor-Plan.zh-CN.md',
      '/D:/GPT-codex/octopus_repo/docs/DETAILED_EXECUTION_WORKFLOW.zh-CN.md',
      'docs/LLM-Gateway-Refactor-Plan.zh-CN.md',
    ],
  },
];

const wrapperFamilies = {
  thin_forwarders: thinForwarderChecks.map((check) => check.file),
  shared_roots: rootWrapperChecks
    .map((check) => check.file)
    .filter((file) => file !== 'scripts/verify-ai-automation-learning-browser-smoke.ps1'),
  specialized_roots: ['scripts/verify-ai-automation-learning-browser-smoke.ps1'],
};

const expectedBrowserSmokeScripts = Object.values(wrapperFamilies).flat();

assertSameMembers(
  'browser smoke PowerShell wrapper inventory',
  listBrowserSmokePowerShellScripts(),
  expectedBrowserSmokeScripts,
);

for (const check of thinForwarderChecks) {
  const text = readText(check.file);
  assertIncludes(check.file, text, check.include);
  assertExcludes(check.file, text, check.exclude ?? copiedWrapperMarkers);

  const lineCount = countLines(text);
  if (lineCount > check.maxLines) {
    throw new Error(`${check.file} grew to ${lineCount} lines; expected a thin shared-wrapper forwarder (<= ${check.maxLines}).`);
  }
}

for (const check of rootWrapperChecks) {
  const text = readText(check.file);
  assertIncludes(check.file, text, check.include);
}

for (const check of runtimeSupportChecks) {
  const text = readText(check.file);
  assertIncludes(check.file, text, check.include);
}

for (const check of activeStatusDocChecks) {
  const text = readText(check.file);
  assertIncludes(check.file, text, check.include);
  assertExcludes(check.file, text, check.exclude);
}

for (const check of activeArchiveInputDocChecks) {
  const text = readText(check.file);
  assertIncludes(check.file, text, check.include);
  assertExcludes(check.file, text, check.exclude);
}

console.log('browser smoke wrapper alignment passed');
